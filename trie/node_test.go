/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package trie

import (
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
	bytes, err := node.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	newNode := new(TrieNode)
	err = newNode.Deserialize(bytes)
	if err != nil {
		t.Fatal(err)
	}

	if newNode.nodeType != node.nodeType {
		t.Fatal("invalid type")
	}

	//if !reflect.DeepEqual(newNode.children, node.children) {
	//	t.Fatal("invalid children", "act ", util.ToIndentString(newNode.children),
	//		"exp ", util.ToIndentString(node.children))
	//}
}
