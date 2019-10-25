package abi

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	JsonRep = `
	[
		{"type":"function","name":"RepReward","inputs":[
			{"name":"account","type":"address"},
			{"name":"beneficial","type":"address"},
			{"name":"startHeight","type":"uint64"},
			{"name":"endHeight","type":"uint64"},
			{"name":"rewardBlocks","type":"uint64"},
			{"name":"rewardAmount","type":"uint256"}
		]},
		{"type":"variable","name":"RepRewardInfo","inputs":[
			{"name":"endHeight","type":"uint64"},
			{"name":"rewardBlocks","type":"uint64"},
			{"name":"timestamp","type":"int64"},
			{"name":"rewardAmount","type":"uint256"}
		]}
	]`

	MethodNameRepReward   = "RepReward"
	VariableNameRepReward = "RepRewardInfo"
)

var (
	RepABI, _ = abi.JSONToABIContract(strings.NewReader(JsonRep))
)

type RepRewardParam struct {
	Account      types.Address `json:"account"`
	Beneficial   types.Address `json:"beneficial"`
	StartHeight  uint64        `json:"startHeight"`
	EndHeight    uint64        `json:"endHeight"`
	RewardBlocks uint64        `json:"rewardBlocks"`
	RewardAmount *big.Int      `json:"rewardAmount"`
}

func (p *RepRewardParam) Verify() (bool, error) {
	if p.Account.IsZero() {
		return false, fmt.Errorf("account is zero")
	}

	if p.Beneficial.IsZero() {
		return false, fmt.Errorf("beneficial is zero")
	}

	if p.StartHeight < common.PovMinerRewardHeightStart {
		return false, fmt.Errorf("startHeight %d less than %d", p.StartHeight, common.PovMinerRewardHeightStart)
	}

	if p.StartHeight > p.EndHeight {
		return false, fmt.Errorf("startHeight %d greater than endHeight %d", p.StartHeight, p.EndHeight)
	}

	gapCount := p.EndHeight - p.StartHeight + 1
	if gapCount > common.PovMinerMaxRewardBlocksPerCall {
		return false, fmt.Errorf("gap count %d exceed max blocks %d", p.StartHeight, common.PovMinerMaxRewardBlocksPerCall)
	}

	return true, nil
}

type RepRewardInfo struct {
	Beneficial   types.Address `json:"beneficial"`
	StartHeight  uint64        `json:"startHeight"`
	EndHeight    uint64        `json:"endHeight"`
	RewardBlocks uint64        `json:"rewardBlocks"`
	RewardAmount *big.Int      `json:"rewardAmount"`
	Timestamp    int64         `json:"_"`
}

func GetLastRepRewardHeightByAccount(ctx *vmstore.VMContext, account types.Address) (uint64, error) {
	data, err := ctx.GetStorage(types.RepAddress[:], account[:])
	if err == nil {
		info := new(RepRewardInfo)
		er := RepABI.UnpackVariable(info, VariableNameRepReward, data)
		if er != nil {
			return 0, er
		}
		return info.EndHeight, nil
	} else {
		return 0, err
	}
}

func RepCalcRewardEndHeight(startHeight uint64, maxEndHeight uint64) uint64 {
	if maxEndHeight < common.PovMinerRewardHeightStart {
		return 0
	}

	if startHeight < common.PovMinerRewardHeightStart {
		startHeight = common.PovMinerRewardHeightStart
	}

	endHeight := startHeight + common.PovMinerMaxRewardBlocksPerCall - 1
	if endHeight > maxEndHeight {
		endHeight = maxEndHeight
	}

	endHeight = MinerRoundPovHeight(endHeight, common.PovMinerRewardHeightRound)
	if endHeight < common.PovMinerRewardHeightStart {
		return 0
	}

	return endHeight
}

// height begin from 0, so height + 1 == blocks count
func RepPovHeightToCount(height uint64) uint64 {
	return height + 1
}

func RepRoundPovHeight(height uint64, round uint64) uint64 {
	roundCount := (RepPovHeightToCount(height) / round) * round
	if roundCount == 0 {
		return 0
	}
	return roundCount - 1
}
