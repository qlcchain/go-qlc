/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package trie

import (
	"bytes"
	"fmt"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

const (
	idPrefixTrie = 101
)

type Trie struct {
	db    storage.Store
	cache *NodePool
	log   *zap.SugaredLogger

	Root *TrieNode

	unSavedRefValueMap map[types.Hash][]byte
}

func NewTrie(db storage.Store, rootHash *types.Hash, pool *NodePool) *Trie {
	trie := &Trie{
		db:    db,
		cache: pool,
		log:   log.NewLogger("trie"),

		unSavedRefValueMap: make(map[types.Hash][]byte),
	}

	trie.loadFromDb(rootHash)
	return trie
}

func (trie *Trie) getNodeFromDb(key *types.Hash) *TrieNode {
	if trie.db == nil {
		return nil
	}

	trieNode := new(TrieNode)
	k := encodeKey(key[:])
	val, err := trie.db.Get(k)
	if len(val) == 0 || err != nil {
		return nil
	}

	err = trieNode.Deserialize(val)
	if err != nil {
		return nil
	}
	return trieNode

	//if err := txn.Get(k, func(val []byte, b byte) error {
	//	if len(val) > 0 {
	//		return trieNode.Deserialize(val)
	//	}
	//	return errors.New("invalid trie node data")
	//}); err == nil {
	//	return trieNode
	//} else {
	//	return nil
	//}
}

func (trie *Trie) saveNodeToDb(b storage.Batch, node *TrieNode) error {
	if data, err := node.Serialize(); err != nil {
		return fmt.Errorf("serialize trie node failed, error is %s", err)
	} else {
		h := node.Hash()
		k := encodeKey(h[:])
		err := b.Put(k, data)
		//trie.log.Debugf("save %s, s%==>%s", h.String(), hex.EncodeToString(k), node.String())
		if err != nil {
			return err
		}
	}
	return nil
}

func (trie *Trie) deleteUnSavedRefValueMap(node *TrieNode) {
	if node == nil || node.NodeType() != HashNode {
		return
	}

	valueHash, err := types.BytesToHash(node.value)
	if err != nil {
		return
	}
	if _, ok := trie.unSavedRefValueMap[valueHash]; ok {
		delete(trie.unSavedRefValueMap, valueHash)
	}
}

func (trie *Trie) saveRefValueMap(b storage.Batch) {
	for key, value := range trie.unSavedRefValueMap {
		encodeKey := encodeKey(key[:])
		err := b.Put(encodeKey[:], value)
		if err != nil {
			trie.log.Errorf("save %s, error %s", key.String(), err)
		}
	}
}

func (trie *Trie) getRefValue(key []byte) ([]byte, error) {
	hashKey, err := types.BytesToHash(key)
	if err != nil {
		return nil, err
	}

	if value, ok := trie.unSavedRefValueMap[hashKey]; ok {
		return value, nil
	}

	if trie.db == nil {
		return nil, nil
	}

	k := encodeKey(key)
	var result []byte
	i, err := trie.db.Get(k)
	if err != nil {
		return nil, err
	}
	result = make([]byte, len(i))
	copy(result, i)
	return result, nil
}

func (trie *Trie) getNode(key *types.Hash) *TrieNode {
	if trie.cache != nil {
		if node := trie.cache.Get(key); node != nil {
			return node
		}
	}

	node := trie.getNodeFromDb(key)
	if node != nil && trie.cache != nil {
		trie.cache.Set(key, node)
	}
	return node
}

func (trie *Trie) loadFromDb(rootHash *types.Hash) {
	if rootHash == nil {
		return
	}

	trie.Root = trie.traverseLoad(rootHash)
}

func (trie *Trie) traverseLoad(hash *types.Hash) *TrieNode {
	node := trie.getNode(hash)
	if node == nil {
		return nil
	}

	switch node.NodeType() {
	case FullNode:
		node.LeafCallback(func() {
			for key, child := range node.children {
				node.children[key] = trie.traverseLoad(child.Hash())
			}
			if node.child != nil {
				node.child = trie.traverseLoad(node.child.Hash())
			}
		})
	case ShortNode:
		node.child = trie.traverseLoad(node.child.Hash())
	}
	return node
}

func (trie *Trie) Hash() *types.Hash {
	if trie.Root == nil {
		return nil
	}

	return trie.Root.Hash()
}

func (trie *Trie) Clone() *Trie {
	newTrie := &Trie{
		db:    trie.db,
		cache: trie.cache,
		log:   trie.log,

		unSavedRefValueMap: make(map[types.Hash][]byte),
	}
	if trie.Root != nil {
		newTrie.Root = trie.Root.Clone(true)
	}
	return newTrie
}

func (trie *Trie) Save(batch ...storage.Batch) (fn func(), err error) {
	var b storage.Batch
	if len(batch) > 0 {
		b = batch[0]
	} else {
		b = trie.db.Batch(true)
		defer func() {
			if err := trie.db.PutBatch(b); err != nil {
				trie.log.Error(err)
			}
		}()
	}
	err = trie.traverseSave(b, trie.Root)
	if err != nil {
		return nil, err
	}

	trie.saveRefValueMap(b)

	return func() {
		trie.unSavedRefValueMap = make(map[types.Hash][]byte)
	}, nil
}

func (trie *Trie) Remove(batch ...storage.Batch) (err error) {
	var b storage.Batch
	if len(batch) > 0 {
		b = batch[0]
	} else {
		b = trie.db.Batch(true)
		defer func() {
			if err := trie.db.PutBatch(b); err != nil {
				trie.log.Error(err)
			}
		}()
	}
	if err = trie.traverseRemove(b, trie.Root); err != nil {
		return err
	}
	return nil
}

func (trie *Trie) SaveInTxn(batch storage.Batch) (func(), error) {
	err := trie.traverseSave(batch, trie.Root)
	if err != nil {
		return nil, err
	}

	trie.saveRefValueMap(batch)

	return func() {
		trie.unSavedRefValueMap = make(map[types.Hash][]byte)
	}, nil
}

func (trie *Trie) traverseSave(b storage.Batch, node *TrieNode) error {
	if node == nil {
		return nil
	}

	// Cached, no save
	if trie.getNode(node.Hash()) != nil {
		return nil
	}

	err := trie.saveNodeToDb(b, node)
	if err != nil {
		return err
	}

	switch node.NodeType() {
	case FullNode:
		if node.child != nil {
			err := trie.traverseSave(b, node.child)
			if err != nil {
				return err
			}
		}

		for _, child := range node.children {
			err := trie.traverseSave(b, child)
			if err != nil {
				return err
			}
		}

	case ShortNode:
		err := trie.traverseSave(b, node.child)
		if err != nil {
			return err
		}
	}
	return nil
}

func (trie *Trie) SetValue(key []byte, value []byte) {
	var leafNode *TrieNode
	if len(value) > 32 {
		valueHash := types.HashData(value)
		leafNode = NewHashNode(&valueHash)
		defer func() {
			trie.unSavedRefValueMap[valueHash] = value
		}()
	} else {
		leafNode = NewValueNode(value)
	}

	trie.Root = trie.setValue(trie.Root, key, leafNode)
}

func (trie *Trie) NewNodeIterator(fn func(*TrieNode) bool) <-chan *TrieNode {
	ch := make(chan *TrieNode)

	go func(root *TrieNode) {
		if root == nil {
			close(ch)
			return
		}
		var nodes []*TrieNode
		nodes = append(nodes, root)

		for {
			if len(nodes) <= 0 {
				break
			}
			node := nodes[0]
			if fn == nil || fn(node) {
				ch <- node
			}
			nodes[0] = nil
			nodes = nodes[1:]

			switch node.NodeType() {
			case FullNode:
				for _, child := range node.SortedChildren() {
					nodes = append(nodes, child)
				}
				if node.child != nil {
					nodes = append(nodes, node.child)
				}
			case ShortNode:
				if node.child != nil {
					nodes = append(nodes, node.child)
				}
			}
		}
		close(ch)
	}(trie.Root)

	return ch
}

func (trie *Trie) setValue(node *TrieNode, key []byte, leafNode *TrieNode) *TrieNode {
	// Create short_node when node is nil
	if node == nil {
		if len(key) != 0 {
			shortNode := NewShortNode(key, nil)
			shortNode.SetChild(leafNode)
			return shortNode
		} else {
			// Final node
			return leafNode
		}
	}

	// Normal node
	switch node.NodeType() {
	case FullNode:
		newNode := node.Clone(false)

		if len(key) > 0 {
			firstChar := key[0]
			newNode.children[firstChar] = trie.setValue(newNode.children[firstChar], key[1:], leafNode)
		} else {
			trie.deleteUnSavedRefValueMap(newNode.child)
			newNode.child = leafNode
		}
		return newNode
	case ShortNode:
		// sometimes is nil
		var keyChar *byte
		var restKey []byte

		var index = 0
		for ; index < len(key); index++ {
			char := key[index]
			if index >= len(node.key) || node.key[index] != char {
				keyChar = &char
				restKey = key[index+1:]
				break
			}
		}

		var fullNode *TrieNode

		if index >= len(node.key) {
			if len(key) == index && (node.child.NodeType() == ValueNode || node.child.NodeType() == HashNode) {
				trie.deleteUnSavedRefValueMap(node.child)

				newNode := node.Clone(false)
				newNode.SetChild(leafNode)

				return newNode
			} else if node.child.NodeType() == FullNode {
				fullNode = node.child.Clone(false)
			}
		}

		if fullNode == nil {
			fullNode = NewFullNode(nil)
			// sometimes is nil
			var nodeChar *byte
			var nodeRestKey []byte
			if index < len(node.key) {
				nodeChar = &node.key[index]
				nodeRestKey = node.key[index+1:]
			}

			if nodeChar != nil {
				fullNode.children[*nodeChar] = trie.setValue(fullNode.children[*nodeChar], nodeRestKey, node.child)
			} else {
				fullNode.child = node.child
			}
		}

		if keyChar != nil {
			fullNode.children[*keyChar] = trie.setValue(fullNode.children[*keyChar], restKey, leafNode)
		} else {
			fullNode.child = leafNode
		}
		if index > 0 {
			shortNode := NewShortNode(key[0:index], nil)
			shortNode.SetChild(fullNode)
			return shortNode
		} else {
			return fullNode
		}
	default:
		if len(key) > 0 {
			fullNode := NewFullNode(nil)
			fullNode.child = node
			fullNode.children[key[0]] = trie.setValue(nil, key[1:], leafNode)
			return fullNode
		} else {
			trie.deleteUnSavedRefValueMap(node)
			return leafNode
		}
	}
}

func (trie *Trie) LeafNodeValue(leafNode *TrieNode) []byte {
	if leafNode == nil {
		return nil
	}

	switch leafNode.NodeType() {
	case ValueNode:
		return leafNode.value
	case HashNode:
		value, _ := trie.getRefValue(leafNode.value)
		return value
	default:
		return nil
	}
}

func (trie *Trie) GetValue(key []byte) []byte {
	leafNode := trie.getLeafNode(trie.Root, key)

	return trie.LeafNodeValue(leafNode)
}

func (trie *Trie) NewIterator(prefix []byte) *Iterator {
	return NewIterator(trie, prefix)
}

func (trie *Trie) getLeafNode(node *TrieNode, key []byte) *TrieNode {
	if node == nil {
		return nil
	}

	if len(key) == 0 {
		switch node.NodeType() {
		case HashNode:
			fallthrough
		case ValueNode:
			return node
		case FullNode:
			return node.child
		default:
			return nil
		}
	}

	switch node.NodeType() {
	case FullNode:
		return trie.getLeafNode(node.children[key[0]], key[1:])
	case ShortNode:
		if !bytes.HasPrefix(key, node.key) {
			return nil
		}
		return trie.getLeafNode(node.child, key[len(node.key):])
	default:
		return nil
	}
}

func (trie *Trie) removeNodeFromDb(batch storage.Batch, node *TrieNode) error {
	h := node.Hash()
	k := encodeKey(h[:])

	fmt.Println("=========delete trie key ", k)
	if err := batch.Delete(k); err != nil {
		return err
	}

	return nil
}

func (trie *Trie) traverseRemove(b storage.Batch, node *TrieNode) error {
	if node == nil {
		return nil
	}

	trie.deleteUnSavedRefValueMap(node)
	err := trie.removeNodeFromDb(b, node)
	if err != nil {
		return err
	}

	switch node.NodeType() {
	case FullNode:
		if node.child != nil {
			err := trie.traverseRemove(b, node.child)
			if err != nil {
				return err
			}
		}

		for _, child := range node.children {
			err := trie.traverseRemove(b, child)
			if err != nil {
				return err
			}
		}

	case ShortNode:
		err := trie.traverseRemove(b, node.child)
		if err != nil {
			return err
		}
	}
	return nil
}

func encodeKey(key []byte) []byte {
	result := make([]byte, len(key)+1)
	result[0] = idPrefixTrie
	copy(result[1:], key)
	return result
}
