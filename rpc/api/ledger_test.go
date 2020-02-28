package api

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	rpc "github.com/qlcchain/jsonrpc2"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCaseLedger(t *testing.T) (func(t *testing.T), *ledger.Ledger, *LedgerAPI) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "rewards", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cfg, _ := cm.Load()
	for _, v := range cfg.Genesis.GenesisBlocks {
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: v.Mintage,
			GenesisBlock:        v.Genesis,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
	}

	l := ledger.NewLedger(cm.ConfigFile)

	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	eb := cc.EventBus()

	ledgerApi := NewLedgerApi(context.Background(), l, eb, cc)

	return func(t *testing.T) {
		//err := l.DBStore.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l, ledgerApi
}

func TestLedger_GetBlockCacheLock(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupTestCaseLedger(t)
	defer teardownTestCase(t)

	ledgerApi.ledger.EB.Publish(topic.EventPovSyncState, topic.SyncDone)
	chainToken := common.ChainToken()
	gasToken := common.GasToken()
	addr, _ := types.HexToAddress("qlc_361j3uiqdkjrzirttrpu9pn7eeussymty4rz4gifs9ijdx1p46xnpu3je7sy")
	_ = ledgerApi.getProcessLock(addr, chainToken)
	if ledgerApi.processLock.Len() != 1 {
		t.Fatal("lock len error for addr")
	}
	_ = ledgerApi.getProcessLock(addr, gasToken)
	if ledgerApi.processLock.Len() != 2 {
		t.Fatal("lock error for different token")
	}

	for i := 0; i < 998; i++ {
		a := mock.Address()
		ledgerApi.getProcessLock(a, chainToken)
	}
	if ledgerApi.processLock.Len() != 1000 {
		t.Fatal("lock len error for 1000 addresses")
	}
	sb := mock.StateBlockWithAddress(addr)
	_, _ = ledgerApi.Process(sb)
	addr2, _ := types.HexToAddress("qlc_1gnggt8b6cwro3b4z9gootipykqd6x5gucfd7exsi4xqkryiijciegfhon4u")
	_ = ledgerApi.getProcessLock(addr2, chainToken)
	if ledgerApi.processLock.Len() != 1001 {
		t.Fatal("get error when delete idle lock")
	}
}

func TestLedgerApi_Subscription(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupTestCaseLedger(t)
	defer teardownTestCase(t)

	addr := mock.Address()
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				blk := mock.StateBlock()
				blk.Address = addr
				blk.Type = types.Send
				ledgerApi.ledger.EB.Publish(topic.EventAddRelation, blk)
				time.Sleep(2 * time.Second)
			}
		}
	}()
	ctx := rpc.SubscriptionContext()
	r, err := ledgerApi.NewBlock(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)

	ctx2 := rpc.SubscriptionContext()
	r, err = ledgerApi.NewBlock(ctx2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)

	ctx3 := rpc.SubscriptionContext()
	r, err = ledgerApi.BalanceChange(ctx3, addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)

	ctx4 := rpc.SubscriptionContext()
	r, err = ledgerApi.BalanceChange(ctx4, addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
	done <- true
}
