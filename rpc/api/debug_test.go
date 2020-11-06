package api

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/mock/mocks"
)

func setupDefaultDebugAPI(t *testing.T) (func(t *testing.T), *ledger.Ledger, *DebugApi) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "debug", uuid.New().String())
	_ = os.RemoveAll(dir)
	fmt.Println("start: ", t.Name())
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()

	l := ledger.NewLedger(cm.ConfigFile)
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	eb := cc.EventBus()
	debugApi := NewDebugApi(cm.ConfigFile, eb)

	return func(t *testing.T) {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("end: ", t.Name())
	}, l, debugApi
}

func setupMockDebugAPI(t *testing.T) (func(t *testing.T), *mocks.Store, *DebugApi) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	fmt.Println("start: ", t.Name())
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)

	l := new(mocks.Store)
	debugApi := NewDebugApi(cm.ConfigFile, cc.EventBus())
	debugApi.ledger = l
	return func(t *testing.T) {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
		fmt.Println("end: ", t.Name())
	}, l, debugApi
}

func TestDebugApi_BlockCaches(t *testing.T) {
	teardownTestCase, l, debugApi := setupDefaultDebugAPI(t)
	defer teardownTestCase(t)

	if err := l.AddBlockCache(mock.StateBlockWithoutWork()); err != nil {
		t.Fatal(err)
	}
	if err := l.AddBlockCache(mock.StateBlockWithoutWork()); err != nil {
		t.Fatal(err)
	}
	if err := l.AddBlockCache(mock.StateBlockWithoutWork()); err != nil {
		t.Fatal(err)
	}
	if r, err := debugApi.BlockCacheCount(); r["blockCache"] != 3 || err != nil {
		t.Fatal(err, r)
	}
	if r, err := debugApi.BlockCaches(); len(r) != 3 || err != nil {
		t.Fatal(err, r)
	}
}

func TestDebugApi_Action(t *testing.T) {
	teardownTestCase, l, debugApi := setupMockDebugAPI(t)
	defer teardownTestCase(t)

	l.On("Action", storage.Dump, 0).Return("done", nil)
	r, err := debugApi.Action(storage.Dump, 0)
	if err != nil {
		t.Fatal(err)
	}
	if r != "done" {
		t.Fatal(r)
	}
}

func TestDebugApi_BlockLink(t *testing.T) {
	teardownTestCase, l, debugApi := setupDefaultDebugAPI(t)
	defer teardownTestCase(t)

	// BlockLink
	blk := mock.StateBlockWithoutWork()
	blkKey1, _ := storage.GetKeyOfParts(storage.KeyPrefixLink, blk.GetHash())
	link := mock.Hash()
	linkVal, _ := link.Serialize()
	if err := l.DBStore().Put(blkKey1, linkVal); err != nil {
		t.Fatal(err)
	}
	blkKey2, _ := storage.GetKeyOfParts(storage.KeyPrefixChild, blk.GetHash())
	child := mock.Hash()
	childVal, _ := child.Serialize()
	if err := l.DBStore().Put(blkKey2, childVal); err != nil {
		t.Fatal(err)
	}

	r, err := debugApi.BlockLink(blk.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)

	// BlockLinks

	pre := mock.Hash()
	blk.Previous = pre
	blkKey, _ := storage.GetKeyOfParts(storage.KeyPrefixBlock, blk.GetHash())
	blkVal, _ := blk.Serialize()
	if err := l.DBStore().Put(blkKey, blkVal); err != nil {
		t.Fatal(err)
	}

	r2, err := debugApi.BlockLinks(pre)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r2)

}

