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

func TestNewSimpleTrieNodePool(t *testing.T) {
	if pool := GetGlobalTriePool(); pool == nil {
		t.Fatal()
	} else {
		h := mock.Hash()
		pool.Set(&h, nil)
		if get := pool.Get(&h); get != nil {
			t.Fatal()
		}
		if i := pool.Len(); i != 1 {
			t.Fatal()
		}
		pool.Clear()
		if i := pool.Len(); i != 0 {
			t.Fatal()
		}
	}

	if pool := NewSimpleTrieNodePool(); pool == nil {
		t.Fatal()
	}
}
