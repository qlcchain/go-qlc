package pov

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common/storage"

	"github.com/qlcchain/go-qlc/common/statedb"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
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

func (c *mockPovVerifierChainReader) CalcBlockReward(header *types.PovHeader) (types.Balance, types.Balance, error) {
	return types.ZeroBalance, types.ZeroBalance, nil
}

func (c *mockPovVerifierChainReader) TransitStateDB(height uint64, txs []*types.PovTransaction, gsdb *statedb.PovGlobalStateDB) error {
	return nil
}

func (c *mockPovVerifierChainReader) TrieDb() storage.Store {
	return nil
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

func (c *mockPovConsensusChainReader) TrieDb() storage.Store {
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
	cm := config.NewCfgManager(lDir)
	_, _ = cm.Load()
	md.ledger = ledger.NewLedger(cm.ConfigFile)

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

	genBlk, _ := mock.GenerateGenesisPovBlock()
	stat := verifier.VerifyFull(genBlk)
	if stat.Result != process.Progress {
		//t.Fatalf("result %s err %s", stat1.Result, stat1.ErrMsg)
	}

	blk1, _ := mock.GeneratePovBlock(genBlk, 5)
	stat1 := verifier.VerifyFull(blk1)
	if stat1.Result != process.Progress {
		//t.Fatalf("result %s err %s", stat1.Result, stat1.ErrMsg)
	}
}

func TestPovVerifier_VerifyFull_Aux(t *testing.T) {
	teardownTestCase, md := setupPovVerifierTestCase(t)
	defer teardownTestCase(t)

	verifier := NewPovVerifier(md.ledger, md.chainV, md.cs)

	genBlk, _ := mock.GenerateGenesisPovBlock()
	stat := verifier.VerifyFull(genBlk)
	if stat.Result != process.Progress {
		//t.Fatalf("result %s err %s", stat1.Result, stat1.ErrMsg)
	}

	blk1, _ := mock.GeneratePovBlock(genBlk, 5)
	blk1.Header.AuxHdr = mock.GenerateAuxPow(blk1.GetHash())
	stat1 := verifier.VerifyFull(blk1)
	if stat1.Result != process.Progress {
		//t.Fatalf("result %s err %s", stat1.Result, stat1.ErrMsg)
	}
}
