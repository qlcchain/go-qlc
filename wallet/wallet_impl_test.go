/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/storage"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto"
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

func TestSession_SetDeterministicIndex(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)
	type fields struct {
		Store storage.Store
		//lock            sync.RWMutex
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

	secureString, _ := crypto.NewSecureString("password")

	type fields struct {
		Store           storage.Store
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
		{
			name: "setPassword",
			fields: fields{
				Store:    store,
				walletId: mock.Address().Bytes(),
			},
			args: args{"v3FGe68mFYGewjQ3zjb9"},
		},
		{
			name: "setPassword",
			fields: fields{
				Store:    store,
				walletId: mock.Address().Bytes(),
				password: secureString,
			},
			args: args{"v3FGe68mFYGewjQ3zjb9"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Store:           tt.fields.Store,
				lock:            sync.RWMutex{},
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

func TestSession_GetWalletId(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	a := mock.Address()
	s := store.NewSession(a)
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	if id, err := s.GetWalletId(); err != nil {
		t.Fatal(err)
	} else {
		if !bytes.Equal(id, a[:]) {
			t.Fatalf("exp: %v, act: %v", a[:], id)
		}
	}
}

func TestSession_GetRepresentative(t *testing.T) {
	teardownTestCase, store := setupTestCase(t)
	defer teardownTestCase(t)

	a := mock.Address()
	s := store.NewSession(a)
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	rep := mock.Address()
	if err := s.SetRepresentative(rep); err != nil {
		t.Fatal(err)
	}

	if rep2, err := s.GetRepresentative(); err != nil {
		t.Fatal(err)
	} else {
		if rep != rep2 {
			t.Fatalf("exp: %s, act: %s", rep, rep2)
		}
	}
}
