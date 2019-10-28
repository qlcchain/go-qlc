package abi

import (
	"math/big"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestRepRewardParam_Verify(t *testing.T) {
	param := new(RepRewardParam)
	param.Account = types.ZeroAddress
	param.RewardAmount = big.NewInt(10)
	param.RewardBlocks = uint64(common.POVChainBlocksPerDay)
	param.EndHeight = common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay)
	param.Beneficial = mock.Address()
	param.StartHeight = common.PovMinerRewardHeightStart

	ok, _ := param.Verify()
	if ok {
		t.Fatal("coinbase verify err")
	}

	param.Account = mock.Address()
	param.Beneficial = types.ZeroAddress
	ok, _ = param.Verify()
	if ok {
		t.Fatal("beneficial verify err")
	}

	param.Beneficial = mock.Address()
	param.StartHeight = 0
	ok, _ = param.Verify()
	if ok {
		t.Fatal("StartHeight verify err")
	}

	param.StartHeight = common.PovMinerRewardHeightStart
	param.EndHeight = 0
	ok, _ = param.Verify()
	if ok {
		t.Fatal("EndHeight verify err")
	}

	param.EndHeight = common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay) +
		common.PovMinerMaxRewardBlocksPerCall
	ok, _ = param.Verify()
	if ok {
		t.Fatal("max reward verify err")
	}
}

func TestGetLastRepRewardHeightByAccount(t *testing.T) {
	clean, l := ledger.NewTestLedger()
	defer clean()

	param := new(RepRewardParam)
	param.Account = mock.Address()
	param.RewardAmount = big.NewInt(10)
	param.RewardBlocks = uint64(common.POVChainBlocksPerDay)
	param.EndHeight = common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay)
	param.Beneficial = mock.Address()
	param.StartHeight = common.PovMinerRewardHeightStart

	ctx := vmstore.NewVMContext(l)
	data, _ := RepABI.PackVariable(VariableNameRepReward, param.EndHeight,
		param.RewardBlocks, time.Now().Unix(), param.RewardAmount)
	err := ctx.SetStorage(types.RepAddress.Bytes(), param.Account[:], data)
	if err != nil {
		t.Fatal(err)
	}

	height, _ := GetLastRepRewardHeightByAccount(ctx, param.Account)
	if height != param.EndHeight {
		t.Fatalf("get height err [%d-%d]", height, param.EndHeight)
	}
}

func TestRepCalcRewardEndHeight(t *testing.T) {
	startHeight := common.PovMinerRewardHeightStart
	maxEndHeight := common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay*10)
	expectHeight := common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay*7) - 1

	height := RepCalcRewardEndHeight(startHeight, maxEndHeight)
	if height != expectHeight {
		t.Fatalf("get end height err [%d-%d]", height, expectHeight)
	}
}

func TestRepRoundPovHeight(t *testing.T) {
	roundHeight := common.PovMinerRewardHeightRound
	height := common.PovMinerRewardHeightRound + 1
	expectHeight := common.PovMinerRewardHeightRound - 1

	h := RepRoundPovHeight(height, roundHeight)
	if h != expectHeight {
		t.Fatalf("get round height err [%d-%d]", h, expectHeight)
	}
}
