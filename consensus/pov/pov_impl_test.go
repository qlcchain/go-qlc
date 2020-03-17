package pov

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

type povImplMockData struct {
	cfg *config.Config
	cm  *config.CfgManager

	eb     event.EventBus
	ledger ledger.Store
}

func setupPovImplTestCase(t *testing.T) (func(t *testing.T), *povImplMockData) {
	t.Parallel()

	md := &povImplMockData{}

	uid := uuid.New().String()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uid)
	_ = os.RemoveAll(dir)
	md.cm = config.NewCfgManager(dir)
	_, _ = md.cm.Load()

	md.ledger = ledger.NewLedger(md.cm.ConfigFile)

	md.cfg, _ = md.cm.Config()

	md.eb = event.GetEventBus(uid)

	return func(t *testing.T) {
		err := md.ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}

		err = md.eb.Close()
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestPovImpl_StartStop(t *testing.T) {
	teardownTestCase, md := setupPovImplTestCase(t)
	defer teardownTestCase(t)

	povImpl, err := NewPovEngine(md.cm.ConfigFile, false)
	if err != nil {
		t.Fatal(err)
	}

	err = povImpl.Init()
	if err != nil {
		t.Fatal(err)
	}

	err = povImpl.Start()
	if err != nil {
		t.Fatal(err)
	}

	blk1, _ := mock.GeneratePovBlock(nil, 0)
	err = povImpl.onRecvPovBlock(blk1, types.PovBlockFromRemoteBroadcast, "testPeer1")
	if err != nil {
		t.Fatal(err)
	}

	blk2, _ := mock.GeneratePovBlock(blk1, 0)
	err = povImpl.AddMinedBlock(blk2)
	if err != nil {
		t.Fatal(err)
	}

	msg := &topic.EventRPCSyncCallMsg{
		Name:         "Debug.PovInfo",
		In:           make(map[string]interface{}),
		Out:          make(map[string]interface{}),
		ResponseChan: make(chan interface{}, 1),
	}
	povImpl.onEventRPCSyncCall(msg)
	select {
	case <-msg.ResponseChan:
	case <-time.After(10 * time.Millisecond):
		t.Fatal("ResponseChan timeout")
	}

	err = povImpl.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPovImpl_ManyBlocks(t *testing.T) {
	teardownTestCase, md := setupPovImplTestCase(t)
	defer teardownTestCase(t)

	povImpl, err := NewPovEngine(md.cm.ConfigFile, true)
	if err != nil {
		t.Fatal(err)
	}

	err = povImpl.Init()
	if err != nil {
		t.Fatal(err)
	}

	err = povImpl.Start()
	if err != nil {
		t.Fatal(err)
	}

	var prevBlk *types.PovBlock
	for cnt := 0; cnt < common.POVChainBlocksPerDay*2; cnt++ {
		blk, _ := mock.GeneratePovBlockByFakePow(prevBlk, 0)
		blk.Header.BasHdr.Nonce = blk.Header.BasHdr.Timestamp
		err = povImpl.AddMinedBlock(blk)
		if err != nil {
			t.Fatal(err)
		}
		prevBlk = blk
	}

	err = povImpl.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
