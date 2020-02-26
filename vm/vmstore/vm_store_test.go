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
	"sync"
	"testing"

	"github.com/qlcchain/go-qlc/common/storage"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *VMContext) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "vm", uuid.New().String())
	_ = os.RemoveAll(dir)

	cm := config.NewCfgManager(dir)
	cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)

	v := NewVMContext(l)

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
	}, v
}

func TestLedger_Storage(t *testing.T) {
	teardownTestCase, context := setupTestCase(t)
	defer teardownTestCase(t)

	prefix := mock.Hash()
	key := []byte{10, 20, 30}
	value := []byte{10, 20, 30, 40}
	if err := context.SetStorage(prefix[:], key, value); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		if err := context.SetStorage(prefix[:], []byte{10, 20, 40, byte(i)}, value); err != nil {
			t.Fatal(err)
		}
	}

	v, err := context.GetStorage(prefix[:], key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(value, v) {
		t.Fatal("err store")
	}

	storageKey := getStorageKey(prefix[:], key)
	if get, err := context.get(storageKey); err == nil {
		t.Fatal("invalid storage", err)
	} else {
		t.Log(get, err)
	}

	err = context.SaveStorage()
	if err != nil {
		t.Fatal(err)
	}

	if get, err := context.get(storageKey); get == nil {
		t.Fatal("invalid storage", err)
	} else {
		if !bytes.EqualFold(get, value) {
			t.Fatal("invalid val")
		} else {
			t.Log(get, err)
		}
	}

	if err := context.RemoveStorage(prefix[:], key); err != nil {
		t.Fatal(err)
	}

	if storage, err := context.GetStorage(prefix[:], key); err == nil && storage != nil {
		t.Fatal("failed to delete storage")
	}

	counter := 0
	if err = context.Iterator(prefix[:], func(key []byte, value []byte) error {
		counter++
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if counter != 10 {
		t.Fatal("failed to iterator context data")
	}

	storage := context.Cache.Storage()
	if len(storage) != 10 {
		t.Fatal("failed to iterator cache data")
	}
	cacheTrie := context.Cache.Trie()
	if cacheTrie == nil {
		t.Fatal("invalid trie")
	}

	if hash := cacheTrie.Hash(); hash == nil {
		t.Fatal("invalid hash")
	} else {
		t.Log(hash.String())
	}
	if err = context.SaveTrie(); err != nil {
		t.Fatal(err)
	}
}

func TestGetStorageKey(t *testing.T) {
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prefix := make([]byte, 10)
			key := make([]byte, 5)
			storageKey := getStorageKey(prefix, key)
			if storageKey[0] != storage.KeyPrefixTrieVMStorage {
				t.Errorf("invalid prefix 0x%x", storageKey[:1])
			}
			if !bytes.HasPrefix(storageKey[1:], prefix) {
				t.Error("invalid prefix data")
			}

			if !bytes.HasSuffix(storageKey, key) {
				t.Error("invalid key data")
			}
		}()
	}

	wg.Wait()
}

func TestVMCache_AppendLog(t *testing.T) {
	teardownTestCase, ctx := setupTestCase(t)
	defer teardownTestCase(t)

	ctx.Cache.AppendLog(&types.VmLog{
		Topics: []types.Hash{mock.Hash()},
		Data:   []byte{10, 20, 30, 40},
	})

	ctx.Cache.AppendLog(&types.VmLog{
		Topics: []types.Hash{mock.Hash()},
		Data:   []byte{10, 20, 30, 50},
	})

	logs := ctx.Cache.LogList()
	for idx, l := range logs.Logs {
		t.Log(idx, " >>>", l.Topics, ":", l.Data)
	}

	ctx.Cache.Clear()

	if len(ctx.Cache.logList.Logs) != 0 {
		t.Fatal("invalid logs ")
	}
}
