/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
)

func TestNewWalletStore(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String(), "wallet_test")
	cm := config.NewCfgManager(dir)
	cm.Load()
	cm2 := config.NewCfgManager(dir)
	cm2.Load()
	store1 := NewWalletStore(cm.ConfigFile)
	store2 := NewWalletStore(cm2.ConfigFile)
	if store1 == nil || store2 == nil {
		t.Fatal("error create store")
	}
	t.Logf("store1:%p, store2:%p", store1, store2)
	if !reflect.DeepEqual(store1, store2) {
		t.Fatal("store1!=store2")
	}
	defer func() {
		err := store1.Close()
		if err != nil {
			t.Fatal(err)
		}
		//store2.Close()
		_ = os.RemoveAll(dir)
	}()
}

func TestNewWalletStore2(t *testing.T) {
	dir1 := filepath.Join(config.QlcTestDataDir(), "wallet_test1")
	dir2 := filepath.Join(config.QlcTestDataDir(), "wallet_test2")
	cm1 := config.NewCfgManager(dir1)
	cm1.Load()
	cm2 := config.NewCfgManager(dir2)
	cm2.Load()
	store1 := NewWalletStore(cm1.ConfigFile)
	store2 := NewWalletStore(cm2.ConfigFile)
	if store1 == nil || store2 == nil {
		t.Fatal("error create store")
	}
	t.Logf("store1:%p, store2:%p", store1, store2)
	if reflect.DeepEqual(store1, store2) {
		t.Fatal("store1==store2")
	}
	defer func() {
		err := store1.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = store2.Close()
		if err != nil {
			t.Fatal(err)
		}
		_ = os.RemoveAll(dir1)
		_ = os.RemoveAll(dir2)
	}()
}

func TestWalletStore_NewWallet(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	ids, err := store.WalletIds()
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 0 {
		bytes, _ := json.Marshal(ids)
		t.Fatal("invalid ids", string(bytes))
	}

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(id.String())
	id2, err := store.CurrentId()
	if err != nil {
		t.Fatal(err)
	}
	if id != id2 {
		t.Fatal("id!=id2")
	}

	ids2, err := store.WalletIds()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids2) != 1 || ids2[0] != id2 {
		t.Fatal("ids2 failed")
	}

	err = store.RemoveWallet(id2)
	if err != nil {
		t.Fatal(err)
	}

	ids3, err := store.WalletIds()
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range ids3 {
		t.Log(id.String())
	}

	if len(ids3) > 0 {
		t.Fatal("invalid ids3 =>", len(ids3))
	}

	currentId, err := store.CurrentId()
	if err != nil && err != ErrEmptyCurrentId {
		t.Fatal(err)
	}
	t.Log(currentId.String())
}

func TestWalletStore_NewWalletBySeed(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	seed, err := types.NewSeed()
	if err != nil {
		t.Fatal(err)
	}

	account, err := store.NewWalletBySeed(seed.String(), "1111")

	if err != nil {
		t.Fatal(err)
	}
	s := store.NewSession(account)
	b, err := s.VerifyPassword("1111")
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("invalid password")
	}
}
