/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package trie

import (
	"encoding/hex"
	"sort"
	"sync"

	"github.com/qlcchain/go-qlc/common/util"

	"github.com/qlcchain/go-qlc/common/types"
)

const (
	UnknownNode = byte(iota)
	FullNode
	ShortNode
	ValueNode
	HashNode
)

type TrieNode struct {
	hash     *types.Hash
	nodeType byte

	// fullNode
	children map[byte]*TrieNode
	lock     sync.RWMutex

	// shortNode
	key   []byte
	child *TrieNode

	// hashNode and valueNode
	value []byte
}

func NewFullNode(children map[byte]*TrieNode) *TrieNode {
	if children == nil {
		children = make(map[byte]*TrieNode)
	}
	node := &TrieNode{
		children: children,
		nodeType: FullNode,
	}

	return node
}

func NewShortNode(key []byte, child *TrieNode) *TrieNode {
	node := &TrieNode{
		key:   key,
		child: child,

		nodeType: ShortNode,
	}

	return node
}

func NewHashNode(hash *types.Hash) *TrieNode {
	node := &TrieNode{
		value:    hash[:],
		nodeType: HashNode,
	}

	return node
}

func NewValueNode(value []byte) *TrieNode {
	node := &TrieNode{
		value:    value,
		nodeType: ValueNode,
	}

	return node
}

func (t *TrieNode) Value() []byte {
	return t.value
}

func (t *TrieNode) LeafCallback(completeFunc func()) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.nodeType == FullNode {
		if t.child != nil {
			if t.child.nodeType == UnknownNode {
				completeFunc()
			}
			return
		}

		for _, child := range t.children {
			if child.nodeType == UnknownNode {
				completeFunc()
			}
			return
		}

	}
}

func (t *TrieNode) Clone(copyHash bool) *TrieNode {
	t.lock.RLock()
	defer t.lock.RUnlock()

	newNode := &TrieNode{
		nodeType: t.nodeType,

		child: t.child,
	}

	if t.children != nil {
		newNode.children = make(map[byte]*TrieNode)
		for key, child := range t.children {
			newNode.children[key] = child
		}
	}

	newNode.key = make([]byte, len(t.key))
	copy(newNode.key, t.key)

	newNode.value = make([]byte, len(t.value))
	copy(newNode.value, t.value)

	if copyHash && t.hash != nil {
		newHash := *(t.hash)
		newNode.hash = &newHash
	}

	return newNode
}

func (t *TrieNode) IsLeafNode() bool {
	return t.nodeType == HashNode || t.nodeType == ValueNode
}

func (t *TrieNode) Hash() *types.Hash {
	if t.hash == nil {
		var source []byte
		switch t.NodeType() {
		case FullNode:
			source = []byte{FullNode}
			if t.child != nil {
				source = append(source, t.child.Hash()[:]...)
			}

			sc := sortChildren(t.children)
			for _, c := range sc {
				source = append(source, c.Key)
				source = append(source, c.Value.Hash()[:]...)
			}
		case ShortNode:
			source = []byte{ShortNode}
			source = append(source, t.key[:]...)
			source = append(source, t.child.Hash()[:]...)
		case HashNode:
			source = []byte{HashNode}
			source = t.value
		case ValueNode:
			source = []byte{ValueNode}
			source = t.value
		}

		hash := types.HashData(source)
		t.hash = &hash
	}
	return t.hash
}

func (t *TrieNode) SetChild(child *TrieNode) {
	t.child = child
}

func (t *TrieNode) NodeType() byte {
	return t.nodeType
}

func (t *TrieNode) Children() map[byte]*TrieNode {
	return t.children
}

func (t *TrieNode) SortedChildren() []*TrieNode {
	scKV := sortChildren(t.children)
	var scRet []*TrieNode
	for _, sc := range scKV {
		scRet = append(scRet, sc.Value)
	}
	return scRet
}

func (t *TrieNode) String() string {
	tn := &types.TrieNode{
		Type: t.NodeType(),
	}
	tn.Hash = t.Hash()
	switch t.NodeType() {
	case FullNode:
		tn.Children = t.serializeChildren(t.children)
		if t.child != nil {
			tn.Child = t.child.Hash()[:]
		}
	case ShortNode:
		tn.Key = t.key
		tn.Child = t.child.Hash()[:]
	case HashNode:
		fallthrough
	case ValueNode:
		tn.Value = t.value
	}

	return util.ToIndentString(tn)
}

func (t *TrieNode) serializeChildren(children map[byte]*TrieNode) map[string][]byte {
	if children == nil {
		return nil
	}

	var parsedChildren = make(map[string][]byte, len(children))
	for key, child := range children {
		ks := hex.EncodeToString([]byte{key})
		parsedChildren[ks] = child.Hash()[:]
	}
	return parsedChildren
}

func (t *TrieNode) Serialize() ([]byte, error) {
	tn := &types.TrieNode{
		Type: t.NodeType(),
	}
	tn.Hash = t.Hash()
	switch t.NodeType() {
	case FullNode:
		tn.Children = t.serializeChildren(t.children)
		if t.child != nil {
			tn.Child = t.child.Hash()[:]
		}
	case ShortNode:
		tn.Key = t.key
		tn.Child = t.child.Hash()[:]
	case HashNode:
		fallthrough
	case ValueNode:
		tn.Value = t.value
	}

	return tn.MarshalMsg(nil)
}

func (t *TrieNode) parseChildren(children map[string][]byte) (map[byte]*TrieNode, error) {
	var result = make(map[byte]*TrieNode)
	for key, child := range children {
		childHash, err := types.BytesToHash(child)
		if err != nil {
			return nil, err
		}
		tmp, err := hex.DecodeString(key)
		if err != nil {
			return nil, err
		}
		result[tmp[0]] = &TrieNode{
			hash: &childHash,
		}
	}

	return result, nil
}

func (t *TrieNode) Deserialize(buf []byte) error {
	tn := &types.TrieNode{}
	if _, err := tn.UnmarshalMsg(buf); err != nil {
		return err
	}

	t.nodeType = tn.Type
	t.hash = tn.Hash
	switch tn.Type {
	case FullNode:
		var err error
		t.children, err = t.parseChildren(tn.Children)
		if err != nil {
			return err
		}

		if len(tn.Child) > 0 {
			childHash, err := types.BytesToHash(tn.Child)
			if err != nil {
				return err
			}
			t.child = &TrieNode{
				hash: &childHash,
			}
		}

	case ShortNode:
		t.key = tn.Key
		childHash, err := types.BytesToHash(tn.Child)
		if err != nil {
			return err
		}
		t.child = &TrieNode{
			hash: &childHash,
		}
	case HashNode:
		fallthrough
	case ValueNode:
		t.value = tn.Value
	}

	return nil
}

type children struct {
	Key   byte
	Value *TrieNode
}

func sortChildren(c map[byte]*TrieNode) []*children {
	var s []*children
	for key, child := range c {
		s = append(s, &children{
			Key:   key,
			Value: child,
		})
	}

	sort.Slice(s, func(i, j int) bool {
		return s[i].Key < s[j].Key
	})
	return s
}
