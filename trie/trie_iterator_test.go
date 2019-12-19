/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package trie

import (
	"fmt"
	"strings"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestNewIterator(t *testing.T) {
	teardownTestCase, trie := setupTestCase(t)
	defer teardownTestCase(t)

	var key1 []byte
	key2 := []byte("IamG")
	key3 := []byte("IamGood")
	key4 := []byte("tesab")
	key5 := []byte("tesa")
	key6 := []byte("tes")
	key7 := []byte("tesabcd")
	key8 := []byte("t")
	key9 := []byte("te")

	value1 := []byte("NilNilNilNilNil")
	value2 := []byte("ki10$%^%&@#!@#")
	value3 := []byte("a1230xm90zm19ma")
	value4 := []byte("value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555")
	value5 := []byte("value.555val")
	value6 := []byte("vale....asdfasdfasdfvalue.555val")
	value7 := []byte("asdfvale....asdfasdfasdfvalue.555val")
	value8 := []byte("asdfvale....asdfasdfasdfvalue.555valasd")
	value9 := []byte("AVDED09%^$%@#@#")

	trie.SetValue(key1, value1)
	trie.SetValue(key2, value2)
	trie.SetValue(key3, value3)
	trie.SetValue(key4, value4)

	trie.SetValue(key4, value5)
	trie.SetValue(key5, value6)
	trie.SetValue(key5, value6)
	trie.SetValue(key6, value7)
	trie.SetValue(key7, value7)
	trie.SetValue(key8, value8)
	trie.SetValue(key9, value9)

	iterator := trie.NewIterator([]byte("t"))
	for {
		key, value, ok := iterator.Next()
		if !ok {
			break
		}

		fmt.Printf("%s: %s\n", key, value)
	}
	fmt.Println()

	iterator1 := trie.NewIterator([]byte("te"))
	for {
		key, value, ok := iterator1.Next()
		if !ok {
			break
		}

		fmt.Printf("%s: %s\n", key, value)
	}
	fmt.Println()

	iterator2 := trie.NewIterator([]byte("I"))
	for {
		key, value, ok := iterator2.Next()
		if !ok {
			break
		}

		fmt.Printf("%s: %s\n", key, value)
	}
	fmt.Println()

	iterator3 := trie.NewIterator(nil)
	for {
		key, value, ok := iterator3.Next()
		if !ok {
			break
		}

		fmt.Printf("%s: %s\n", key, value)
	}
	fmt.Println()
}

func TestIteratorByRoot(t *testing.T) {
	teardownTestCase, trie := setupTestCase(t)
	defer teardownTestCase(t)

	store := trie.db
	cache := make(map[types.Hash]types.Address)

	for i := 0; i < 10; i++ {
		k := mock.Hash()
		v := mock.Address()
		trie.SetValue(k[:], v[:])
		cache[k] = v
	}
	root := trie.Root.Hash()
	if callback, err := trie.Save(); err == nil {
		callback()
	}
	t.Log(strings.Repeat("*", 64))

	for idx := 0; idx < 10; idx++ {
		t2 := NewTrie(store, nil, NewSimpleTrieNodePool())
		for i := 0; i < 10; i++ {
			k := mock.Hash()
			v := mock.Address()
			trie.SetValue(k[:], v[:])
		}
		if callback, err := t2.Save(); err == nil {
			callback()
		}
	}
	t.Log("TO FETCH ALL NODES BY ROOT")
	t3 := NewTrie(store, root, NewSimpleTrieNodePool())
	iterator := t3.NewIterator(nil)
	for {
		if key, value, ok := iterator.Next(); !ok {
			break
		} else {
			hash, _ := types.BytesToHash(key)
			address, _ := types.BytesToAddress(value)
			if v, ok := cache[hash]; ok && v == address {
				t.Logf("%s: %s\n", hash, address)
			} else {
				t.Errorf("can not find %s, %s->%s", hash, v, address)
			}
		}
	}
	t.Log(strings.Repeat("-", 64))

	t.Log("TO REMOVE NODES BY ROOT")
	t4 := NewTrie(store, root, NewSimpleTrieNodePool())
	if err := t4.Remove(); err != nil {
		t.Fatal(err)
	}

	t5 := NewTrie(store, root, NewSimpleTrieNodePool())
	newIterator := t5.NewIterator(nil)
	counter := 0
	for {
		if key, value, ok := newIterator.Next(); !ok {
			break
		} else {
			counter++
			hash, _ := types.BytesToHash(key)
			address, _ := types.BytesToAddress(value)
			t.Logf("recover %s: %s\n", hash, address)
		}
	}
	if counter > 0 {
		t.Fatal("failed to remove nodes", counter)
	}
}
