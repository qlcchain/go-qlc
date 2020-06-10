package apis

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	qctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

type mockDataTestPovApi struct {
	cfg *config.Config
	eb  event.EventBus
	feb *event.FeedEventBus
	l   *ledger.Ledger
	cc  *qctx.ChainContext
	api *PovAPI

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

	md.api = NewPovAPI(context.Background(), md.cfg, md.l, md.eb, md.cc)

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
				if msg.Name == "Miner.GetMiningInfo" {
					latestBlock, _ := md.l.GetLatestPovBlock()

					outArgs := msg.Out.(map[interface{}]interface{})
					outArgs["err"] = nil
					outArgs["latestBlock"] = latestBlock
					outArgs["syncState"] = int(topic.SyncDone)
					outArgs["pooledTx"] = uint32(0)

					outArgs["minerAddr"] = types.ZeroAddress
					outArgs["minerAlgo"] = types.ALGO_UNKNOWN
					outArgs["cpuMining"] = false

					msg.ResponseChan <- msg.Out
					t.Log("febRpcMsgCh", "in", msg.In, "out", msg.Out)
				} else if msg.Name == "Miner.GetWork" {
					mineBlk := types.NewPovMineBlock()

					outArgs := msg.Out.(map[interface{}]interface{})
					outArgs["err"] = nil
					outArgs["mineBlock"] = mineBlk

					msg.ResponseChan <- msg.Out
					t.Log("febRpcMsgCh", "in", msg.In, "out", msg.Out)
				} else if msg.Name == "Miner.SubmitWork" {
					outArgs := msg.Out.(map[interface{}]interface{})
					outArgs["err"] = nil

					msg.ResponseChan <- msg.Out
					t.Log("febRpcMsgCh", "in", msg.In, "out", msg.Out)
				} else if msg.Name == "Miner.StartMining" || msg.Name == "Miner.StopMining" {
					outArgs := msg.Out.(map[interface{}]interface{})
					outArgs["err"] = nil

					msg.ResponseChan <- msg.Out
					t.Log("febRpcMsgCh", "in", msg.In, "out", msg.Out)
				} else {
					t.Log("febRpcMsgCh", "in", msg.In)
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
