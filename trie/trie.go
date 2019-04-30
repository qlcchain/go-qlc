/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package trie

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

const (
	idPrefixTrie = 101
)

type Trie struct {
	db    db.Store
	cache *NodePool
	log   *zap.SugaredLogger

	Root *TrieNode

	unSavedRefValueMap map[types.Hash][]byte
}

func NewTrie(db db.Store, rootHash *types.Hash, pool *NodePool) *Trie {
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
	txn := trie.db.NewTransaction(false)
	defer func() {
		txn.Commit(nil)
		txn.Discard()
	}()

	trieNode := new(TrieNode)
	k := trie.encodeKey(key[:])
	if err := txn.Get(k, func(val []byte, b byte) error {
		if len(val) > 0 {
			return trieNode.Deserialize(val)
		}
		return errors.New("invalid trie node data")
	}); err == nil {
		return trieNode
	} else {
		return nil
	}
}

func (trie *Trie) saveNodeToDb(txn db.StoreTxn, node *TrieNode) error {
	if data, err := node.Serialize(); err != nil {
		return fmt.Errorf("serialize trie node failed, error is %s", err)
	} else {
		h := node.Hash()
		k := trie.encodeKey(h[:])
		trie.log.Debugf("save node %s, %s", hex.EncodeToString(k), node.String())
		err := txn.Set(k, data)
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

func (trie *Trie) saveRefValueMap(txn db.StoreTxn) {
	for key, value := range trie.unSavedRefValueMap {
		k := trie.encodeKey(key[:])
		err := txn.Set(k, value)
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
	txn := trie.db.NewTransaction(false)
	defer func() {
		txn.Commit(nil)
		txn.Discard()
	}()

	k := trie.encodeKey(key[:])
	var result []byte
	if err = txn.Get(k, func(i []byte, b byte) error {
		result = make([]byte, len(i))
		copy(result, i)
		return nil
	}); err == nil {
		return result, nil
	} else {
		return nil, err
	}
}

func (trie *Trie) getNode(key *types.Hash) *TrieNode {
	if trie.cache != nil {
		if node := trie.cache.Get(key); node != nil {
			return node
		}
	}

	node := trie.getNodeFromDb(key)
	if node != nil && trie.cache != nil {
		trie.log.Debug("load from db ", node)
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
		newTrie.Root = trie.Root.Clone()
	}
	return newTrie
}

func (trie *Trie) Save(txns ...db.StoreTxn) (func(), error) {
	txn := trie.db.NewTransaction(true)
	defer func() {
		txn.Commit(nil)
		txn.Discard()
	}()

	err := trie.traverseSave(txn, trie.Root)
	if err != nil {
		return nil, err
	}

	trie.saveRefValueMap(txn)

	return func() {
		trie.unSavedRefValueMap = make(map[types.Hash][]byte)
	}, nil
}

func (trie *Trie) SaveInTxn(txn db.StoreTxn) (func(), error) {
	err := trie.traverseSave(txn, trie.Root)
	if err != nil {
		return nil, err
	}

	trie.saveRefValueMap(txn)

	return func() {
		trie.unSavedRefValueMap = make(map[types.Hash][]byte)
	}, nil
}

func (trie *Trie) traverseSave(txn db.StoreTxn, node *TrieNode) error {
	if node == nil {
		return nil
	}

	// Cached, no save
	if trie.getNode(node.Hash()) != nil {
		return nil
	}

	err := trie.saveNodeToDb(txn, node)
	if err != nil {
		return err
	}

	switch node.NodeType() {
	case FullNode:
		if node.child != nil {
			err := trie.traverseSave(txn, node.child)
			if err != nil {
				return err
			}
		}

		for _, child := range node.children {
			err := trie.traverseSave(txn, child)
			if err != nil {
				return err
			}
		}

	case ShortNode:
		err := trie.traverseSave(txn, node.child)
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

	go func(node *TrieNode) {
		if fn == nil || fn(node) {
			switch node.NodeType() {
			case FullNode:
				for _, child := range node.children {
					ch <- child
				}
				if node.child != nil {
					ch <- node.child
				}
			case ShortNode:
				ch <- node.child
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
		newNode := node.Clone()

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

				newNode := node.Clone()
				newNode.SetChild(leafNode)

				return newNode
			} else if node.child.NodeType() == FullNode {
				fullNode = node.child.Clone()
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

func (trie *Trie) encodeKey(key []byte) []byte {
	result := make([]byte, len(key)+1)
	result[0] = idPrefixTrie
	copy(result[1:], key)
	return result
}
