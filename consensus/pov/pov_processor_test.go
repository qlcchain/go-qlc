package pov

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
)

type povProcessorMockData struct {
	config *config.Config
	eb     event.EventBus
	ledger ledger.Store

	chain    *mockPovProcessorChainReader
	verifier *mockPovProcessorVerifier
	syncer   *mockPovProcessorSyncer
}

type mockPovProcessorChainReader struct {
	blocks map[types.Hash]*types.PovBlock
}

func (c *mockPovProcessorChainReader) HasBestBlock(hash types.Hash, height uint64) bool {
	if c.blocks[hash] != nil {
		return true
	}
	return false
}

func (c *mockPovProcessorChainReader) GetBlockByHash(hash types.Hash) *types.PovBlock {
	genesisBlk := common.GenesisPovBlock()
	if hash == genesisBlk.GetHash() {
		return &genesisBlk
	}
	blk := c.blocks[hash]
	if blk != nil {
		return blk
	}

	return nil
}

func (c *mockPovProcessorChainReader) InsertBlock(block *types.PovBlock, sdb *statedb.PovGlobalStateDB) error {
	c.blocks[block.GetHash()] = block
	return nil
}

func (c *mockPovProcessorChainReader) LatestHeader() *types.PovHeader {
	return nil
}

type mockPovProcessorVerifier struct {
	resultStat *PovVerifyStat
}

func (v *mockPovProcessorVerifier) VerifyFull(block *types.PovBlock) *PovVerifyStat {
	if v.resultStat != nil {
		return v.resultStat
	}
	return &PovVerifyStat{
		Result: process.Progress,
	}
}

type mockPovProcessorSyncer struct{}

func (s *mockPovProcessorSyncer) requestBlocksByHashes(reqBlkHashes []*types.Hash, peerID string) {}

func (s *mockPovProcessorSyncer) requestSyncFrontiers(peerID string) {}

