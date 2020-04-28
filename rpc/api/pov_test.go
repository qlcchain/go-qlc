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
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
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

func mockPovApiGeneratePovBlocksToLedger(t *testing.T, md *mockDataTestPovApi, blkNum int) []*types.PovBlock {
	var prevBlk *types.PovBlock
	var allBlks []*types.PovBlock
	for i := 0; i < blkNum; i++ {
		blk1, td1 := mock.GeneratePovBlock(prevBlk, 0)
		err := md.l.AddPovBlock(blk1, td1)
		if err != nil {
			t.Fatal(err)
		}
		for txIdx, txPov := range blk1.GetAllTxs() {
			txl := &types.PovTxLookup{BlockHash: blk1.GetHash(), BlockHeight: blk1.GetHeight(), TxIndex: uint64(txIdx)}
			_ = md.l.AddPovTxLookup(txPov.Hash, txl)
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

	allBlks := mockPovApiGeneratePovBlocksToLedger(t, md, 3)

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

	_, err = md.api.GetBlockTDByHeight(bd.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	_, err = md.api.GetBlockTDByHash(bd.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	allTxs := bd.GetAllTxs()
	_, err = md.api.GetTransaction(allTxs[0].Hash)
	_, err = md.api.GetTransactionByBlockHashAndIndex(bd.GetHash(), 0)
	_, err = md.api.GetTransactionByBlockHeightAndIndex(bd.GetHeight(), 0)
}

func TestPovAPI_Mining(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	_ = md.cc.Start()
	defer func() {
		_ = md.cc.Stop()
	}()
	time.Sleep(10 * time.Millisecond)

	md.eb.Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(10 * time.Millisecond)

	mockPovApiGeneratePovBlocksToLedger(t, md, 120)

	minerAcc := mock.Account()

	rspStatus, err := md.api.GetPovStatus()
	if err != nil {
		t.Fatal(err)
	}
	if rspStatus == nil {
		t.Fatalf("failed to GetPovStatus")
	}

	hashInfo, err := md.api.GetHashInfo(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if hashInfo == nil {
		t.Fatalf("failed to GetHashInfo")
	}

	rspMinerInfo, err := md.api.GetMiningInfo()
	if err != nil {
		t.Fatal(err)
	}
	if rspMinerInfo == nil {
		t.Fatalf("failed to GetMiningInfo")
	}

	err = md.api.StartMining(minerAcc.Address(), "SHA256D")
	if err != nil {
		t.Fatal(err)
	}

	err = md.api.StopMining()
	if err != nil {
		t.Fatal(err)
	}

	mds0 := types.NewPovMinerDayStat()
	mds0.DayIndex = 1
	mds0.MinerStats = make(map[string]*types.PovMinerStatItem)
	mds0.MinerStats["miner1"] = &types.PovMinerStatItem{BlockNum: 10}
	mds0.MinerNum = uint32(len(mds0.MinerStats))
	err = md.l.AddPovMinerStat(mds0)
	if err != nil {
		t.Fatal(err)
	}

	rspMinerDs, err := md.api.GetMinerDayStat(1)
	if err != nil {
		t.Fatal(err)
	}
	if rspMinerDs == nil {
		t.Fatalf("failed to GetMinerDayStat")
	}

	rspMinerDs2, err := md.api.GetMinerDayStatByHeight(2879)
	if err != nil {
		t.Fatal(err)
	}
	if rspMinerDs2 == nil {
		t.Fatalf("failed to GetMinerDayStatByHeight")
	}

	rspMinerStats, err := md.api.GetMinerStats(nil)
	if err != nil {
		t.Fatal(err)
	}
	if rspMinerStats == nil {
		t.Fatalf("failed to GetMinerStats")
	}

	rspGetWork, err := md.api.GetWork(minerAcc.Address(), "SHA256D")
	if err != nil {
		t.Fatal(err)
	}
	if rspGetWork == nil {
		t.Fatalf("failed to rspGetWork")
	}

	reqSubmitWork := new(PovApiSubmitWork)
	reqSubmitWork.WorkHash = rspGetWork.WorkHash
	err = md.api.SubmitWork(reqSubmitWork)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPovAPI_ManyBlocks(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	_ = md.cc.Start()
	defer func() {
		_ = md.cc.Stop()
	}()
	time.Sleep(10 * time.Millisecond)

	md.eb.Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(10 * time.Millisecond)

	minerAcc := mock.Account()

	allBlks := mockPovApiGeneratePovBlocksToLedger(t, md, 4320)

	// mock trie state in global db
	lastMockBlk, lastMockTd := mock.GeneratePovBlock(allBlks[len(allBlks)-1], 0)
	gsdb := statedb.NewPovGlobalStateDB(md.l.DBStore(), allBlks[0].GetStateHash())
	as := types.NewPovAccountState()
	as.Balance = types.NewBalance(1234)
	gsdb.SetAccountState(minerAcc.Address(), as)
	rs := types.NewPovRepState()
	rs.Balance = types.NewBalance(4321)
	gsdb.SetRepState(minerAcc.Address(), rs)
	gsdb.CommitToTrie()
	txn := md.l.DBStore().Batch(true)
	gsdb.CommitToDB(txn)
	err := md.l.DBStore().PutBatch(txn)
	if err != nil {
		t.Fatal(err)
	}
	lastMockBlk.Header.CbTx.StateHash = gsdb.GetCurHash()
	mock.UpdatePovHash(lastMockBlk)

	err = md.l.AddPovBlock(lastMockBlk, lastMockTd)
	if err != nil {
		t.Fatal(err)
	}
	for txIdx, txPov := range lastMockBlk.GetAllTxs() {
		txl := &types.PovTxLookup{BlockHash: lastMockBlk.GetHash(), BlockHeight: lastMockBlk.GetHeight(), TxIndex: uint64(txIdx)}
		_ = md.l.AddPovTxLookup(txPov.Hash, txl)
	}

	err = md.l.AddPovBestHash(lastMockBlk.GetHeight(), lastMockBlk.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	err = md.l.SetPovLatestHeight(lastMockBlk.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	// mock account meta
	am := mock.AccountMeta(minerAcc.Address())
	md.l.AddAccountMeta(am, md.l.Cache().GetCache())

	mds0 := types.NewPovMinerDayStat()
	mds0.DayIndex = 0
	mds0.MinerStats = make(map[string]*types.PovMinerStatItem)
	mds0.MinerStats[minerAcc.Address().String()] = &types.PovMinerStatItem{BlockNum: 100, RepBlockNum: 480}
	mds0.MinerNum = uint32(len(mds0.MinerStats))
	err = md.l.AddPovMinerStat(mds0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = md.api.GetLedgerStats()
	if err != nil {
		t.Fatal(err)
	}

	_, err = md.api.GetHashInfo(0, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = md.api.GetLastNHourInfo(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = md.api.GetLastNHourInfo(0, 7200)
	if err != nil {
		t.Fatal(err)
	}

	_, err = md.api.CheckAllAccountStates()
	if err != nil {
		t.Fatal(err)
	}

	latestHdr, err := md.l.GetLatestPovHeader()
	if err != nil {
		t.Fatal(err)
	}

	_, err = md.api.GetRepStats([]types.Address{minerAcc.Address()})
	_ = md.api.GetRepStatesByHeightAndAccount(latestHdr, minerAcc.Address())

	_, err = md.api.GetAllRepStatesByBlockHash(latestHdr.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	_, err = md.api.GetAllRepStatesByBlockHeight(latestHdr.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
	_ = md.api.GetAllOnlineRepStates(latestHdr)

	_, err = md.api.GetLatestAccountState(minerAcc.Address())
	_, err = md.api.GetAccountStateByBlockHash(minerAcc.Address(), latestHdr.GetHash())
	_, err = md.api.GetAccountStateByBlockHeight(minerAcc.Address(), latestHdr.GetHeight())

	_, err = md.api.DumpBlockState(latestHdr.GetHash())
	_, err = md.api.DumpContractState(latestHdr.GetStateHash(), contractaddress.PubKeyDistributionAddress)

	_, err = md.api.GetDiffDayStat(0)
	_, err = md.api.GetDiffDayStatByHeight(latestHdr.GetHeight())
}

func TestPovAPI_PubSub_NewBlock(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	allBlks := mockPovApiGeneratePovBlocksToLedger(t, md, 3)

	rpcCtx := rpc.SubscriptionContext()

	subBlk, err := md.api.NewBlock(rpcCtx)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)

	blk1 := allBlks[0]
	md.api.pubsub.setBlocks(blk1)
	time.Sleep(10 * time.Millisecond)
	md.api.pubsub.fetchBlocks(subBlk.ID)

	md.api.pubsub.removeChan(subBlk.ID)
}
