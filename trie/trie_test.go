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
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/storage/db"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *Trie) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "trie", uuid.New().String())
	_ = os.RemoveAll(dir)
	store, err := db.NewBadgerStore(dir)
	//t.Log("db ", dir)
	if err != nil {
		fmt.Println(err.Error())
		return func(t *testing.T) {

		}, nil
	}
	trie := NewTrie(store, nil, NewSimpleTrieNodePool())
	return func(t *testing.T) {
		_ = store.Close()
	}, trie
}

func TestSetGetCase1(t *testing.T) {
	teardownTestCase, trie := setupTestCase(t)
	defer teardownTestCase(t)

	key := []byte("tesabcd")
	value := []byte("value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4")

	key2 := []byte("tesab")
	value2 := []byte("value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555")

	trie.SetValue(key, value)
	trie.SetValue(key2, value2)

	getValue := trie.GetValue(key)
	getValue2 := trie.GetValue(key2)

	if !bytes.Equal(value, getValue) {
		t.Error("error!")
	}

	if !bytes.Equal(value2, getValue2) {
		t.Error("error!")
	}
}

func TestSetGetCase2(t *testing.T) {
	teardownTestCase, trie := setupTestCase(t)
	defer teardownTestCase(t)

	var key1 []byte
	value1 := []byte("NilNilNilNilNil")

	key2 := []byte("IamG")
	value2 := []byte("ki10$%^%&@#!@#")

	key3 := []byte("IamGood")
	value3 := []byte("a1230xm90zm19ma")

	trie.SetValue(key1, value1)
	trie.SetValue(key2, value2)
	trie.SetValue(key3, value3)

	getValue1 := trie.GetValue(key1)
	getValue2 := trie.GetValue(key2)
	getValue3 := trie.GetValue(key3)
	if !bytes.Equal(value1, getValue1) {
		t.Error("error!")
	}
	if !bytes.Equal(value2, getValue2) {
		t.Error("error!")
	}
	if !bytes.Equal(value3, getValue3) {
		t.Error("error!")
	}
}

