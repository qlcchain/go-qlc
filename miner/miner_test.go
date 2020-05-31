package miner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/google/uuid"

	qctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

type mockDataTestMiner struct {
	cfg *config.Config
	eb  event.EventBus
	feb *event.FeedEventBus
	l   *ledger.Ledger
	cc  *qctx.ChainContext

	ch *mockPovChainReader
	cs *mockPovConsensusReader
	tp *mockPovTxPoolReader

	m *Miner
}

type mockPovChainReader struct {
	md            *mockDataTestMiner
	mockPovBlocks map[string]*types.PovBlock
}

func (ch *mockPovChainReader) LatestHeader() *types.PovHeader {
	return ch.mockPovBlocks["LatestHeader"].GetHeader()
}

func (ch *mockPovChainReader) LatestBlock() *types.PovBlock {
	return ch.mockPovBlocks["LatestBlock"]
}

func (ch *mockPovChainReader) TransitStateDB(height uint64, txs []*types.PovTransaction, gsdb *statedb.PovGlobalStateDB) error {
	return nil
}

func (ch *mockPovChainReader) CalcBlockReward(header *types.PovHeader) (types.Balance, types.Balance, error) {
	return types.ZeroBalance, types.ZeroBalance, nil
}

func (ch *mockPovChainReader) CalcPastMedianTime(prevHeader *types.PovHeader) uint32 {
	return 0
}

type mockPovTxPoolReader struct {
	md *mockDataTestMiner
}

func (tp *mockPovTxPoolReader) SelectPendingTxs(gsdb *statedb.PovGlobalStateDB, limit int) []*types.StateBlock {
	return nil
}

func (tp *mockPovTxPoolReader) LastUpdated() time.Time {
	return time.Now()
}

func (tp *mockPovTxPoolReader) GetPendingTxNum() uint32 {
	return 0
}

type mockPovConsensusReader struct {
	md *mockDataTestMiner
}

func (cs *mockPovConsensusReader) PrepareHeader(header *types.PovHeader) error {
	return nil
}

func setupTestCasePov(t *testing.T) (func(t *testing.T), *mockDataTestMiner) {
	t.Parallel()

	md := new(mockDataTestMiner)

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

	subscriber := event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *topic.EventPovRecvBlockMsg:
			if msg.ResponseChan != nil {
				msg.ResponseChan <- nil
			}
		}
	}), md.eb)

	if err := subscriber.Subscribe(topic.EventPovRecvBlock); err != nil {
		t.Fatalf("failed to subscribe events")
	}

	md.ch = new(mockPovChainReader)
	md.ch.mockPovBlocks = make(map[string]*types.PovBlock)
	md.cs = new(mockPovConsensusReader)
	md.tp = new(mockPovTxPoolReader)

	md.m = NewMiner(cm.ConfigFile, md.ch, md.tp, md.cs)

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

func TestMiner_Work(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	_ = md.cc.Start()
	defer func() {
		_ = md.cc.Stop()
	}()
	time.Sleep(10 * time.Millisecond)

	_ = md.m.Init()
	_ = md.m.Start()

	md.eb.Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(10 * time.Millisecond)

	allPovBlks, err := mockMinerGeneratePovBlocksToLedger(md.l, 1)
	if err != nil {
		t.Fatal(err)
	}

	md.ch.mockPovBlocks["LatestHeader"] = allPovBlks[0]
	md.ch.mockPovBlocks["LatestBlock"] = allPovBlks[0]

	minerAcc := mock.Account()

	md.m.GetAccounts()
	md.m.povWorker.GetMinerAccount()

	inArgs1 := make(map[interface{}]interface{})
	inArgs1["minerAddr"] = minerAcc.Address()
	inArgs1["algoName"] = types.ALGO_SHA256D.String()
	outArgs1 := make(map[interface{}]interface{})
	md.m.povWorker.OnEventRpcSyncCall(&topic.EventRPCSyncCallMsg{Name: "Miner.GetWork", In: inArgs1, Out: outArgs1})
	if outArgs1["mineBlock"] == nil {
		t.Fatal("failed to GetWork")
	}
	mineBlk := outArgs1["mineBlock"].(*types.PovMineBlock)

	inArgs2 := make(map[interface{}]interface{})
	mineRes := types.NewPovMineResult()
	mineRes.WorkHash = mineBlk.WorkHash

	mineRes.CoinbaseExtra = []byte{1, 2, 3, 4, 5, 6, 7, 8}
	mineBlk.Header.CbTx.TxIns[0].Extra = mineRes.CoinbaseExtra
	mineRes.CoinbaseHash = mineBlk.Header.CbTx.ComputeHash()
	mineBlk.AllTxHashes[0] = &mineRes.CoinbaseHash
	mineRes.MerkleRoot = merkle.CalcMerkleTreeRootHash(mineBlk.AllTxHashes)

	mineBlk.Header.BasHdr.Timestamp = mineRes.Timestamp
	mineBlk.Header.BasHdr.Nonce = mineRes.Nonce
	mineBlk.Header.BasHdr.MerkleRoot = mineRes.MerkleRoot
	mineRes.BlockHash = mineBlk.Header.ComputeHash()

	inArgs2["mineResult"] = mineRes
	outArgs2 := make(map[interface{}]interface{})
	md.m.povWorker.OnEventRpcSyncCall(&topic.EventRPCSyncCallMsg{Name: "Miner.SubmitWork", In: inArgs2, Out: outArgs2})

	//md.m.povWorker.submitBlock()

	_ = md.m.Stop()
}

