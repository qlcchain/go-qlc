package api

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	rpc "github.com/qlcchain/jsonrpc2"

	qctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

type mockDataTestPovApi struct {
	cfg *config.Config
	eb  event.EventBus
	l   *ledger.Ledger
	api *PovApi
}

func setupTestCasePov(t *testing.T) (func(t *testing.T), *mockDataTestPovApi) {
	t.Parallel()

	md := new(mockDataTestPovApi)

	dir := filepath.Join(config.QlcTestDataDir(), "rewards", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()

	md.l = ledger.NewLedger(cm.ConfigFile)

	cc := qctx.NewChainContext(cm.ConfigFile)

	md.cfg, _ = cc.Config()

	md.eb = cc.EventBus()

	md.api = NewPovApi(md.cfg, md.l, md.eb, context.Background())

	return func(t *testing.T) {
		_ = md.eb.Close()
		err := md.l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestPovAPI_GetHeaders(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	var prevBlk *types.PovBlock
	for i := 0; i < 3; i++ {
		blk1, td1 := mock.GeneratePovBlock(prevBlk, 0)
		err := md.l.AddPovBlock(blk1, td1)
		if err != nil {
			t.Fatal(err)
		}
		err = md.l.AddPovBestHash(blk1.GetHeight(), blk1.GetHash())
		if err != nil {
			t.Fatal(err)
		}
		err = md.l.SetPovLatestHeight(blk1.GetHeight())
		if err != nil {
			t.Fatal(err)
		}

		prevBlk = blk1
	}

	hdr, err := md.api.GetLatestHeader()
	if err != nil {
		t.Fatal(err)
	}
	if hdr == nil {
		t.Fatalf("failed to get latest header")
	}

	hdr, err = md.api.GetHeaderByHeight(hdr.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
	if hdr == nil {
		t.Fatalf("failed to get header by height")
	}

	hdr, err = md.api.GetHeaderByHash(hdr.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	if hdr == nil {
		t.Fatalf("failed to get header by hash")
	}

	bd, err := md.api.GetLatestBlock(0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if bd == nil {
		t.Fatalf("failed to get latest body")
	}

	bd, err = md.api.GetBlockByHeight(bd.GetHeight(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if bd == nil {
		t.Fatalf("failed to get body by height")
	}

	bd, err = md.api.GetBlockByHash(bd.GetHash(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if bd == nil {
		t.Fatalf("failed to get body by hash")
	}
}

func TestPovAPI_NewBlock(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	_, err := md.api.NewBlock(rpc.SubscriptionContext())
	if err != nil {
		t.Fatal(err)
	}

	blk1, _ := mock.GeneratePovBlock(nil, 0)
	md.api.pubsub.setBlocks(blk1)
}