func TestNewTrie(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "trie", uuid.New().String())
	_ = os.RemoveAll(dir)
	store, err := db.NewBadgerStore(dir)
	defer func() {
		if store != nil {
			_ = store.Close()
		}
	}()
	if err != nil {
		t.Fatal(err)
	}

	pool := NewSimpleTrieNodePool()

	trie := NewTrie(store, nil, pool)

	// case1
	value1 := []byte("NilNilNilNilNil")
	trie.SetValue(nil, value1)
	getValue1 := trie.GetValue(nil)
	if !bytes.Equal(value1, getValue1) {
		t.Error("error!")
	}
	if len(trie.unSavedRefValueMap) != 0 {
		t.Error("error!")
	}
	// case2
	value2 := []byte("NilNilNilNilNil234")
	trie.SetValue(nil, value2)
	getValue2 := trie.GetValue(nil)
	if !bytes.Equal(value2, getValue2) {
		t.Error("error!")
	}
	if len(trie.unSavedRefValueMap) != 0 {
		t.Error("error!")
	}

	// case3
	key3 := []byte("test")
	value3 := []byte("value.hash")
	trie.SetValue(key3, value3)

	getValue3 := trie.GetValue(key3)
	if !bytes.Equal(value3, getValue3) {
		t.Error("error!")
	}
	if len(trie.unSavedRefValueMap) != 0 {
		t.Error("error!")
	}

	// case4
	key4 := []byte("tesa")
	value4 := []byte("value.hash2")
	trie.SetValue(key4, value4)

	getValue4 := trie.GetValue(nil)
	if !bytes.Equal(value2, getValue4) {
		t.Error("error!")
	}
	getValue4_2 := trie.GetValue(key3)
	if !bytes.Equal(value3, getValue4_2) {
		t.Error("error!")
	}

	getValue4_3 := trie.GetValue(key4)
	if !bytes.Equal(value4, getValue4_3) {
		t.Error("error!")
	}
	if len(trie.unSavedRefValueMap) != 0 {
		t.Error("error!")
	}

	// case 5
	key5 := []byte("aofjas")
	value5 := []byte("value.content1")
	trie.SetValue(key5, value5)

	getValue5 := trie.GetValue(nil)
	if !bytes.Equal(value2, getValue5) {
		t.Error("error!")
	}

	getValue5_1 := trie.GetValue(key3)
	if !bytes.Equal(value3, getValue5_1) {
		t.Error("error!")
	}

	getValue5_2 := trie.GetValue(key4)
	if !bytes.Equal(value4, getValue5_2) {
		t.Error("error!")
	}

	getValue5_3 := trie.GetValue(key5)
	if !bytes.Equal(value5, getValue5_3) {
		t.Error("error!")
	}

	if len(trie.unSavedRefValueMap) != 0 {
		t.Error("error!")
	}

	// case6
	fmt.Println(6)
	trie.SetValue([]byte("aofjas"), []byte("value.content2value.content2value.content2value.content2value.content2value.content2value.content2value.content2"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	// case7
	fmt.Println(7)
	value7 := []byte("value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3")
	trie.SetValue(key4, value7)
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println(8)
	value8 := []byte("value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value.hash3value09909")
	trie.SetValue(key4, value8)
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println(9)
	trie.SetValue([]byte("tesabcd"), []byte("value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4value.hash4"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println(10)
	// case10
	key10 := []byte("tesab")
	value10 := []byte("value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555")

	trie.SetValue(key10, value10)
	getValue10 := trie.GetValue(nil)
	if !bytes.Equal(value2, getValue10) {
		t.Error("error!")
	}
	getValue10_1 := trie.GetValue(key3)
	if !bytes.Equal(value3, getValue10_1) {
		t.Error("error!")
	}

	getValue10_2 := trie.GetValue(key4)
	if !bytes.Equal(value8, getValue10_2) {
		t.Error("error!")
	}

	getValue10_3 := trie.GetValue(key10)
	if !bytes.Equal(value10, getValue10_3) {
		t.Error("error!")
	}
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println("10.1")
	trie.SetValue([]byte("tesab"), []byte("value.555"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println(11)
	trie.SetValue([]byte("t"), []byte("somethinghiOkYesYourMyASDKJBNXA1239xnm.0j8n120k0k12nz$0231*&^$@!!())$S@@ST&&@@SDT&(OL<><:PP_+}}GC~@@@#$%^&&HCXZkasldjf1009100"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", trie.GetValue([]byte("t")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println("11.1")
	trie.SetValue([]byte("t"), []byte("abc"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", trie.GetValue([]byte("t")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println(12)
	trie.SetValue([]byte("a"), []byte("a1230xm9"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", trie.GetValue([]byte("t")))
	fmt.Printf("%s\n", trie.GetValue([]byte("a")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println("12.1")
	trie.SetValue([]byte("a"), []byte("a10xm9"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", trie.GetValue([]byte("t")))
	fmt.Printf("%s\n", trie.GetValue([]byte("a")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println(13)
	trie.SetValue([]byte("IamGood"), []byte("a1230xm90zm19ma"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", trie.GetValue([]byte("t")))
	fmt.Printf("%s\n", trie.GetValue([]byte("a")))
	fmt.Printf("%s\n", trie.GetValue([]byte("IamGood")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println(14)
	trie.SetValue([]byte("IamGood"), []byte("hahaheheh"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", trie.GetValue([]byte("t")))
	fmt.Printf("%s\n", trie.GetValue([]byte("a")))
	fmt.Printf("%s\n", trie.GetValue([]byte("IamGood")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println(15)
	trie.SetValue([]byte("IamGoo"), []byte("ijukh"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", trie.GetValue([]byte("t")))
	fmt.Printf("%s\n", trie.GetValue([]byte("a")))
	fmt.Printf("%s\n", trie.GetValue([]byte("IamGood")))
	fmt.Printf("%s\n", trie.GetValue([]byte("IamGoo")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println(16)
	trie.SetValue([]byte("IamG"), []byte("ki10$%^%&@#!@#"))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", trie.GetValue([]byte("t")))
	fmt.Printf("%s\n", trie.GetValue([]byte("a")))
	fmt.Printf("%s\n", trie.GetValue([]byte("IamGood")))
	fmt.Printf("%s\n", trie.GetValue([]byte("IamGoo")))
	fmt.Printf("%s\n", trie.GetValue([]byte("IamG")))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()

	fmt.Println(17)
	trie.SetValue(nil, []byte("isNil"))
	fmt.Printf("%s\n", trie.GetValue([]byte("test")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", trie.GetValue([]byte("aofjas")))
	fmt.Printf("%s\n", trie.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", trie.GetValue([]byte("t")))
	fmt.Printf("%s\n", trie.GetValue([]byte("a")))
	fmt.Printf("%s\n", trie.GetValue([]byte("IamGood")))
	fmt.Printf("%s\n", trie.GetValue([]byte("IamGoo")))
	fmt.Printf("%s\n", trie.GetValue([]byte("IamG")))
	fmt.Printf("%s\n", trie.GetValue(nil))
	fmt.Printf("%d\n", len(trie.unSavedRefValueMap))
	fmt.Println()
}

func TestTrieHash(t *testing.T) {
	teardownTestCase, trie := setupTestCase(t)
	defer teardownTestCase(t)

	// case1
	trie.SetValue(nil, []byte("NilNilNilNilNil"))
	hash1 := trie.Hash()
	if hash1.String() != "c6cafbbd9f060a8cde7e159d378c76e12ecbc36fcd6125ee51b81d316f019ef1" {
		t.Error("errro!")
	}
	hash1_2 := trie.Hash()
	if hash1_2.String() != "c6cafbbd9f060a8cde7e159d378c76e12ecbc36fcd6125ee51b81d316f019ef1" {
		t.Error("errro!")
	}

	hash1_3 := trie.Hash()
	if hash1_3.String() != "c6cafbbd9f060a8cde7e159d378c76e12ecbc36fcd6125ee51b81d316f019ef1" {
		t.Error("errro!")
	}

	// case2
	trie.SetValue(nil, []byte("isNil"))
	hash2 := trie.Hash()
	if hash2.String() != "402d3ba71597bb87129ada70588db179817a886a97a5b22e6d8b930cdd673d04" {
		t.Error("errro!")
	}
	hash2_2 := trie.Hash()
	if hash2_2.String() != "402d3ba71597bb87129ada70588db179817a886a97a5b22e6d8b930cdd673d04" {
		t.Error("errro!")
	}

	hash2_3 := trie.Hash()
	if hash2_3.String() != "402d3ba71597bb87129ada70588db179817a886a97a5b22e6d8b930cdd673d04" {
		t.Error("errro!")
	}

	trie.SetValue([]byte("IamG"), []byte("ki10$%^%&@#!@#"))
	fmt.Println(trie.Hash())
	trie.SetValue([]byte("IamGood"), []byte("a1230xm90zm19ma"))
	fmt.Println(trie.Hash())
	fmt.Println(trie.Hash())
	fmt.Println(trie.Hash())
	fmt.Println(trie.Hash())
	fmt.Println(trie.Hash())
	trie.SetValue([]byte("tesab"), []byte("value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555"))
	fmt.Println(trie.Hash())
	fmt.Println(trie.Hash())
	fmt.Println(trie.Hash())
	fmt.Println(trie.Hash())
	fmt.Println(trie.Hash())
	trie.SetValue([]byte("tesab"), []byte("value.555val"))
	fmt.Println(trie.Hash())
	trie.SetValue([]byte("tesa"), []byte("vale....asdfasdfasdfvalue.555val"))
	fmt.Println(trie.Hash())
	trie.SetValue([]byte("tes"), []byte("asdfvale....asdfasdfasdfvalue.555val"))
	fmt.Println(trie.Hash())
	trie.SetValue([]byte("tesabcd"), []byte("asdfvale....asdfasdfasdfvalue.555val"))
	fmt.Println(trie.Hash())
	trie.SetValue([]byte("t"), []byte("asdfvale....asdfasdfasdfvalue.555valasd"))
	fmt.Println(trie.Hash())
}

func TestTrieSaveAndLoadCase1(t *testing.T) {
	teardownTestCase, trie := setupTestCase(t)
	defer teardownTestCase(t)

	var key1 []byte
	value1 := []byte("NilNilNilNilNil")

	key2 := []byte("IamG")
	value2 := []byte("ki10$%^%&@#!@#")

	key3 := []byte("IamGood")
	value3 := []byte("a1230xm90zm19ma")

	trie.SetValue(key1, value1)
	trie.SetValue(key2, value2)
	trie.SetValue(key3, value3)

	getValue1 := trie.GetValue(key1)
	getValue2 := trie.GetValue(key2)
	getValue3 := trie.GetValue(key3)
	if !bytes.Equal(value1, getValue1) {
		t.Error("error!")
	}
	if !bytes.Equal(value2, getValue2) {
		t.Error("error!")
	}
	if !bytes.Equal(value3, getValue3) {
		t.Error("error!")
	}

	// save db
	callback, err := trie.Save()
	if err != nil {
		t.Fatal(err)
	}
	callback()

	rootHash := trie.Hash()
	t.Log("root_hash: ", rootHash.String())
	store := trie.db
	cache := NewSimpleTrieNodePool()
	trie = nil

	newTrie := NewTrie(store, rootHash, cache)
	newGetValue1 := newTrie.GetValue(key1)
	newGetValue2 := newTrie.GetValue(key2)
	newGetValue3 := newTrie.GetValue(key3)

	if !bytes.Equal(value1, newGetValue1) {
		t.Error("error!", hex.EncodeToString(value1), "==>", hex.EncodeToString(newGetValue1))
	}
	if !bytes.Equal(value2, newGetValue2) {
		t.Error("error!", hex.EncodeToString(value2), "==>", hex.EncodeToString(newGetValue2))
	}
	if !bytes.Equal(value3, newGetValue3) {
		t.Error("error!", hex.EncodeToString(value3), "==>", hex.EncodeToString(newGetValue3))
	}
}

func TestTrieSaveAndLoad(t *testing.T) {
	teardownTestCase, trie := setupTestCase(t)
	defer teardownTestCase(t)

	trie.SetValue(nil, []byte("NilNilNilNilNil"))
	trie.SetValue([]byte("IamG"), []byte("ki10$%^%&@#!@#"))
	trie.SetValue([]byte("IamGood"), []byte("a1230xm90zm19ma"))
	trie.SetValue([]byte("tesab"), []byte("value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555"))

	trie.SetValue([]byte("tesab"), []byte("value.555val"))
	trie.SetValue([]byte("tesa"), []byte("vale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("tesa"), []byte("vale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("tes"), []byte("asdfvale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("tesabcd"), []byte("asdfvale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("t"), []byte("asdfvale....asdfasdfasdfvalue.555valasd"))
	fmt.Println(trie.Hash())
	fmt.Println()

	callback, err := trie.Save()
	if err != nil {
		t.Fatal(err)
	}
	callback()

	rootHash := trie.Hash()
	store := trie.db
	cache := NewSimpleTrieNodePool()
	trie = nil

	newTrie := NewTrie(store, rootHash, cache)

	fmt.Printf("%s\n", newTrie.GetValue([]byte("IamG")))
	fmt.Printf("%s\n", newTrie.GetValue([]byte("IamGood")))
	fmt.Printf("%s\n", newTrie.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", newTrie.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", newTrie.GetValue([]byte("tes")))
	fmt.Printf("%s\n", newTrie.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", newTrie.GetValue([]byte("t")))
	fmt.Println(newTrie.Hash())
	fmt.Println()
	newTrie = nil

	newTri2 := NewTrie(store, rootHash, cache)

	fmt.Printf("%s\n", newTri2.GetValue([]byte("IamG")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("IamGood")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tes")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("t")))
	fmt.Println(newTri2.Hash())
	fmt.Println()

	newTri2.SetValue([]byte("tesab"), []byte("value.hahaha123"))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("IamG")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("IamGood")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tes")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("t")))
	fmt.Println(newTri2.Hash())
	fmt.Println()

	newTri2.SetValue([]byte("IamGood"), []byte("Yes you are good."))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("IamG")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("IamGood")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tes")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("t")))
	fmt.Println(newTri2.Hash())
	fmt.Println()

	callback2, _ := newTri2.Save()
	callback2()

	rootHash2 := newTri2.Hash()
	newTrie3 := NewTrie(store, rootHash2, cache)
	fmt.Printf("%s\n", newTrie3.GetValue([]byte("IamG")))
	fmt.Printf("%s\n", newTrie3.GetValue([]byte("IamGood")))
	fmt.Printf("%s\n", newTrie3.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", newTrie3.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", newTrie3.GetValue([]byte("tes")))
	fmt.Printf("%s\n", newTrie3.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", newTrie3.GetValue([]byte("t")))
	fmt.Println(newTrie3.Hash())
	fmt.Println()

	fmt.Printf("%s\n", newTri2.GetValue([]byte("IamG")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("IamGood")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesab")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesa")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tes")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("tesabcd")))
	fmt.Printf("%s\n", newTri2.GetValue([]byte("t")))
	fmt.Println(newTri2.Hash())
	fmt.Println()
}

func TestTrieConcurrence(t *testing.T) {
	teardownTestCase, trie := setupTestCase(t)
	defer teardownTestCase(t)

	trie.SetValue(nil, []byte("NilNilNilNilNil"))
	trie.SetValue([]byte("IamG"), []byte("ki10$%^%&@#!@#"))
	trie.SetValue([]byte("IamGood"), []byte("a1230xm90zm19ma"))
	trie.SetValue([]byte("tesab"), []byte("value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555"))

	trie.SetValue([]byte("tesab"), []byte("value.555val"))
	trie.SetValue([]byte("tesa"), []byte("vale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("tesa"), []byte("vale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("tes"), []byte("asdfvale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("tesabcd"), []byte("asdfvale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("t"), []byte("asdfvale....asdfasdfasdfvalue.555valasd"))
	fmt.Println(trie.Hash())
	fmt.Println()

	var sw sync.WaitGroup
	errs := make(chan error)
	for i := 0; i < 1000; i++ {
		sw.Add(1)
		go func() {
			defer sw.Done()

			_, err := trie.Save()
			if err != nil {
				errs <- err
			}
		}()
	}

	go func() {
		sw.Wait()
		close(errs)
	}()

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	store := trie.db
	cache := NewSimpleTrieNodePool()

	trie = nil

	rootHash, _ := types.NewHash("ece19924b34f6bf264e6fcc7feaabe8481939a5eb1a1a7a9825468128f526797")
	for i := 0; i < 1000; i++ {
		var sw sync.WaitGroup
		for i := 0; i < 10; i++ {
			sw.Add(1)
			go func() {
				defer sw.Done()
				trie := NewTrie(store, &rootHash, cache)
				trie.SetValue([]byte("tes"), []byte("asdfvale....asdfasdfasdfvalue.555val"+strconv.FormatInt(time.Now().UnixNano(), 10)))
				trie.SetValue([]byte("tesab"), []byte("value.555val"+strconv.FormatInt(time.Now().UnixNano(), 10)))
				fmt.Printf("%s\n", trie.GetValue([]byte("tes")))
				fmt.Println(trie.Hash())
			}()
		}
		sw.Wait()
	}

	trie2 := NewTrie(store, &rootHash, cache)
	fmt.Printf("%s\n", trie2.GetValue([]byte("tes")))
	fmt.Printf("%s\n", trie2.GetValue([]byte("tesab")))
	fmt.Println(trie2.Hash())
}

func TestEncodeKey(t *testing.T) {
	h := mock.Hash()
	key := encodeKey(h[:])
	if len(key) != types.HashSize+1 {
		t.Fatal("invalid size")
	}
	if key[0] != idPrefixTrie {
		t.Fatal("invalid prefix")
	}

	if !bytes.HasSuffix(key, h[:]) {
		t.Fatal("invalid suffix")
	}

	t.Log(h[:])
	t.Log(key)
}

func TestTrie_Remove(t *testing.T) {
	teardownTestCase, trie := setupTestCase(t)
	defer teardownTestCase(t)

	trie.SetValue(nil, []byte("NilNilNilNilNil"))
	trie.SetValue([]byte("IamG"), []byte("ki10$%^%&@#!@#"))
	trie.SetValue([]byte("IamGood"), []byte("a1230xm90zm19ma"))
	trie.SetValue([]byte("tesab"), []byte("value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555"))

	trie.SetValue([]byte("tesab"), []byte("value.555val"))
	trie.SetValue([]byte("tesa"), []byte("vale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("tesa"), []byte("vale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("tes"), []byte("asdfvale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("tesabcd"), []byte("asdfvale....asdfasdfasdfvalue.555val"))
	trie.SetValue([]byte("t"), []byte("asdfvale....asdfasdfasdfvalue.555valasd"))
	root := trie.Hash()

	_, _ = trie.Save()
	store := trie.db

	t2 := NewTrie(store, root, NewSimpleTrieNodePool())
	if err := t2.Remove(); err != nil {
		t.Fatal(err)
	}

	t3 := NewTrie(store, root, NewSimpleTrieNodePool())
	if t3.Root != nil {
		t.Fatal("load trie error")
	}
}
