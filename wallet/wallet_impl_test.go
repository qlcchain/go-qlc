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
	"github.com/qlcchain/go-qlc/common/storage"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *WalletStore) {
	t.Parallel()
	start := time.Now()

	dir := filepath.Join(config.QlcTestDataDir(), "wallet", uuid.New().String())
	t.Log("setup store test case", dir)
	_ = os.RemoveAll(dir)

	cm := config.NewCfgManager(dir)
	cm.Load()
	store := NewWalletStore(cm.ConfigFile)
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

	am := mock.AccountMeta(addr)
	l := session.ledger

	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	am2, err := l.GetAccountMeta(addr)
	if err != nil {
		t.Fatal(err)
	}
	if util.ToString(am) != util.ToString(am2) {
		t.Log(util.ToIndentString(am))
		t.Log(util.ToIndentString(am2))
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

	addr3 := mock.Address()
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
	hash := mock.Hash()

	sign := acc1.Sign(hash)
	addr := types.PubToAddress(pub)

	acc2, err := session.GetRawKey(addr)
	if err != nil {
		t.Fatal(err)
	}

	if verify := acc2.Address().Verify(hash[:], sign[:]); !verify {
		t.Fatal("verify failed.")
	}

	addr2 := mock.Address()
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
	err = session.setSeed(seed[:])
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
	addr := mock.Address()

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
	session := store.NewSession(id)
	go func() {
		hash := mock.Hash()
		work := session.generateWork(hash)
		if !work.IsValid(hash) {
			done <- ""
		}
		done <- fmt.Sprintf("hash[%s]=>%s", hash.String(), work.String())
	}()
	r := <-done
	if r == "" {
		t.Fatal("failed to generate work...")
	} else {
		t.Log(r)
	}
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
	for i := 0; i < 5; i++ {
		account, _ := ss.Account(uint32(i))
		addr := account.Address()
		am := mock.AccountMeta(addr)
		err = s.AddAccountMeta(am, s.Cache().GetCache())
		if err != nil {
			t.Fatal(err)
		}
		meta, err := s.GetAccountMeta(addr)
		if err != nil {
			t.Fatal(err)
		}
		if util.ToString(am) != util.ToString(meta) {
			t.Log(am, meta)
			t.Fatal("save am failed")
		}
		addresses = append(addresses, addr)
		t.Log(addr.String())
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

func TestSession_SetDeterministicIndex(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)
	type fields struct {
		Store storage.Store
		//lock            sync.RWMutex
		ledger          *ledger.Ledger
		logger          *zap.SugaredLogger
		maxAccountCount uint64
		walletId        []byte
		password        *crypto.SecureString
	}
	type args struct {
		index int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"SetDeterministicIndex", fields{Store: store, walletId: mock.Address().Bytes()}, args{3}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Store:           tt.fields.Store,
				lock:            sync.RWMutex{},
				ledger:          tt.fields.ledger,
				logger:          tt.fields.logger,
				maxAccountCount: tt.fields.maxAccountCount,
				walletId:        tt.fields.walletId,
				password:        tt.fields.password,
			}
			if err := s.SetDeterministicIndex(tt.args.index); (err != nil) != tt.wantErr {
				t.Errorf("Session.SetDeterministicIndex() error = %v, wantErr %v", err, tt.wantErr)
				if i, err := s.GetDeterministicIndex(); err == nil && i == tt.args.index {

				} else {
					t.Error("fail to set DeterministicIndex")
				}
			}
		})
	}
}