func TestDebugApi_BlocksCountByType(t *testing.T) {
	teardownTestCase, l, debugApi := setupDefaultDebugAPI(t)
	defer teardownTestCase(t)

	if err := l.AddStateBlock(mock.StateBlockWithoutWork()); err != nil {
		t.Fatal(err)
	}
	if err := l.AddStateBlock(mock.StateBlockWithoutWork()); err != nil {
		t.Fatal(err)
	}
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	r, err := debugApi.BlocksCountByType("address")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
	r, err = debugApi.BlocksCountByType("type")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
	r, err = debugApi.BlocksCountByType("token")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestDebugApi_GetSyncBlockNum(t *testing.T) {
	teardownTestCase, l, debugApi := setupMockDebugAPI(t)
	defer teardownTestCase(t)

	l.On("CountUncheckedSyncBlocks").Return(uint64(10), nil)
	l.On("CountUnconfirmedSyncBlocks").Return(uint64(20), nil)
	r, err := debugApi.GetSyncBlockNum()
	if err != nil {
		t.Fatal(err)
	}
	if r["uncheckedSync"] != uint64(10) || r["unconfirmedSync"] != uint64(20) {
		t.Fatal(r)
	}
}

func TestDebugApi_Representative(t *testing.T) {
	teardownTestCase, l, debugApi := setupDefaultDebugAPI(t)
	defer teardownTestCase(t)

	am := mock.AccountMeta(mock.Address())
	am.Tokens[0].Type = config.ChainToken()
	am.Tokens[0].Representative = am.Address
	key, _ := storage.GetKeyOfParts(storage.KeyPrefixAccount, am.Address)
	val, _ := am.Serialize()
	if err := l.DBStore().Put(key, val); err != nil {
		t.Fatal(err)
	}
	r, err := debugApi.Representative(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestDebugApi_Pending(t *testing.T) {
	teardownTestCase, l, debugApi := setupDefaultDebugAPI(t)
	defer teardownTestCase(t)
	setLedgerStatus(l, t)
	blk := mock.StateBlockWithoutWork()
	pk := &types.PendingKey{
		Address: mock.Address(),
		Hash:    blk.GetHash(),
	}
	pi := &types.PendingInfo{
		Source: mock.Address(),
		Type:   config.ChainToken(),
	}
	pkey, _ := storage.GetKeyOfParts(storage.KeyPrefixPending, pk)
	pVal, _ := pi.Serialize()
	if err := l.DBStore().Put(pkey, pVal); err != nil {
		t.Fatal(err)
	}
	bKey, _ := storage.GetKeyOfParts(storage.KeyPrefixBlock, blk.GetHash())
	bVal, _ := blk.Serialize()
	if err := l.DBStore().Put(bKey, bVal); err != nil {
		t.Fatal(err)
	}
	r, err := debugApi.AccountPending(pk.Address, pk.Hash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)

	// get pendings
	pk2 := &types.PendingKey{
		Address: pk.Address,
		Hash:    blk.GetHash(),
	}
	pi2 := &types.PendingInfo{
		Source: mock.Address(),
		Type:   config.ChainToken(),
	}
	pkey2, _ := storage.GetKeyOfParts(storage.KeyPrefixPending, pk2)
	pVal2, _ := pi2.Serialize()
	if err := l.DBStore().Put(pkey2, pVal2); err != nil {
		t.Fatal(err)
	}

	pa, err := debugApi.PendingsAmount()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pa)

	pc, err := debugApi.PendingsCount()
	if err != nil {
		t.Fatal(err)
	}
	if pc != 1 {
		t.Fatal(pc)
	}
}

func TestDebugApi_ContractCount(t *testing.T) {
	teardownTestCase, l, debugApi := setupDefaultDebugAPI(t)
	defer teardownTestCase(t)
	setLedgerStatus(l, t)
	r, err := debugApi.ContractCount()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestDebugApi_GetCache(t *testing.T) {
	teardownTestCase, l, debugApi := setupDefaultDebugAPI(t)
	defer teardownTestCase(t)

	c := l.Cache().GetCache()
	if err := c.Put([]byte{1, 2, 3}, []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if err := c.Put([]byte{1, 2, 4}, []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1200 * time.Millisecond)
	if err := debugApi.GetCache(); err != nil {
		t.Fatal(err)
	}
	r := debugApi.GetCacheStat()
	t.Log(r)
	s := debugApi.GetCacheStatus()
	t.Log(s)
}

func TestDebugApi_UncheckBlock(t *testing.T) {
	teardownTestCase, l, debugApi := setupDefaultDebugAPI(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()
	gh1 := blk2.GetHash()
	err := l.AddUncheckedBlock(gh1, blk1, types.UncheckedKindPrevious, types.InvalidSynchronized)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddStateBlock(blk2)
	if err != nil {
		t.Fatal(err)
	}

	uis, err := debugApi.UncheckBlock(blk1.GetHash())
	if err != nil || len(uis) == 0 {
		t.Fatal(err)
	}

	if uis[0].Hash != blk1.GetHash() {
		t.Fatal()
	}
}

func TestDebugApi_UncheckAnalysis(t *testing.T) {
	teardownTestCase, l, debugApi := setupDefaultDebugAPI(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()
	gh1 := blk2.GetHash()
	err := l.AddUncheckedBlock(gh1, blk1, types.UncheckedKindLink, types.InvalidSynchronized)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddStateBlock(blk2)
	if err != nil {
		t.Fatal(err)
	}

	blk3 := mock.StateBlockWithoutWork()
	err = l.AddGapPovBlock(100, blk3, types.InvalidSynchronized)
	if err != nil {
		t.Fatal(err)
	}

	uis, err := debugApi.UncheckAnalysis()
	if err != nil || len(uis) != 2 {
		t.Fatal(err)
	}

	for _, ui := range uis {
		if ui.GapType == "gap link" && ui.Hash != blk1.GetHash() {
			t.Fatal()
		}

		if ui.GapType == "gap pov" && ui.Hash != blk3.GetHash() {
			t.Fatal()
		}
	}
}

func TestDebugApi_UncheckBlocks(t *testing.T) {
	teardownTestCase, l, debugApi := setupDefaultDebugAPI(t)
	defer teardownTestCase(t)

	block := mock.StateBlockWithoutWork()
	hash := block.GetLink()
	kind := types.UncheckedKindLink
	if err := l.AddUncheckedBlock(hash, block, kind, types.UnSynchronized); err != nil {
		t.Fatal(err)
	}

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()

	err := l.AddGapPovBlock(10, blk1, types.Synchronized)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddGapPovBlock(100, blk2, types.UnSynchronized)
	if err != nil {
		t.Fatal(err)
	}

	if r, err := debugApi.UncheckBlocksCount(); err != nil || r["Total"] != 3 {
		t.Fatal(err)
	}
	if r, err := debugApi.UncheckBlocks(); err != nil || len(r) != 3 {
		t.Fatal(err)
	}
	r, err := debugApi.UncheckBlock(block.GetHash())
	t.Log(r, err)
	if _, err = debugApi.UncheckAnalysis(); err != nil {
		t.Fatal(err)
	}
}
