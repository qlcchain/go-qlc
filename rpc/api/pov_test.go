package api

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	rpc "github.com/qlcchain/jsonrpc2"

	qctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

type mockDataTestPovApi struct {
	cfg *config.Config
	eb  event.EventBus
	feb *event.FeedEventBus
	l   *ledger.Ledger
	cc  *qctx.ChainContext
	api *PovApi

	febRpcMsgCh chan *topic.EventRPCSyncCallMsg
	ctx         context.Context
	cancelCtx   context.CancelFunc
}

func setupTestCasePov(t *testing.T) (func(t *testing.T), *mockDataTestPovApi) {
	t.Parallel()

	md := new(mockDataTestPovApi)

	dir := filepath.Join(config.QlcTestDataDir(), "rewards", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()

	md.l = ledger.NewLedger(cm.ConfigFile)

	md.cc = qctx.NewChainContext(cm.ConfigFile)
	md.cc.Init(nil)

	md.cfg, _ = md.cc.Config()

	md.eb = md.cc.EventBus()
	md.feb = md.cc.FeedEventBus()

	md.api = NewPovApi(context.Background(), md.cfg, md.l, md.eb, md.cc)

	md.ctx, md.cancelCtx = context.WithCancel(context.Background())
	md.febRpcMsgCh = make(chan *topic.EventRPCSyncCallMsg, 1)
	md.feb.Subscribe(topic.EventRpcSyncCall, md.febRpcMsgCh)

	go func(md *mockDataTestPovApi) {
		time.Sleep(time.Millisecond)
		for {
			select {
			case <-md.ctx.Done():
				return
			case msg := <-md.febRpcMsgCh:
				t.Log("febRpcMsgCh", msg)
				if msg.Name == "Miner.GetMiningInfo" {
					latestBlock, _ := md.l.GetLatestPovBlock()

					outArgs := msg.Out.(map[interface{}]interface{})
					outArgs["err"] = nil
					outArgs["latestBlock"] = latestBlock
					msg.ResponseChan <- msg.Out
				}
			}
		}
	}(md)

	return func(t *testing.T) {
		md.cancelCtx()
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

func generatePovBlocksToLedger(t *testing.T, md *mockDataTestPovApi) []*types.PovBlock {
	var prevBlk *types.PovBlock
	var allBlks []*types.PovBlock
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
		allBlks = append(allBlks, blk1)

		prevBlk = blk1
	}
	return allBlks
}

func TestPovAPI_GetHeaders(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	_ = md.cc.Start()
	defer func() {
		_ = md.cc.Stop()
	}()
	time.Sleep(10 * time.Millisecond)

	md.eb.Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(10 * time.Millisecond)

	allBlks := generatePovBlocksToLedger(t, md)

	hdr, err := md.api.GetLatestHeader()
	if err != nil {
		t.Fatal(err)
	}
	if hdr == nil {
		t.Fatalf("failed to get latest header")
	}

	hdr, err = md.api.GetFittestHeader(0)
	if err != nil {
		t.Fatal(err)
	}
	if hdr == nil {
		t.Fatalf("failed to get fittest header")
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

	rspHdrs, err := md.api.BatchGetHeadersByHeight(allBlks[0].GetHeight(), uint64(len(allBlks)), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(rspHdrs.Headers) != len(allBlks) {
		t.Fatal("BatchGetHeadersByHeight ascend err", len(rspHdrs.Headers), len(allBlks))
	}

	rspHdrs, err = md.api.BatchGetHeadersByHeight(allBlks[len(allBlks)-1].GetHeight(), uint64(len(allBlks)), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rspHdrs.Headers) != len(allBlks) {
		t.Fatal("BatchGetHeadersByHeight descend err", len(rspHdrs.Headers), len(allBlks))
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

func TestPovAPI_Miner(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	_ = md.cc.Start()
	defer func() {
		_ = md.cc.Stop()
	}()
	time.Sleep(10 * time.Millisecond)

	md.eb.Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(10 * time.Millisecond)

	generatePovBlocksToLedger(t, md)

	mds0 := types.NewPovMinerDayStat()
	mds0.DayIndex = 1
	mds0.MinerStats = make(map[string]*types.PovMinerStatItem)
	mds0.MinerStats["miner1"] = &types.PovMinerStatItem{BlockNum: 10}
	mds0.MinerNum = uint32(len(mds0.MinerStats))
	err := md.l.AddPovMinerStat(mds0)
	if err != nil {
		t.Fatal(err)
	}

	//rspMinerInfo, err := md.api.GetMiningInfo()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//if rspMinerInfo == nil {
	//	t.Fatalf("failed to GetMiningInfo")
	//}

	rspMinerDs, err := md.api.GetMinerDayStat(1)
	if err != nil {
		t.Fatal(err)
	}
	if rspMinerDs == nil {
		t.Fatalf("failed to GetMinerDayStat")
	}

	rspMinerStats, err := md.api.GetMinerStats(nil)
	if err != nil {
		t.Fatal(err)
	}
	if rspMinerStats == nil {
		t.Fatalf("failed to GetMinerStats")
	}
}

func TestPovAPI_PubSub_NewBlock(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	allBlks := generatePovBlocksToLedger(t, md)

	rpcCtx := rpc.SubscriptionContext()

	subBlk, err := md.api.NewBlock(rpcCtx)
	if err != nil {
		t.Fatal(err)
	}

	blk1 := allBlks[0]
	md.api.pubsub.setBlocks(blk1)
	time.Sleep(10 * time.Millisecond)
	md.api.pubsub.fetchBlocks(subBlk.ID)

	md.api.pubsub.removeChan(subBlk.ID)
}
