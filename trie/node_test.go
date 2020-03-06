/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package trie

import (
	"bytes"
	"testing"

	"github.com/qlcchain/go-qlc/mock"
)

func TestTrieNode_Deserialize(t *testing.T) {
	h1 := mock.Hash()
	node1 := &TrieNode{
		nodeType: HashNode,
		value:    h1[:],
	}
	node := &TrieNode{
		nodeType: FullNode,
		children: map[byte]*TrieNode{
			byte(73): node1,
		},
	}
	data, err := node.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	newNode := new(TrieNode)
	err = newNode.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if newNode.nodeType != node.nodeType {
		t.Fatal("invalid type")
	}

	v1 := node.Value()
	v2 := newNode.Value()
	if !bytes.Equal(v1, v2) {
		t.Fatal()
	}

	children := node.Children()
	if len(children) != 1 {
		t.Fatal()
	}

	nodes := node.SortedChildren()
	for _, n := range nodes {
		t.Log(n.String())
	}

	//if !reflect.DeepEqual(newNode.children, node.children) {
	//	t.Fatal("invalid children", "act ", util.ToIndentString(newNode.children),
	//		"exp ", util.ToIndentString(node.children))
	//}
}
