package pov

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/trie"
)

type povChainMockData struct {
	config *config.Config
	ledger ledger.Store
	eb     event.EventBus
}

func setupPovChainTestCase(t *testing.T) (func(t *testing.T), *povChainMockData) {
	t.Parallel()

	md := &povChainMockData{}

	uid := uuid.New().String()
	rootDir := filepath.Join(config.QlcTestDataDir(), uid)
	md.config, _ = config.DefaultConfig(rootDir)

	lDir := filepath.Join(rootDir, "ledger")
	_ = os.RemoveAll(lDir)
	cm := config.NewCfgManager(lDir)
	_, _ = cm.Load()
	md.ledger = ledger.NewLedger(cm.ConfigFile)

	genBlk, genTd := mock.GenerateGenesisPovBlock()
	_ = md.ledger.AddPovBlock(genBlk, genTd)

	md.eb = event.GetEventBus(uid)

	return func(t *testing.T) {
		err := md.ledger.DBStore().Close()
		if err != nil {
			t.Fatal(err)
		}

		err = os.RemoveAll(rootDir)
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestPovChain_DebugInfo(t *testing.T) {
	teardownTestCase, md := setupPovChainTestCase(t)
	defer teardownTestCase(t)

	chain := NewPovBlockChain(md.config, md.eb, md.ledger)

	_ = chain.Init()
	_ = chain.Start()

	info := chain.GetDebugInfo()
	if info == nil || len(info) == 0 {
		t.Fatal("debug info not exist")
	}

	_ = chain.Stop()
}

func TestPovChain_InsertBlocks(t *testing.T) {
	teardownTestCase, md := setupPovChainTestCase(t)
	defer teardownTestCase(t)

	chain := NewPovBlockChain(md.config, md.eb, md.ledger)

	_ = chain.Init()
	_ = chain.Start()

	genesisBlk := chain.GenesisBlock()
	latestBlk := chain.LatestBlock()

	if latestBlk.GetHash() != genesisBlk.GetHash() {
		t.Fatal("genesis hash invalid")
	}

	stateHash := latestBlk.GetStateHash()
	statTrie := trie.NewTrie(md.ledger.DBStore(), &stateHash, trie.NewSimpleTrieNodePool())

	blk1, _ := mock.GeneratePovBlock(latestBlk, 0)
	err := chain.InsertBlock(blk1, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk2, _ := mock.GeneratePovBlock(blk1, 10)
	setupPovTxBlock2Ledger(md, blk2)
	err = chain.InsertBlock(blk2, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk3, _ := mock.GeneratePovBlock(blk2, 0)
	err = chain.InsertBlock(blk3, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	retBlk1 := chain.GetBlockByHash(blk1.GetHash())
	if retBlk1 == nil || retBlk1.GetHash() != blk1.GetHash() {
		t.Fatalf("failed to get block1 %s", blk1.GetHash())
	}

	retBlk2 := chain.GetBlockByHash(blk2.GetHash())
	if retBlk2 == nil || retBlk2.GetHash() != blk2.GetHash() {
		t.Fatalf("failed to get block2 %s", blk2.GetHash())
	}

	retBlk3 := chain.GetBlockByHash(blk3.GetHash())
	if retBlk3 == nil || retBlk3.GetHash() != blk3.GetHash() {
		t.Fatalf("failed to get block3 %s", blk3.GetHash())
	}

	retBlk3 = chain.GetBestBlockByHash(blk3.GetHash())
	if retBlk3 == nil || retBlk3.GetHash() != blk3.GetHash() {
		t.Fatalf("failed to get best block3 %s", blk3.GetHash())
	}

	retBlk3, _ = chain.GetBlockByHeight(blk3.GetHeight())
	if retBlk3 == nil || retBlk3.GetHash() != blk3.GetHash() {
		t.Fatalf("failed to get block3 by height %d", blk3.GetHeight())
	}

	retHdr3 := chain.GetHeaderByHeight(blk3.GetHeight())
	if retHdr3 == nil || retHdr3.GetHash() != blk3.GetHash() {
		t.Fatalf("failed to get header3 by height %d", blk3.GetHeight())
	}

	chain.CalcPastMedianTime(blk3.GetHeader())

	_ = chain.Stop()
}

func TestPovChain_ForkChain(t *testing.T) {
	teardownTestCase, md := setupPovChainTestCase(t)
	defer teardownTestCase(t)

	chain := NewPovBlockChain(md.config, md.eb, md.ledger)

	_ = chain.Init()
	_ = chain.Start()

	genesisBlk := chain.GenesisBlock()
	latestBlk := chain.LatestBlock()

	if latestBlk.GetHash() != genesisBlk.GetHash() {
		t.Fatal("genesis hash invalid")
	}

	stateHash := latestBlk.GetStateHash()
	statTrie := trie.NewTrie(md.ledger.DBStore(), &stateHash, trie.NewSimpleTrieNodePool())

	blk1, _ := mock.GeneratePovBlock(latestBlk, 0)
	err := chain.InsertBlock(blk1, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk21, _ := mock.GeneratePovBlock(blk1, 0)
	err = chain.InsertBlock(blk21, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk22, _ := mock.GeneratePovBlock(blk1, 0)
	err = chain.InsertBlock(blk22, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk3, _ := mock.GeneratePovBlock(blk22, 0)
	err = chain.InsertBlock(blk3, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	retBlk22, _ := chain.GetBlockByHeight(blk22.GetHeight())
	if retBlk22 == nil || retBlk22.GetHash() != retBlk22.GetHash() {
		t.Fatalf("failed to get block22 %s", blk22.GetHash())
	}

	retBlk3, _ := chain.GetBlockByHeight(blk3.GetHeight())
	if retBlk3 == nil || retBlk3.GetHash() != blk3.GetHash() {
		t.Fatalf("failed to get block3 %s", blk3.GetHash())
	}

	_ = chain.Stop()
}

func TestPovChain_ForkChain_WithTx(t *testing.T) {
	teardownTestCase, md := setupPovChainTestCase(t)
	defer teardownTestCase(t)

	chain := NewPovBlockChain(md.config, md.eb, md.ledger)

	_ = chain.Init()
	_ = chain.Start()

	genesisBlk := chain.GenesisBlock()
	latestBlk := chain.LatestBlock()

	if latestBlk.GetHash() != genesisBlk.GetHash() {
		t.Fatal("genesis hash invalid")
	}

	stateHash := latestBlk.GetStateHash()
	statTrie := trie.NewTrie(md.ledger.DBStore(), &stateHash, trie.NewSimpleTrieNodePool())

	blk1, _ := mock.GeneratePovBlock(latestBlk, 5)
	setupPovTxBlock2Ledger(md, blk1)
	err := chain.InsertBlock(blk1, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk21, _ := mock.GeneratePovBlock(blk1, 5)
	setupPovTxBlock2Ledger(md, blk21)
	err = chain.InsertBlock(blk21, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk22, _ := mock.GeneratePovBlock(blk1, 5)
	setupPovTxBlock2Ledger(md, blk22)
	err = chain.InsertBlock(blk22, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk3, _ := mock.GeneratePovBlock(blk22, 5)
	setupPovTxBlock2Ledger(md, blk3)
	err = chain.InsertBlock(blk3, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	retBlk22, _ := chain.GetBlockByHeight(blk22.GetHeight())
	if retBlk22 == nil || retBlk22.GetHash() != retBlk22.GetHash() {
		t.Fatalf("failed to get block22 %s", blk22.GetHash())
	}

	retBlk3, _ := chain.GetBlockByHeight(blk3.GetHeight())
	if retBlk3 == nil || retBlk3.GetHash() != blk3.GetHash() {
		t.Fatalf("failed to get block3 %s", blk3.GetHash())
	}

	_ = chain.Stop()
}

func TestPovChain_TrieState(t *testing.T) {
	teardownTestCase, md := setupPovChainTestCase(t)
	defer teardownTestCase(t)

	chain := NewPovBlockChain(md.config, md.eb, md.ledger)

	_ = chain.Init()
	_ = chain.Start()

	latestBlk := chain.LatestBlock()

	prevStateHash := latestBlk.GetStateHash()

	blk1, _ := mock.GeneratePovBlock(latestBlk, 5)
	setupPovTxBlock2Ledger(md, blk1)

	accTxsBlk1 := blk1.GetAccountTxs()
	curStatTrie, err := chain.GenStateTrie(blk1.GetHeight(), prevStateHash, accTxsBlk1)
	if err != nil {
		t.Fatal(err)
	}

	err = chain.InsertBlock(blk1, curStatTrie)
	if err != nil {
		t.Fatal(err)
	}

	curStatHash := curStatTrie.Hash()
	if *curStatHash == prevStateHash {
		t.Fatalf("state hash should not equal")
	}

	curStatTrieInDB := trie.NewTrie(md.ledger.DBStore(), curStatHash, nil)

	for _, accTx := range accTxsBlk1 {
		as := chain.GetAccountState(curStatTrieInDB, accTx.Block.GetAddress())
		if as == nil || as.Balance.Compare(accTx.Block.Balance) != types.BalanceCompEqual {
			t.Fatalf("invalid account state in state trie")
		}
		repAddr := accTx.Block.GetRepresentative()
		if !repAddr.IsZero() {
			rs := chain.GetRepState(curStatTrieInDB, repAddr)
			if rs == nil {
				t.Fatalf("invalid rep state in state trie")
			}
		}
	}

	_ = chain.Stop()
}

func setupPovTxBlock2Ledger(md *povChainMockData, povBlock *types.PovBlock) {
	for _, txPov := range povBlock.Body.Txs {
		if txPov.Block != nil {
			_ = md.ledger.AddStateBlock(txPov.Block)
		}
	}
}
