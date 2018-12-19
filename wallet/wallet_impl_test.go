/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/db"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *WalletStore) {
	t.Parallel()
	start := time.Now()
	cfg, _ := config.DefaultConfig()
	cfg.DataDir = filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	dir := cfg.WalletDir()
	t.Log("setup store test case", dir)
	_ = os.RemoveAll(dir)

	store := NewWalletStore(cfg)
	if store == nil {
		t.Fatal("create store failed")
	}
	t.Logf("NewWalletStore cost %s", time.Since(start))
	return func(t *testing.T) {
		err := store.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("teardown wallet test case: %s", time.Since(start))
	}, store
}

func TestWalletStore_NewSession(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	seedArray, err := session.GetSeed()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(seedArray))
}

func TestSession_ChangePassword(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	seedArray, err := session.GetSeed()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(seedArray))

	err = session.ChangePassword("37yBR94bvj4wbkYc")
	if err != nil {
		t.Fatal(err)
	}

	seedArray2, err := session.GetSeed()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(seedArray, seedArray2) {
		t.Fatal("seed mismatch")
	}
}

func TestSession_ValidPassword(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	seedArray, err := session.GetSeed()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(seedArray))

	err = session.ChangePassword("37yBR94bvj4wbkYd")
	if err != nil {
		t.Fatal(err)
	}

	isvalid := session.ValidPassword()

	if !isvalid {
		t.Fatal("password mismatch")
	}
}

func TestSession_IsAccountExist(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	seedArray, _ := session.GetSeed()
	seed, _ := types.BytesToSeed(seedArray)
	account, err := seed.Account(2)

	if err != nil {
		t.Fatal(err)
	}

	addr := account.Address()
	t.Log(addr.String())

	am := common.MockAccountMeta(addr)
	l := session.ledger

	err = l.AddAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}

	am2, err := l.GetAccountMeta(addr)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(am, am2) {
		t.Fatal("am!=am2")
	}

	if exist := session.IsAccountExist(addr); !exist {
		t.Fatal("IsAccountExist1 failed", addr.String())
	}

	account2, err := seed.Account(101)

	if err != nil {
		t.Fatal(err)
	}
	addr2 := account2.Address()

	if exist := session.IsAccountExist(addr2); exist {
		t.Fatal("IsAccountExist2 failed", addr2.String())
	}

	addr3 := common.MockAddress()
	if exist := session.IsAccountExist(addr3); exist {
		t.Fatal("IsAccountExist3 failed", addr2.String())
	}
}

func TestSession_GetRawKey(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	seedArray, err := session.GetSeed()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(seedArray))
	seed := hex.EncodeToString(seedArray)
	pub, priv, err := types.KeypairFromSeed(seed, 2)

	acc1 := types.NewAccount(priv)
	hash := common.MockHash()

	sign := acc1.Sign(hash)
	addr := types.PubToAddress(pub)

	acc2, err := session.GetRawKey(addr)
	if err != nil {
		t.Fatal(err)
	}

	if verify := acc2.Address().Verify(hash[:], sign[:]); !verify {
		t.Fatal("verify failed.")
	}

	addr2 := common.MockAddress()
	if _, err = session.GetRawKey(addr2); err == nil {
		t.Fatal("get invalid raw key failed")
	}
}

func TestSession_SetSeed(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	seed, err := types.NewSeed()
	if err != nil {
		t.Fatal(err)
	}
	err = session.SetSeed(seed[:])
	if err != nil {
		t.Fatal(err)
	}

	bytes, err := session.GetSeed()
	if err != nil {
		t.Fatal(err)
	}
	seed2 := types.Seed{}
	_ = seed2.UnmarshalBinary(bytes)

	if reflect.DeepEqual(seed2, seed) {
		t.Fatal("seed != seed")
	}
}

func TestSession_SetVersion(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	session := store.NewSession(id)
	t.Logf("NewSession cost %s", time.Since(start))
	start = time.Now()
	err = session.SetVersion(2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("SetVersion cost %s", time.Since(start))

	start = time.Now()
	i, err := session.GetVersion()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("GetVersion cost %s", time.Since(start))
	if i != 2 {
		t.Fatal("set version failed")
	}
}

func TestSession_DeterministicIndex(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	err = session.SetDeterministicIndex(10)
	if err != nil {
		t.Fatal(err)
	}
	i, err := session.GetDeterministicIndex()

	if err != nil {
		t.Fatal(err)
	}
	if i != 10 {
		t.Fatal("SetDeterministicIndex failed. ")
	}

	err = session.ResetDeterministicIndex()
	if err != nil {
		t.Fatal(err)
	}
	index, err := session.GetDeterministicIndex()
	if index != 0 {
		t.Fatal("ResetDeterministicIndex failed")
	}
}

func TestSession_GetWork(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	//h:=mockHash()
	addr := common.MockAddress()

	work, err := session.GetWork(addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(work.String())
}

func TestGenerateWork(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan string)
	go func() {
		session := store.NewSession(id)
		hash := common.MockHash()
		work := session.generateWork(hash)
		if !work.IsValid(hash) {
			t.Fatal("generateWork failed =>", hash.String())
		}
		done <- fmt.Sprintf("hash[%s]=>%s", hash.String(), work.String())
	}()

	t.Log(<-done)
}

func TestSession_GetAccounts(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	s := session.ledger

	seed, err := session.GetSeed()
	//seedString := hex.EncodeToString(seed)
	//addresses, err := insertAccountMeta(session, seedString, t)
	ss, _ := types.BytesToSeed(seed)
	var addresses []types.Address
	err = s.BatchUpdate(func(txn db.StoreTxn) error {
		for i := 0; i < 5; i++ {
			account, _ := ss.Account(uint32(i))
			addr := account.Address()
			am := common.MockAccountMeta(addr)
			err = s.AddAccountMeta(am, txn)
			if err != nil {
				t.Fatal(err)
			}
			meta, err := s.GetAccountMeta(addr, txn)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(am, meta) {
				t.Log(am, meta)
				t.Fatal("save am failed")
			}
			addresses = append(addresses, addr)
			t.Log(addr.String())
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	_ = session.SetDeterministicIndex(5)

	accounts2, err := session.GetAccounts()
	if err != nil {
		t.Log(err)
	}

	if !reflect.DeepEqual(addresses, accounts2) {
		t.Log("addresses", strings.Repeat("*", 20))
		for _, a := range addresses {
			t.Log(a.String())
			meta, err := s.GetAccountMeta(a)
			if err != nil {
				t.Log(err)
			}
			t.Log(meta)
		}

		t.Log("accounts2", strings.Repeat("*", 20))
		for _, a := range accounts2 {
			t.Log(a.String())
		}

		t.Fatal("GetAccounts failed")
	}
}
