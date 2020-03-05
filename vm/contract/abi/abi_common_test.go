package abi

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"testing"
)

func TestPovHeightToCount(t *testing.T) {
	if PovHeightToCount(1) != 2 {
		t.Fatal()
	}
}

func TestPovHeightRound(t *testing.T) {
	height := uint64(100)
	round := uint64(30)

	if PovHeightRound(height, round) != 89 {
		t.Fatal()
	}
}

func TestPovGetNodeRewardHeightByDay(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	_, err := PovGetNodeRewardHeightByDay(ctx)
	if err == nil {
		t.Fatal()
	}

	povBlock, td := mock.GeneratePovBlock(nil, 0)
	povBlock.Header.BasHdr.Height = common.PovMinerRewardHeightStart - 1
	l.AddPovBlock(povBlock, td)
	l.AddPovBestHash(povBlock.GetHeight(), povBlock.GetHash())
	l.SetPovLatestHeight(povBlock.GetHeight())

	height, err := PovGetNodeRewardHeightByDay(ctx)
	if height != 0 {
		t.Fatal()
	}

	povBlock, td = mock.GeneratePovBlock(povBlock, 0)
	povBlock.Header.BasHdr.Height = common.PovMinerRewardHeightStart + common.PovMinerRewardHeightGapToLatest + uint64(common.POVChainBlocksPerDay)
	l.AddPovBlock(povBlock, td)
	l.AddPovBestHash(povBlock.GetHeight(), povBlock.GetHash())
	l.SetPovLatestHeight(povBlock.GetHeight())

	height, err = PovGetNodeRewardHeightByDay(ctx)
	if height != common.PovMinerRewardHeightStart+common.PovMinerRewardHeightGapToLatest-1 {
		t.Fatal()
	}
}
