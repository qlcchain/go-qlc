package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
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
	}, l, debugApi
}

func setupMockDebugAPI(t *testing.T) (func(t *testing.T), *mocks.Store, *DebugApi) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
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
	}, l, debugApi
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

func TestDebugApi_BlockCacheCount(t *testing.T) {
	teardownTestCase, l, debugApi := setupMockDebugAPI(t)
	defer teardownTestCase(t)

	expect := uint64(10)
	l.On("CountBlocksCache").Return(expect, nil)
	r, err := debugApi.BlockCacheCount()
	if err != nil {
		t.Fatal(err)
	}
	if r["blockCache"] != expect {
		t.Fatal(r)
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
