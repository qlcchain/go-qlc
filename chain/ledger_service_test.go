/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/trie"
)

func TestNewLedgerService(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	ls := NewLedgerService(cm.ConfigFile)

	err = ls.Init()
	if err != nil {
		t.Fatal(err)
	}
	if ls.State() != 2 {
		t.Fatal("ledger init failed")
	}
	_ = ls.Start()
	err = ls.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if ls.Status() != 6 {
		t.Fatal("stop failed.")
	}

	ls2 := NewLedgerService(cm.ConfigFile)
	err = ls2.Init()
	if err != nil {
		t.Fatal(err)
	}
	_ = ls2.Start()
	err = ls2.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedgerService_RemoveUselessTrie(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	cfg, err := cm.Load()
	cfg.DBOptimize.SyncWriteHeight = 0
	cfg.DBOptimize.Enable = true
	cm.Save()
	if err != nil {
		t.Fatal(err)
	}
	ls := NewLedgerService(cm.ConfigFile)
	l := ls.Ledger
	defer func() {
		l.Close()
		_ = os.RemoveAll(dir)
	}()

	if err := ls.removeUselessTrie(); err != nil {
		t.Fatal()
	}

	for i := 0; i < 4; i++ {
		blk := mock.StateBlock()
		blk.Type = types.Send
		blk.Previous = types.ZeroHash
		if err := ls.Ledger.AddStateBlock(blk); err != nil {
			t.Fatal(err)
		}
	}

	mockAccountTrie(ls.Ledger, t)

	if err := ls.backupAccountTrieData(); err != nil {
		t.Fatal(err)
	}
	trCount, _ := ls.Ledger.DBStore().Count([]byte{byte(storage.KeyPrefixTrie)})
	trTempCount, _ := ls.Ledger.DBStore().Count([]byte{byte(storage.KeyPrefixTrieTemp)})
	if trCount == 0 || trCount != trTempCount {
		t.Fatal(trCount, trTempCount)
	}

	tr2 := trie.NewTrie(ls.Ledger.DBStore(), nil, trie.NewSimpleTrieNodePool())
	key1 := []byte("PIamGood")
	key2 := []byte("Ptesa")
	value1 := []byte("value.555val")
	value2 := []byte("asdfvale....asdvalue.asdvalue.555val")

	tr2.SetValue(key1, value1)
	tr2.SetValue(key2, value2)
	fn2, err := tr2.Save()
	if err != nil {
		t.Fatal(err)
	}
	fn2()
	root2 := tr2.Hash()

	if err := ls.backupPovTrieData(); err == nil {
		t.Fatal()
	}
	block, td := mock.GeneratePovBlock(nil, 0)
	block.Header.CbTx.StateHash = *root2
	if err := l.AddPovBlock(block, td); err != nil {
		t.Fatal(err)
	}
	if err := l.AddPovBestHash(block.GetHeight(), block.GetHash()); err != nil {
		t.Fatal(err)
	}
	if err := l.SetPovLatestHeight(block.GetHeight()); err != nil {
		t.Fatal(err)
	}

	if err := ls.backupPovTrieData(); err != nil {
		t.Fatal(err)
	}

	trCount, _ = ls.Ledger.DBStore().Count([]byte{byte(storage.KeyPrefixTrie)})
	trTempCount, _ = ls.Ledger.DBStore().Count([]byte{byte(storage.KeyPrefixTrieTemp)})
	if trCount == 0 || trCount != trTempCount {
		t.Fatal(trCount, trTempCount)
	}
	t.Log(trCount)

	tr3 := trie.NewTrie(ls.Ledger.DBStore(), nil, trie.NewSimpleTrieNodePool()) // add redundant data
	key1 = []byte("RIamGood")
	value1 = []byte("Rvalue.555val")

	tr3.SetValue(key1, value1)
	tr2.SetValue(key2, value2)
	fn3, err := tr3.Save()
	if err != nil {
		t.Fatal(err)
	}
	fn3()

	if err := ls.removeUselessTrie(); err != nil {
		t.Fatal(err)
	}
	if trTempCount, err = ls.Ledger.DBStore().Count([]byte{byte(storage.KeyPrefixTrieTemp)}); trTempCount != 0 || err != nil {
		t.Fatal(trTempCount, err)
	}
}

func TestLedgerService_CleanTrie(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	cfg, err := cm.Load()
	cfg.DBOptimize.SyncWriteHeight = 0
	cfg.DBOptimize.Enable = true
	cm.Save()
	if err != nil {
		t.Fatal(err)
	}
	ls := NewLedgerService(cm.ConfigFile)
	l := ls.Ledger
	defer func() {
		l.Close()
		_ = os.RemoveAll(dir)
	}()

	mockAccountTrie(ls.Ledger, t)

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
	trCount, _ := ls.Ledger.DBStore().Count([]byte{byte(storage.KeyPrefixTrie)})
	t.Log(trCount)
	ls.Ledger.AddTrieCleanHeight(0)
	duration := time.Duration(1) * time.Second
	go ls.clean(duration, 0)
	time.Sleep(2 * time.Second)
	trCount, _ = ls.Ledger.DBStore().Count([]byte{byte(storage.KeyPrefixTrie)})
	t.Log(trCount)
	if trCount != 0 {
		t.Fatal(trCount)
	}
	ls.cancel()
	time.Sleep(200 * time.Millisecond)
}

func mockAccountTrie(l *ledger.Ledger, t *testing.T) {
	tr := trie.NewTrie(l.DBStore(), nil, trie.NewSimpleTrieNodePool())
	key1 := []byte("IamGood")
	key2 := []byte("tesab")
	key3 := []byte("tesa")

	value1 := []byte("a1230xm90zm19ma")
	value2 := []byte("value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555value.555")
	value3 := []byte("value.555val")

	tr.SetValue(key1, value1)
	tr.SetValue(key2, value2)
	tr.SetValue(key3, value3)
	fn, err := tr.Save()
	if err != nil {
		t.Fatal(err)
	}
	fn()
	root := tr.Hash()

	blk := mock.StateBlock()
	blk.Type = types.ContractSend
	blk.Previous = types.ZeroHash
	blk.Extra = root

	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal(err)
	}

	account := mock.AccountMeta(blk.Address)
	token := mock.TokenMeta(blk.Address)
	token.Type = blk.Token
	token.Header = blk.GetHash()
	account.Tokens = []*types.TokenMeta{token}
	if err := l.AddAccountMeta(account, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
}

func TestLedgerService_Monitor(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	ls := NewLedgerService(cm.ConfigFile)
	duration := time.Duration(1) * time.Second
	go ls.ledgerMonitor(duration)
	time.Sleep(2 * time.Second)
	ls.cancel()
	time.Sleep(200 * time.Millisecond)
}
