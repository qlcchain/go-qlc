/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"encoding/hex"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	store *WalletStore
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	cfg, _ := config.DefaultConfig()
	cfg.DataDir = config.QlcTestDataDir()
	dir := cfg.WalletDir()
	t.Log("setup store test case", dir)
	_ = os.RemoveAll(dir)

	store = NewWalletStore(cfg)
	if store == nil {
		t.Fatal("create store failed")
	}
	return func(t *testing.T) {
		t.Log("teardown wallet test case")
		err := store.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestWalletStore_NewSession(t *testing.T) {
	teardownTestCase := setupTestCase(t)
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
	teardownTestCase := setupTestCase(t)
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
	teardownTestCase := setupTestCase(t)
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
	teardownTestCase := setupTestCase(t)
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
	pub, _, err := types.KeypairFromSeed(seed, 2)

	if err != nil {
		t.Fatal(err)
	}

	addr := types.PubToAddress(pub)
	t.Log(addr.String())

	am := mockAccountMeta(addr)
	s := session.ledger.NewLedgerSession(false)
	defer s.Close()

	err = s.AddAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}

	am2, err := s.GetAccountMeta(addr)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(am, am2) {
		t.Fatal("am!=am2")
	}

	if exist := session.IsAccountExist(addr); !exist {
		t.Fatal("IsAccountExist1 failed", addr.String())
	}

	pub2, _, err := types.KeypairFromSeed(seed, 101)

	if err != nil {
		t.Fatal(err)
	}
	addr2 := types.PubToAddress(pub2)

	if exist := session.IsAccountExist(addr2); exist {
		t.Fatal("IsAccountExist2 failed", addr2.String())
	}

	addr3 := mockAddress()
	if exist := session.IsAccountExist(addr3); exist {
		t.Fatal("IsAccountExist3 failed", addr2.String())
	}
}

func mockAccountMeta(addr types.Address) *types.AccountMeta {
	var am types.AccountMeta
	am.Address = addr
	am.Tokens = []*types.TokenMeta{}
	for i := 0; i < 5; i++ {
		s1, _ := random.Intn(math.MaxInt64)
		s2, _ := random.Intn(math.MaxInt64)
		t := types.TokenMeta{
			TokenAccount: mockAddress(),
			Type:         mockHash(),
			BelongTo:     addr,
			Balance:      types.ParseBalanceInts(uint64(s1), uint64(s2)),
			BlockCount:   1,
			OpenBlock:    mockHash(),
			Header:       mockHash(),
			RepBlock:     mockHash(),
			Modified:     time.Now().Unix(),
		}
		am.Tokens = append(am.Tokens, &t)
	}
	return &am
}

func TestSession_GetRawKey(t *testing.T) {
	teardownTestCase := setupTestCase(t)
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
	hash := mockHash()

	sign := acc1.Sign(hash)
	addr := types.PubToAddress(pub)

	acc2, err := session.GetRawKey(addr)
	if err != nil {
		t.Fatal(err)
	}

	if verify := acc2.Address().Verify(hash[:], sign[:]); !verify {
		t.Fatal("verify failed.")
	}

	addr2 := mockAddress()
	if _, err = session.GetRawKey(addr2); err == nil {
		t.Fatal("get invalid raw key failed")
	}
}

func TestSession_SetSeed(t *testing.T) {
	teardownTestCase := setupTestCase(t)
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
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	err = session.SetVersion(2)
	if err != nil {
		t.Fatal(err)
	}

	i, err := session.GetVersion()
	if err != nil {
		t.Fatal(err)
	}
	if i != 2 {
		t.Fatal("set version failed")
	}
}

func TestSession_DeterministicIndex(t *testing.T) {
	teardownTestCase := setupTestCase(t)
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
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	//h:=mockHash()
	addr := mockAddress()

	work, err := session.GetWork(addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(work.String())
}

func TestGenerateWork(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	hash := mockHash()
	work := session.generateWork(hash)
	if !work.IsValid(hash) {
		t.Fatal("generateWork failed =>", hash.String())
	}
	t.Log(hash.String(), "=>", work.String())
}

func TestSession_GetAccounts(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	id, err := store.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	session := store.NewSession(id)
	s := session.ledger.NewLedgerSession(true)
	defer s.Close()

	seed, err := session.GetSeed()
	seedString := hex.EncodeToString(seed)
	//addresses, err := insertAccountMeta(session, seedString, t)

	var addresses []types.Address
	err = s.BatchUpdate(func() error {
		for i := 0; i < 5; i++ {
			pub, _, err := types.KeypairFromSeed(seedString, uint32(i))
			addr := types.PubToAddress(pub)
			am := mockAccountMeta(addr)
			err = s.AddAccountMeta(am)
			if err != nil {
				t.Fatal(err)
			}
			meta, err := s.GetAccountMeta(addr)
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

func mockAddress() types.Address {
	address, _, _ := types.GenerateAddress()

	return address
}

func mockHash() types.Hash {
	b := [types.HashSize]byte{}
	_ = random.Bytes(b[:])
	h := types.Hash{}
	_ = h.UnmarshalBinary(b[:])
	return h
}