func TestSession_generateWork(t *testing.T) {
	t.Parallel()
	work := types.Work(0x880ab6aa90a59d5d)
	var hash types.Hash
	_ = hash.Of("2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E")
	type fields struct {
		Store           storage.Store
		ledger          *ledger.Ledger
		logger          *zap.SugaredLogger
		maxAccountCount uint64
		walletId        []byte
		password        *crypto.SecureString
	}
	type args struct {
		hash types.Hash
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   types.Work
	}{
		{"generateWork", fields{}, args{hash: hash}, work},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Store:           tt.fields.Store,
				lock:            sync.RWMutex{},
				ledger:          tt.fields.ledger,
				logger:          tt.fields.logger,
				maxAccountCount: tt.fields.maxAccountCount,
				walletId:        tt.fields.walletId,
				password:        tt.fields.password,
			}
			if got := s.generateWork(tt.args.hash); !tt.want.IsValid(hash) {
				t.Errorf("Session.generateWork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_setWork(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	type fields struct {
		Store           storage.Store
		ledger          *ledger.Ledger
		logger          *zap.SugaredLogger
		maxAccountCount uint64
		walletId        []byte
		password        *crypto.SecureString
	}
	type args struct {
		account types.Address
		work    types.Work
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"setWork", fields{Store: store}, args{account: mock.Address(), work: types.Work(1111)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Store:           tt.fields.Store,
				lock:            sync.RWMutex{},
				ledger:          tt.fields.ledger,
				logger:          tt.fields.logger,
				maxAccountCount: tt.fields.maxAccountCount,
				walletId:        tt.fields.walletId,
				password:        tt.fields.password,
			}
			if err := s.setWork(tt.args.account, tt.args.work); (err != nil) != tt.wantErr {
				t.Errorf("Session.setWork() error = %v, wantErr %v", err, tt.wantErr)
				_, err := s.GetWork(tt.args.account)
				if err != nil {
					t.Errorf("Session.GetWork() error = %v", err)
				}
			}
		})
	}
}

func TestSession_getKey(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)
	type fields struct {
		Store storage.Store
		//lock            sync.RWMutex
		ledger          *ledger.Ledger
		logger          *zap.SugaredLogger
		maxAccountCount uint64
		walletId        []byte
		password        *crypto.SecureString
	}
	type args struct {
		t byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{"getKey", fields{Store: store, walletId: []byte{2, 3, 4}}, args{byte(1)}, []byte{1, 2, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Store:           tt.fields.Store,
				lock:            sync.RWMutex{},
				ledger:          tt.fields.ledger,
				logger:          tt.fields.logger,
				maxAccountCount: tt.fields.maxAccountCount,
				walletId:        tt.fields.walletId,
				password:        tt.fields.password,
			}
			if got := s.getKey(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.getKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_getPassword(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)
	pwd, _ := crypto.NewSecureString("PRxWPHK4WXmaHrW5hr9m")

	type fields struct {
		Store storage.Store
		//lock            sync.RWMutex
		ledger          *ledger.Ledger
		logger          *zap.SugaredLogger
		maxAccountCount uint64
		walletId        []byte
		password        *crypto.SecureString
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{"getPassword", fields{Store: store, walletId: mock.Address().Bytes(), password: pwd}, pwd.Bytes()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Store:           tt.fields.Store,
				lock:            sync.RWMutex{},
				ledger:          tt.fields.ledger,
				logger:          tt.fields.logger,
				maxAccountCount: tt.fields.maxAccountCount,
				walletId:        tt.fields.walletId,
				password:        tt.fields.password,
			}
			if got := s.getPassword(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.getPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_setPassword(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)
	type fields struct {
		Store storage.Store
		//lock            sync.RWMutex
		ledger          *ledger.Ledger
		logger          *zap.SugaredLogger
		maxAccountCount uint64
		walletId        []byte
		password        *crypto.SecureString
	}
	type args struct {
		password string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"setPassword", fields{Store: store, walletId: mock.Address().Bytes()}, args{"v3FGe68mFYGewjQ3zjb9"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Store:           tt.fields.Store,
				lock:            sync.RWMutex{},
				ledger:          tt.fields.ledger,
				logger:          tt.fields.logger,
				maxAccountCount: tt.fields.maxAccountCount,
				walletId:        tt.fields.walletId,
				password:        tt.fields.password,
			}
			s.setPassword(tt.args.password)
			if string(s.getPassword()) != tt.args.password {
				t.Fatal("invalid password")
			}
		})
	}
}

func Test_max(t *testing.T) {
	t.Parallel()
	type args struct {
		x uint32
		y uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{"max1", args{x: uint32(1), y: uint32(2)}, uint32(2)},
		{"max2", args{x: uint32(3), y: uint32(1)}, uint32(3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := max(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_min(t *testing.T) {
	t.Parallel()
	type args struct {
		a uint32
		b uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{"min1", args{a: uint32(1), b: uint32(2)}, uint32(1)},
		{"min2", args{a: uint32(3), b: uint32(1)}, uint32(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := min(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("min() = %v, want %v", got, tt.want)
			}
		})
	}
}
