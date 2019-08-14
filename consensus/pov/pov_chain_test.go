package pov

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/trie"
)

type povChainMockData struct {
	config *config.Config
	ledger ledger.Store
}

func setupPovChainTestCase(t *testing.T) (func(t *testing.T), *povChainMockData) {
	t.Parallel()

	md := &povChainMockData{}

	uid := uuid.New().String()
	rootDir := filepath.Join(config.QlcTestDataDir(), uid)
	md.config, _ = config.DefaultConfig(rootDir)

	lDir := filepath.Join(rootDir, "ledger")
	_ = os.RemoveAll(lDir)
	md.ledger = ledger.NewLedger(lDir)

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

func TestPovChain_InsertBlocks(t *testing.T) {
	teardownTestCase, md := setupPovChainTestCase(t)
	defer teardownTestCase(t)

	chain := NewPovBlockChain(md.config, md.ledger)

	chain.Init()
	chain.Start()

	genesisBlk := chain.GenesisBlock()
	latestBlk := chain.LatestBlock()

	if latestBlk.Hash != genesisBlk.Hash {
		t.Fatal("genesis hash invalid")
	}

	statTrie := trie.NewTrie(md.ledger.DBStore(), &(latestBlk.StateHash), trie.NewSimpleTrieNodePool())

	blk1, _ := mock.GeneratePovBlock(latestBlk, 0)
	err := chain.InsertBlock(blk1, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk2, _ := mock.GeneratePovBlock(blk1, 0)
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
	if retBlk1 == nil || retBlk1.Hash != blk1.GetHash() {
		t.Fatalf("failed to get block1 %s", blk1.GetHash())
	}

	retBlk2 := chain.GetBlockByHash(blk2.GetHash())
	if retBlk2 == nil || retBlk2.Hash != blk2.GetHash() {
		t.Fatalf("failed to get block2 %s", blk2.GetHash())
	}

	retBlk3 := chain.GetBlockByHash(blk3.GetHash())
	if retBlk3 == nil || retBlk3.Hash != blk3.GetHash() {
		t.Fatalf("failed to get block3 %s", blk3.GetHash())
	}

	chain.Stop()
}

func TestPovChain_ForkChain(t *testing.T) {
	teardownTestCase, md := setupPovChainTestCase(t)
	defer teardownTestCase(t)

	chain := NewPovBlockChain(md.config, md.ledger)

	chain.Init()
	chain.Start()

	genesisBlk := chain.GenesisBlock()
	latestBlk := chain.LatestBlock()

	if latestBlk.Hash != genesisBlk.Hash {
		t.Fatal("genesis hash invalid")
	}

	statTrie := trie.NewTrie(md.ledger.DBStore(), &(latestBlk.StateHash), trie.NewSimpleTrieNodePool())

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

	retBlk22, err := chain.GetBlockByHeight(blk22.GetHeight())
	if retBlk22 == nil || retBlk22.Hash != retBlk22.GetHash() {
		t.Fatalf("failed to get block22 %s", blk22.GetHash())
	}

	retBlk3, err := chain.GetBlockByHeight(blk3.GetHeight())
	if retBlk3 == nil || retBlk3.Hash != blk3.GetHash() {
		t.Fatalf("failed to get block3 %s", blk3.GetHash())
	}

	chain.Stop()
}

func TestPovChain_ForkChain_WithTx(t *testing.T) {
	teardownTestCase, md := setupPovChainTestCase(t)
	defer teardownTestCase(t)

	chain := NewPovBlockChain(md.config, md.ledger)

	chain.Init()
	chain.Start()

	genesisBlk := chain.GenesisBlock()
	latestBlk := chain.LatestBlock()

	if latestBlk.Hash != genesisBlk.Hash {
		t.Fatal("genesis hash invalid")
	}

	statTrie := trie.NewTrie(md.ledger.DBStore(), &(latestBlk.StateHash), trie.NewSimpleTrieNodePool())

	blk1, _ := mock.GeneratePovBlock(latestBlk, 5)
	err := chain.InsertBlock(blk1, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk21, _ := mock.GeneratePovBlock(blk1, 5)
	err = chain.InsertBlock(blk21, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk22, _ := mock.GeneratePovBlock(blk1, 5)
	err = chain.InsertBlock(blk22, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	blk3, _ := mock.GeneratePovBlock(blk22, 5)
	err = chain.InsertBlock(blk3, statTrie)
	if err != nil {
		t.Fatal(err)
	}

	retBlk22, err := chain.GetBlockByHeight(blk22.GetHeight())
	if retBlk22 == nil || retBlk22.Hash != retBlk22.GetHash() {
		t.Fatalf("failed to get block22 %s", blk22.GetHash())
	}

	retBlk3, err := chain.GetBlockByHeight(blk3.GetHeight())
	if retBlk3 == nil || retBlk3.Hash != blk3.GetHash() {
		t.Fatalf("failed to get block3 %s", blk3.GetHash())
	}

	chain.Stop()
}
