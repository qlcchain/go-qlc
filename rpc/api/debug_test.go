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
)

func setupTestCaseDebug(t *testing.T) (func(t *testing.T), *ledger.Ledger, *DebugApi) {
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

func TestDebugApi_UncheckBlock(t *testing.T) {
	teardownTestCase, l, debugApi := setupTestCaseDebug(t)
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
	teardownTestCase, l, debugApi := setupTestCaseDebug(t)
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