func TestMiner_Mining(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	_ = md.cc.Start()
	defer func() {
		_ = md.cc.Stop()
	}()
	time.Sleep(10 * time.Millisecond)

	_ = md.m.Init()
	_ = md.m.Start()

	md.eb.Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(10 * time.Millisecond)

	allPovBlks, err := mockMinerGeneratePovBlocksToLedger(md.l, 1)
	if err != nil {
		t.Fatal(err)
	}

	md.ch.mockPovBlocks["LatestHeader"] = allPovBlks[0]
	md.ch.mockPovBlocks["LatestBlock"] = allPovBlks[0]

	minerAcc := mock.Account()

	inArgs1 := make(map[interface{}]interface{})
	inArgs1["minerAddr"] = minerAcc.Address()
	inArgs1["algoName"] = types.ALGO_SHA256D.String()
	outArgs1 := make(map[interface{}]interface{})
	md.m.povWorker.OnEventRpcSyncCall(&topic.EventRPCSyncCallMsg{Name: "Miner.StartMining", In: inArgs1, Out: outArgs1})

	time.Sleep(3100 * time.Millisecond)

	inArgs2 := make(map[interface{}]interface{})
	outArgs2 := make(map[interface{}]interface{})
	md.m.povWorker.OnEventRpcSyncCall(&topic.EventRPCSyncCallMsg{Name: "Miner.StopMining", In: inArgs2, Out: outArgs2})

	inArgs3 := make(map[interface{}]interface{})
	outArgs3 := make(map[interface{}]interface{})
	md.m.povWorker.OnEventRpcSyncCall(&topic.EventRPCSyncCallMsg{Name: "Miner.GetMiningInfo", In: inArgs3, Out: outArgs3})

	md.m.povWorker.checkValidMiner()

	md.m.povWorker.mineNextBlock()

	_ = md.m.Stop()
}

func mockMinerGeneratePovBlocksToLedger(l ledger.Store, blkNum int) ([]*types.PovBlock, error) {
	var prevBlk *types.PovBlock
	var allBlks []*types.PovBlock
	for i := 0; i < blkNum; i++ {
		blk1, td1 := mock.GeneratePovBlock(prevBlk, 0)
		err := l.AddPovBlock(blk1, td1)
		if err != nil {
			return nil, err
		}
		for txIdx, txPov := range blk1.GetAllTxs() {
			txl := &types.PovTxLookup{BlockHash: blk1.GetHash(), BlockHeight: blk1.GetHeight(), TxIndex: uint64(txIdx)}
			_ = l.AddPovTxLookup(txPov.Hash, txl)
		}

		err = l.AddPovBestHash(blk1.GetHeight(), blk1.GetHash())
		if err != nil {
			return nil, err
		}
		err = l.SetPovLatestHeight(blk1.GetHeight())
		if err != nil {
			return nil, err
		}
		allBlks = append(allBlks, blk1)

		prevBlk = blk1
	}
	return allBlks, nil
}
