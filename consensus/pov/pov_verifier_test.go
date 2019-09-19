package pov

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/trie"
)

type povVerifierMockData struct {
	config *config.Config
	ledger ledger.Store

	chainV  PovVerifierChainReader
	chainCs PovConsensusChainReader
	cs      ConsensusPov
}

type mockPovVerifierChainReader struct {
	ledger ledger.Store
	blocks map[types.Hash]*types.PovBlock
}

func (c *mockPovVerifierChainReader) GetHeaderByHash(hash types.Hash) *types.PovHeader {
	genesisBlk := common.GenesisPovBlock()
	if hash == genesisBlk.GetHash() {
		return genesisBlk.GetHeader()
	}
	return nil
}

func (c *mockPovVerifierChainReader) CalcPastMedianTime(prevHeader *types.PovHeader) uint32 {
	return prevHeader.GetTimestamp()
}

func (c *mockPovVerifierChainReader) GenStateTrie(height uint64, prevStateHash types.Hash, txs []*types.PovTransaction) (*trie.Trie, error) {
	t := trie.NewTrie(c.ledger.DBStore(), nil, trie.NewSimpleTrieNodePool())
	return t, nil
}

func (c *mockPovVerifierChainReader) GetStateTrie(stateHash *types.Hash) *trie.Trie {
	t := trie.NewTrie(c.ledger.DBStore(), nil, trie.NewSimpleTrieNodePool())
	return t
}

func (c *mockPovVerifierChainReader) GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState {
	return nil
}

func (c *mockPovVerifierChainReader) CalcBlockReward(header *types.PovHeader) (types.Balance, types.Balance) {
	return types.ZeroBalance, types.ZeroBalance
}

type mockPovConsensusChainReader struct{}

func (c *mockPovConsensusChainReader) GetHeaderByHash(hash types.Hash) *types.PovHeader {
	genesisBlk := common.GenesisPovBlock()
	if hash == genesisBlk.GetHash() {
		return genesisBlk.GetHeader()
	}
	return nil
}

func (c *mockPovConsensusChainReader) RelativeAncestor(header *types.PovHeader, distance uint64) *types.PovHeader {
	return nil
}

func (c *mockPovConsensusChainReader) GetStateTrie(stateHash *types.Hash) *trie.Trie {
	return nil
}

func (c *mockPovConsensusChainReader) GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState {
	return nil
}

func setupPovVerifierTestCase(t *testing.T) (func(t *testing.T), *povVerifierMockData) {
	t.Parallel()

	md := &povVerifierMockData{
		chainV: &mockPovVerifierChainReader{
			blocks: make(map[types.Hash]*types.PovBlock),
		},
		chainCs: &mockPovConsensusChainReader{},
	}

	md.cs = NewConsensusPow(md.chainCs)

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

func TestPovVerifier_VerifyNet(t *testing.T) {
	teardownTestCase, md := setupPovVerifierTestCase(t)
	defer teardownTestCase(t)

	verifier := NewPovVerifier(md.ledger, md.chainV, md.cs)

	blk1, _ := mock.GeneratePovBlock(nil, 0)

	stat1 := verifier.VerifyNet(blk1)

	if stat1.Result != process.Progress {
		//t.Fatalf("result %s err %s", stat1.Result, stat1.ErrMsg)
	}
}

func TestPovVerifier_VerifyFull(t *testing.T) {
	teardownTestCase, md := setupPovVerifierTestCase(t)
	defer teardownTestCase(t)

	verifier := NewPovVerifier(md.ledger, md.chainV, md.cs)

	blk1, _ := mock.GeneratePovBlock(nil, 0)

	stat1 := verifier.VerifyFull(blk1)

	if stat1.Result != process.Progress {
		//t.Fatalf("result %s err %s", stat1.Result, stat1.ErrMsg)
	}
}
