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
	_, _ = cm.Load()
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
	chainToken := config.ChainToken()
	gasToken := config.GasToken()
	addr, _ := types.HexToAddress("qlc_361j3uiqdkjrzirttrpu9pn7eeussymty4rz4gifs9ijdx1p46xnpu3je7sy")
	_ = ledgerApi.getProcessLock(addr, chainToken)
	if ledgerApi.processLockLen() != 1 {
		t.Fatal("lock len error for addr")
	}
	_ = ledgerApi.getProcessLock(addr, gasToken)
	if ledgerApi.processLockLen() != 2 {
		t.Fatal("lock error for different token")
	}

	for i := 0; i < 998; i++ {
		a := mock.Address()
		ledgerApi.getProcessLock(a, chainToken)
	}
	if ledgerApi.processLockLen() != 1000 {
		t.Fatal("lock len error for 1000 addresses")
	}
	sb := mock.StateBlockWithAddress(addr)
	_, _ = ledgerApi.Process(sb)
	addr2, _ := types.HexToAddress("qlc_1gnggt8b6cwro3b4z9gootipykqd6x5gucfd7exsi4xqkryiijciegfhon4u")
	_ = ledgerApi.getProcessLock(addr2, chainToken)
	if ledgerApi.processLockLen() != 1001 {
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

func TestLedgerAPI_GenesisInfo(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupTestCaseLedger(t)
	defer teardownTestCase(t)
	genesisAddress, _ := types.HexToAddress("qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44")
	gasAddress, _ := types.HexToAddress("qlc_3qe19joxq85rnff5wj5ybp6djqtheqqetfgqc3iogxagnjq4rrbmbp1ews7d")
	addr1 := ledgerApi.GenesisAddress()
	if addr1 != genesisAddress {
		t.Fatalf("get genesis address error,should be %s,but get %s", genesisAddress.String(), addr1.String())
	}
	addr2 := ledgerApi.GasAddress()
	if addr2 != gasAddress {
		t.Fatalf("get genesis address error,should be %s,but get %s", gasAddress.String(), addr2.String())
	}
	var chainToken, gasToken types.Hash
	_ = chainToken.Of("45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad")
	_ = gasToken.Of("ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81")
	token1 := ledgerApi.ChainToken()
	token2 := ledgerApi.GasToken()
	if token1 != chainToken || token2 != gasToken {
		t.Fatal("get chain token or gas token error")
	}
	blk1 := ledgerApi.GenesisMintageBlock()
	blk2 := ledgerApi.GenesisBlock()
	blk3 := ledgerApi.GasBlock()
	if blk1.GetHash() != ledgerApi.GenesisMintageHash() || blk2.GetHash() != ledgerApi.GenesisBlockHash() || blk3.GetHash() != ledgerApi.GasBlockHash() {
		t.Fatal("get genesis block hash error")
	}
	if !ledgerApi.IsGenesisToken(chainToken) {
		t.Fatal("chain token error")
	}
	bs := ledgerApi.AllGenesisBlocks()
	if len(bs) != 4 {
		t.Fatal("get all genesis block error")
	}
	if !ledgerApi.IsGenesisBlock(&bs[0]) {
		t.Fatal("genesis block error")
	}
}
