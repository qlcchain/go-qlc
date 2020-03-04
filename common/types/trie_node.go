/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

//go:generate msgp
type TrieNode struct {
	Hash     *Hash             `msg:"hash,extension" json:"hash"`
	Type     byte              `msg:"type" json:"type"`
	Children map[string][]byte `msg:"children" json:"children,omitempty"`
	Key      []byte            `msg:"key" json:"key,omitempty"`
	Child    []byte            `msg:"child" json:"child,omitempty"`
	Value    []byte            `msg:"value" json:"value,omitempty"`
}
