/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *VMContext, ledger.Store) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "vm", uuid.New().String())
	_ = os.RemoveAll(dir)

	cm := config.NewCfgManager(dir)
	cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)

	block, td := mock.GeneratePovBlock(nil, 0)
	if err := l.AddPovBlock(block, td); err != nil {
		t.Fatal(err)
	}
	if err := l.AddPovBestHash(block.GetHeight(), block.GetHash()); err != nil {
		t.Fatal(err)
	}
	if err := l.SetPovLatestHeight(block.GetHeight()); err != nil {
		t.Fatal(err)
	}

	v := NewVMContext(l, nil)

	return func(t *testing.T) {
		//err := v.db.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, v, l
}

func TestLedger_Storage(t *testing.T) {
	teardownTestCase, ctx, l := setupTestCase(t)
	defer teardownTestCase(t)

	prefix := mock.Hash()
	key := []byte{10, 20, 30}
	value := []byte{10, 20, 30, 40}
	if err := ctx.SetStorage(prefix[:], key, value); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		if err := ctx.SetStorage(prefix[:], []byte{10, 20, 40, byte(i)}, value); err != nil {
			t.Fatal(err)
		}
	}

	v, err := ctx.GetStorage(prefix[:], key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(value, v) {
		t.Fatal("err store")
	}

	storageKey := getStorageKey(prefix[:], key)
	if get, err := ctx.get(storageKey); err == nil {
		t.Fatal("invalid storage", err)
	} else {
		t.Log(get, err)
	}

	cache := ToCache(ctx)
	err = l.SaveStorage(cache)
	if err != nil {
		t.Fatal(err)
	}

	if get, err := ctx.get(storageKey); get == nil {
		t.Fatal("invalid storage", err)
	} else {
		if !bytes.EqualFold(get, value) {
			t.Fatalf("invalid val,exp:%v, got:%v", value, get)
		} else {
			t.Log(get, err)
		}
	}

	counter := 0
	if err = ctx.Iterator(prefix[:], func(key []byte, value []byte) error {
		counter++
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	const iteratorSize = 11
	if counter != iteratorSize {
		t.Fatalf("failed to iterator context data,exp: 11, got: %d", counter)
	}

	storage := ctx.cache.storage
	if len(storage) != iteratorSize {
		t.Fatal("failed to iterator cache data")
	}
	cacheTrie := ctx.cache.Trie()
	if cacheTrie == nil {
		t.Fatal("invalid trie")
	}

	if hash := cacheTrie.Hash(); hash == nil {
		t.Fatal("invalid hash")
	} else {
		t.Log(hash.String())
	}
}

func getStorageKey(prefix []byte, key []byte) []byte {
	var storageKey []byte
	storageKey = append(storageKey, prefix...)
	storageKey = append(storageKey, key...)
	return storageKey
}

func TestVMCache_AppendLog(t *testing.T) {
	teardownTestCase, ctx, _ := setupTestCase(t)
	defer teardownTestCase(t)

	ctx.cache.AppendLog(&types.VmLog{
		Topics: []types.Hash{mock.Hash()},
		Data:   []byte{10, 20, 30, 40},
	})

	ctx.cache.AppendLog(&types.VmLog{
		Topics: []types.Hash{mock.Hash()},
		Data:   []byte{10, 20, 30, 50},
	})

	logs := ctx.cache.LogList()
	for idx, l := range logs.Logs {
		t.Log(idx, " >>>", l.Topics, ":", l.Data)
	}

	ctx.cache.Clear()

	if len(ctx.cache.logList.Logs) != 0 {
		t.Fatal("invalid logs ")
	}
}

func TestVMContext_IsUserAccount(t *testing.T) {
	teardownTestCase, ctx, _ := setupTestCase(t)
	defer teardownTestCase(t)

	if b, err := ctx.IsUserAccount(mock.Address()); err == nil || b {
		t.Fatal()
	}
}
