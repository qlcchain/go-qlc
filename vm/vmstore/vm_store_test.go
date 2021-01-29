/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
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
	block.Header.BasHdr.Height = 0
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

func getStorageKey(prefix []byte, key []byte) []byte {
	var storageKey []byte
	storageKey = append(storageKey, prefix...)
	storageKey = append(storageKey, key...)
	return storageKey
}

func TestVMContext_Storage(t *testing.T) {
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
	cacheTrie := ctx.cache.Trie(func(bytes []byte) []byte {
		return ctx.getRawStorageKey(bytes, nil)
	})
	if cacheTrie == nil {
		t.Fatal("invalid trie")
	}

	if hash := cacheTrie.Hash(); hash == nil {
		t.Fatal("invalid hash")
	} else {
		t.Log(hash.String())
	}
}

func TestVMContext_SetObjectStorage(t *testing.T) {
	teardownTestCase, ctx, _ := setupTestCase(t)
	defer teardownTestCase(t)

	blk := mock.StateBlockWithoutWork()
	key := []byte{10, 20, 30}
	if err := ctx.SetObjectStorage(blk.GetHash().Bytes(), key, blk); err != nil {
		t.Fatal(err)
	}
	if _, err := ctx.GetStorageByRaw(key); err == nil {
		t.Fatal()
	}

}

func TestNewVMContext_Trie(t *testing.T) {
	teardownTestCase, ctx, _ := setupTestCase(t)
	defer teardownTestCase(t)

	prefix := mock.Hash()
	value := []byte{10, 20, 30, 40}
	if err := ctx.SetStorage(prefix[:], []byte{10, 20, 30}, value); err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetStorage(prefix[:], []byte{10, 20, 40}, value); err != nil {
		t.Fatal(err)
	}
	if trie := Trie(ctx); trie == nil {
		t.Fatal()
	}
	if r := TrieHash(ctx); r == nil {
		t.Fatal()
	}
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

func TestVMContext_Block(t *testing.T) {
	teardownTestCase, ctx, l := setupTestCase(t)
	defer teardownTestCase(t)

	b1 := mock.StateBlockWithoutWork()
	b2 := mock.StateBlockWithoutWork()
	b2.Previous = b1.GetHash()
	b2.Type = types.Send

	if err := l.AddStateBlock(b1); err != nil {
		t.Fatal(err)
	}
	if err := l.AddStateBlock(b2); err != nil {
		t.Fatal(err)
	}

	if r, err := ctx.GetBlockChild(b1.GetHash()); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
	if r, err := ctx.GetStateBlock(b2.GetHash()); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r.String())
	}
	if r, err := ctx.GetStateBlockConfirmed(b2.GetHash()); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r.String())
	}
	if r, err := ctx.HasStateBlockConfirmed(b2.GetHash()); err != nil || !r {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
	ctx.EventBus().Publish(topic.EventAddRelation, b1)
}

func TestVMContext_Account(t *testing.T) {
	teardownTestCase, ctx, l := setupTestCase(t)
	defer teardownTestCase(t)

	addr := mock.Address()
	acc := mock.AccountMeta(addr)
	token := mock.TokenMeta(addr)
	acc.Tokens = append(acc.Tokens, token)
	if err := l.AddAccountMeta(acc, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	if r, err := ctx.GetAccountMeta(addr); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
	if r, err := ctx.HasAccountMetaConfirmed(addr); err != nil || !r {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
	if r, err := ctx.GetTokenMeta(addr, token.Type); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
}

func TestVMContext_AccountHistory(t *testing.T) {
	teardownTestCase, ctx, l := setupTestCase(t)
	defer teardownTestCase(t)

	addr := mock.Address()
	am := mock.AccountMeta(addr)
	tm1 := mock.TokenMeta(addr)
	tm1.Type = config.ChainToken()
	tm2 := mock.TokenMeta(addr)
	am.Tokens = []*types.TokenMeta{tm1, tm2}

	blk := mock.StateBlockWithoutWork()
	height := uint64(10)
	blk.PoVHeight = height
	blk.Address = addr
	blk.Token = tm1.Type

	tm1.Header = blk.GetHash()
	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal(err)
	}
	if err := l.AddAccountMeta(am, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	if err := l.AddAccountMetaHistory(tm1, blk, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}

	if _, err := ctx.GetAccountMetaByPovHeight(addr); err == nil {
		t.Fatal(err)
	}

	ctx.poVHeight = height
	if r, err := ctx.GetAccountMetaByPovHeight(addr); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}

	if r, err := ctx.GetTokenMetaByPovHeight(addr, tm1.Type); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}

	if r, err := ctx.GetTokenMetaByBlockHash(blk.GetHash()); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
}

func TestVMContext_POV(t *testing.T) {
	teardownTestCase, ctx, _ := setupTestCase(t)
	defer teardownTestCase(t)

	ctx.contractAddr = &contractaddress.NEP5PledgeAddress
	t.Log(ctx.PovGlobalState())
	if r, err := ctx.PoVContractState(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
	if r := ctx.PovGlobalStateByHeight(0); r == nil {
		t.Fatal()
	} else {
		t.Log(r)
	}
	if r, err := ctx.PoVContractStateByHeight(0); err != nil {
		t.Fatal()
	} else {
		t.Log(r)
	}
	if r, err := ctx.GetLatestPovBlock(); err != nil {
		t.Fatal()
	} else {
		t.Log(r)
	}
	if r, err := ctx.GetLatestPovHeader(); err != nil {
		t.Fatal()
	} else {
		t.Log(r)
	}
	if _, err := ctx.GetPovMinerStat(0); err == nil {
		t.Fatal()
	}
}

func TestVMContext_IsUserAccount(t *testing.T) {
	teardownTestCase, ctx, _ := setupTestCase(t)
	defer teardownTestCase(t)

	if b, err := ctx.IsUserAccount(mock.Address()); err == nil || b {
		t.Fatal()
	}
}

func TestVMContext_CalculateAmount(t *testing.T) {
	teardownTestCase, ctx, _ := setupTestCase(t)
	defer teardownTestCase(t)

	blk := mock.StateBlockWithoutWork()
	r, err := ctx.CalculateAmount(blk)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
}

func TestVMContext_Relation(t *testing.T) {
	teardownTestCase, ctx, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk := mock.StateBlockWithoutWork()
	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal(err)
	}
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	var b types.BlockHash
	_ = l.Flush()
	if err := ctx.GetRelation(&b, fmt.Sprintf("select * from blockhash where hash = '%s'", blk.GetHash())); err != nil {
		t.Fatal(err)
	} else {
		t.Log(b)
	}
	var bs []types.BlockHash
	if err := ctx.SelectRelation(&bs, "select * from blockhash"); err != nil {
		t.Fatal(err)
	} else {
		t.Log(bs)
	}
}

func TestVMContext_NewVmContext(t *testing.T) {
	teardownTestCase, ctx, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx = ctx.WithCache(l.Cache().GetCache())

	blk := mock.StateBlockWithoutWork()
	if c := NewVMContextWithBlock(l, blk); c != nil {
		t.Fatal()
	}

	blk.Type = types.ContractSend
	blk.Link = contractaddress.NEP5PledgeAddress.ToHash()
	if c := NewVMContextWithBlock(l, blk); c == nil {
		t.Fatal()
	}
}
