package abi

import (
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"strings"
)

const (
	jsonRep = `
	[
		{"type":"function","name":"RepReward","inputs":[{"name":"account","type":"address"},{"name":"beneficial","type":"address"},{"name":"startHeight","type":"uint64"},{"name":"endHeight","type":"uint64"},{"name":"rewardBlocks","type":"uint64"}]},
		{"type":"variable","name":"RepRewardInfo","inputs":[{"name":"beneficial","type":"address"},{"name":"startHeight","type":"uint64"},{"name":"endHeight","type":"uint64"},{"name":"rewardBlocks","type":"uint64"}]}
	]`

	MethodNameRepReward       = "RepReward"
	VariableNameRepRewardInfo = "RepRewardInfo"
)

var (
	RepABI, _ = abi.JSONToABIContract(strings.NewReader(jsonRep))
)

type RepRewardParam struct {
	Account      types.Address `json:"account"`
	Beneficial   types.Address `json:"beneficial"`
	StartHeight  uint64        `json:"startHeight"`
	EndHeight    uint64        `json:"endHeight"`
	RewardBlocks uint64        `json:"rewardBlocks"`
	RewardAmount types.Balance `json:"rewardAmount"`
}

func (p *RepRewardParam) Verify() (bool, error) {
	if p.Account.IsZero() {
		return false, fmt.Errorf("account is zero")
	}

	if p.Beneficial.IsZero() {
		return false, fmt.Errorf("beneficial is zero")
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
	RewardAmount types.Balance `json:"rewardAmount"`
}

func GetRepRewardKey(addr types.Address, height uint64) []byte {
	result := []byte(nil)
	result = append(result, addr[:]...)
	hb := util.BE_Uint64ToBytes(height)
	result = append(result, hb...)
	return result
}

func GetRepRewardInfosByAccount(ctx *vmstore.VMContext, account types.Address) ([]*RepRewardInfo, error) {
	var rewardInfos []*RepRewardInfo

	keyPrefix := []byte(nil)
	keyPrefix = append(keyPrefix, types.RepAddress[:]...)
	keyPrefix = append(keyPrefix, account[:]...)

	err := ctx.Iterator(keyPrefix, func(key []byte, value []byte) error {
		//ctx.GetLogger().Infof("key: %v, value: %v", key, value)
		if len(value) > 0 {
			info := new(RepRewardInfo)
			err := MinerABI.UnpackVariable(info, VariableNameRepRewardInfo, value)
			if err == nil {
				rewardInfos = append(rewardInfos, info)
			} else {
				ctx.GetLogger().Error(err)
			}
		}
		return nil
	})
	if err == nil {
		return rewardInfos, nil
	} else {
		return nil, err
	}
}

func CalcMaxRepRewardInfo(rewardInfos []*RepRewardInfo) *RepRewardInfo {
	var maxInfo *RepRewardInfo
	for _, rewardInfo := range rewardInfos {
		if maxInfo == nil {
			maxInfo = rewardInfo
		} else if rewardInfo.EndHeight > maxInfo.EndHeight {
			maxInfo = rewardInfo
		}
	}
	return maxInfo
}

func RepCalcRewardEndHeight(startHeight uint64, maxEndHeight uint64) uint64 {
	endHeight := startHeight + common.PovMinerMaxRewardBlocksPerCall - 1
	if endHeight > maxEndHeight {
		endHeight = maxEndHeight
	}

	endHeight = RepRoundPovHeight(endHeight, uint64(common.PovMinerRewardHeightRound))
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

func RepPovHeightToDayIndex(height uint64) uint32 {
	if height <= uint64(common.POVChainBlocksPerDay) {
		return 0
	}

	dayIndex := uint32(height / uint64(common.POVChainBlocksPerDay))

	remain := uint32(height % uint64(common.POVChainBlocksPerDay))
	if remain > 0 {
		dayIndex += 1
	}

	return dayIndex
}