func setupPovProcessorTestCase(t *testing.T) (func(t *testing.T), *povProcessorMockData) {
	t.Parallel()

	md := &povProcessorMockData{
		chain: &mockPovProcessorChainReader{
			blocks: make(map[types.Hash]*types.PovBlock),
		},
		verifier: &mockPovProcessorVerifier{},
		syncer:   &mockPovProcessorSyncer{},
	}

	uid := uuid.New().String()
	rootDir := filepath.Join(config.QlcTestDataDir(), uid)
	md.config, _ = config.DefaultConfig(rootDir)

	lDir := filepath.Join(rootDir, "ledger")
	_ = os.RemoveAll(lDir)

	cm := config.NewCfgManager(lDir)
	cm.Load()
	md.ledger = ledger.NewLedger(cm.ConfigFile)

	md.eb = event.GetEventBus(lDir)

	return func(t *testing.T) {
		err := md.ledger.Close()
		if err != nil {
			t.Fatal(err)
		}

		err = os.RemoveAll(rootDir)
		if err != nil {
			t.Fatal(err)
		}

		err = md.eb.Close()
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestPovProcessor_SimplteTest(t *testing.T) {
	teardownTestCase, md := setupPovProcessorTestCase(t)
	defer teardownTestCase(t)

	processor := NewPovBlockProcessor(md.eb, md.ledger, md.chain, md.verifier, md.syncer)

	processor.Init()
	processor.Start()

	info := processor.GetDebugInfo()
	if info == nil || len(info) == 0 {
		t.Fatal("debug info not exist")
	}

	processor.Stop()
}

func TestPovProcessor_AddBlock(t *testing.T) {
	teardownTestCase, md := setupPovProcessorTestCase(t)
	defer teardownTestCase(t)

	processor := NewPovBlockProcessor(md.eb, md.ledger, md.chain, md.verifier, md.syncer)

	processor.Init()
	processor.Start()

	genesisBlk := common.GenesisPovBlock()

	blk1, _ := mock.GeneratePovBlock(&genesisBlk, 0)
	processor.AddBlock(blk1, types.PovBlockFromRemoteBroadcast, "test")

	blk2, _ := mock.GeneratePovBlock(blk1, 0)
	processor.AddBlock(blk2, types.PovBlockFromRemoteBroadcast, "test")

	blk3, _ := mock.GeneratePovBlock(blk2, 0)
	processor.AddBlock(blk3, types.PovBlockFromRemoteBroadcast, "test")

	time.Sleep(time.Second)

	retBlk1 := md.chain.GetBlockByHash(blk1.GetHash())
	if retBlk1 == nil {
		t.Fatalf("failed to add block1 %s", blk1.GetHash())
	}

	retBlk2 := md.chain.GetBlockByHash(blk2.GetHash())
	if retBlk2 == nil {
		t.Fatalf("failed to add block2 %s", blk2.GetHash())
	}

	retBlk3 := md.chain.GetBlockByHash(blk3.GetHash())
	if retBlk3 == nil {
		t.Fatalf("failed to add block3 %s", blk3.GetHash())
	}

	processor.Stop()
}

func TestPovProcessor_OrphanBlock(t *testing.T) {
	teardownTestCase, md := setupPovProcessorTestCase(t)
	defer teardownTestCase(t)

	processor := NewPovBlockProcessor(md.eb, md.ledger, md.chain, md.verifier, md.syncer)

	processor.Init()
	processor.Start()

	processor.onPovSyncState(topic.SyncDone)

	genesisBlk := common.GenesisPovBlock()

	blk1, _ := mock.GeneratePovBlock(&genesisBlk, 0)
	processor.AddBlock(blk1, types.PovBlockFromRemoteBroadcast, "test")

	blk2, _ := mock.GeneratePovBlock(blk1, 0)

	blk3, _ := mock.GeneratePovBlock(blk2, 0)
	processor.AddBlock(blk3, types.PovBlockFromRemoteBroadcast, "test")

	blk4, _ := mock.GeneratePovBlock(blk3, 0)
	processor.AddBlock(blk4, types.PovBlockFromRemoteBroadcast, "test")

	time.Sleep(time.Second)

	processor.onRequestOrphanBlocksTimer()

	processor.onCheckOrphanBlocksTimer()

	retBlk1 := md.chain.GetBlockByHash(blk1.GetHash())
	if retBlk1 == nil {
		t.Fatalf("failed to add block1 %s", blk1.GetHash())
	}

	retBlk3 := md.chain.GetBlockByHash(blk3.GetHash())
	if retBlk3 != nil {
		t.Fatalf("block3 %s is not orphan", blk3.GetHash())
	}

	processor.AddBlock(blk2, types.PovBlockFromRemoteBroadcast, "test")

	time.Sleep(time.Second)

	retBlk2 := md.chain.GetBlockByHash(blk2.GetHash())
	if retBlk2 == nil {
		t.Fatalf("failed to add block2 %s", blk2.GetHash())
	}

	retBlk3 = md.chain.GetBlockByHash(blk3.GetHash())
	if retBlk3 == nil {
		t.Fatalf("failed to add block3 %s", blk3.GetHash())
	}

	retBlk4 := md.chain.GetBlockByHash(blk4.GetHash())
	if retBlk4 == nil {
		t.Fatalf("failed to add block4 %s", blk4.GetHash())
	}

	processor.Stop()
}

func TestPovProcessor_PendingBlock(t *testing.T) {
	teardownTestCase, md := setupPovProcessorTestCase(t)
	defer teardownTestCase(t)

	processor := NewPovBlockProcessor(md.eb, md.ledger, md.chain, md.verifier, md.syncer)

	_ = processor.Init()
	_ = processor.Start()

	processor.onPovSyncState(topic.SyncDone)

	genesisBlk := common.GenesisPovBlock()

	blk1, _ := mock.GeneratePovBlock(&genesisBlk, 1)

	md.verifier.resultStat = &PovVerifyStat{
		Result: process.GapTransaction,
		GapTxs: make(map[types.Hash]process.ProcessResult),
	}
	accTxs := blk1.GetAccountTxs()
	for _, txPov := range accTxs {
		md.verifier.resultStat.GapTxs[txPov.Hash] = process.GapTransaction
	}

	_ = processor.AddBlock(blk1, types.PovBlockFromRemoteBroadcast, "test")

	time.Sleep(100 * time.Millisecond)

	retBlk1 := md.chain.GetBlockByHash(blk1.GetHash())
	if retBlk1 != nil {
		t.Fatalf("block1 %s should in pending", blk1.GetHash())
	}

	for _, txPov := range blk1.Body.Txs {
		if txPov.Block != nil {
			_ = md.ledger.AddStateBlock(txPov.Block)
		}
	}
	md.verifier.resultStat = nil

	processor.onCheckTxPendingBlocksTimer()

	time.Sleep(100 * time.Millisecond)

	retBlk1 = md.chain.GetBlockByHash(blk1.GetHash())
	if retBlk1 == nil {
		t.Fatalf("block1 %s should exist", blk1.GetHash())
	}

	_ = processor.Stop()
}
