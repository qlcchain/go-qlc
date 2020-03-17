package pov

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

type mockDataPovPow struct {
	cfg *config.Config
	l   ledger.Store

	mockChain *mockPovPowConsensusChainReader
}

type mockPovPowConsensusChainReader struct {
	md           *mockDataPovPow
	mockPovBlock map[string]*types.PovBlock
}

func (cr *mockPovPowConsensusChainReader) TrieDb() storage.Store {
	return cr.md.l.DBStore()
}

func (cr *mockPovPowConsensusChainReader) GetHeaderByHash(hash types.Hash) *types.PovHeader {
	return cr.mockPovBlock["GetHeaderByHash"].GetHeader()
}

func (cr *mockPovPowConsensusChainReader) RelativeAncestor(header *types.PovHeader, distance uint64) *types.PovHeader {
	return cr.mockPovBlock["RelativeAncestor"].GetHeader()
}

func setupTestCasePovPow(t *testing.T) (func(t *testing.T), *mockDataPovPow) {
	t.Parallel()

	md := &mockDataPovPow{
		mockChain: &mockPovPowConsensusChainReader{},
	}

	md.mockChain.md = md

	uid := uuid.New().String()
	rootDir := filepath.Join(config.QlcTestDataDir(), uid)
	md.cfg, _ = config.DefaultConfig(rootDir)

	lDir := filepath.Join(rootDir, "ledger")
	_ = os.RemoveAll(lDir)

	cm := config.NewCfgManager(lDir)
	cm.Load()
	md.l = ledger.NewLedger(cm.ConfigFile)

	return func(t *testing.T) {
		err := md.l.Close()
		if err != nil {
			t.Fatal(err)
		}

		err = os.RemoveAll(rootDir)
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestPovPow_PrepareHeader(t *testing.T) {
	teardownTestCase, md := setupTestCasePovPow(t)
	defer teardownTestCase(t)

	csPow := NewConsensusPow(md.mockChain)

	_ = csPow.Init()
	_ = csPow.Start()

	md.mockChain.mockPovBlock = make(map[string]*types.PovBlock)

	blkFirstInPrevCycle, _ := mock.GeneratePovBlock(nil, 0)
	blkFirstInPrevCycle.Header.BasHdr.Height = common.PovMinerVerifyHeightStart
	md.mockChain.mockPovBlock["RelativeAncestor"] = blkFirstInPrevCycle

	blkPrev, _ := mock.GeneratePovBlock(nil, 0)
	blkPrev.Header.BasHdr.Height = blkFirstInPrevCycle.GetHeight() + uint64(common.PovChainTargetCycle) - 1
	md.mockChain.mockPovBlock["GetHeaderByHash"] = blkPrev

	blk1, _ := mock.GeneratePovBlock(nil, 0)
	blk1.Header.BasHdr.Height = blkPrev.GetHeight() + 1
	blk1Hdr := blk1.GetHeader()

	err := csPow.PrepareHeader(blk1Hdr)
	if err != nil {
		t.Fatal(err)
	}

	err = csPow.FinalizeHeader(blk1Hdr)
	if err != nil {
		t.Fatal(err)
	}

	_ = csPow.verifyProducer(blk1Hdr)

	_ = csPow.VerifyHeader(blk1Hdr)

	// blk2

	md.mockChain.mockPovBlock["GetHeaderByHash"] = blk1

	blk2, _ := mock.GeneratePovBlock(blk1, 0)
	blk2Hdr := blk2.GetHeader()

	err = csPow.PrepareHeader(blk2Hdr)
	if err != nil {
		t.Fatal(err)
	}

	err = csPow.FinalizeHeader(blk2Hdr)
	if err != nil {
		t.Fatal(err)
	}

	_ = csPow.Stop()
}
