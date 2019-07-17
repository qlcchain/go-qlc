package abi

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"strings"
)

const (
	jsonMiner = `
	[
		{"type":"function","name":"MinerReward","inputs":[{"name":"coinbase","type":"address"},{"name":"beneficial","type":"address"},{"name":"rewardBlocks","type":"uint64"},{"name":"rewardHeight","type":"uint64"}]},
		{"type":"variable","name":"minerInfo","inputs":[{"name":"beneficial","type":"address"},{"name":"rewardHeight","type":"uint64"},{"name":"rewardBlocks","type":"uint64"}]}
	]`

	MethodNameMinerReward = "MinerReward"
	VariableNameMiner     = "minerInfo"
)

var (
	MinerABI, _ = abi.JSONToABIContract(strings.NewReader(jsonMiner))
)

type MinerRewardParam struct {
	Coinbase     types.Address `json:"coinbase"`
	Beneficial   types.Address `json:"beneficial"`
	RewardBlocks uint64        `json:"rewardBlocks"`
	RewardHeight uint64        `json:"rewardHeight"`
	//Signature    types.Signature `json:"signature"`
}

func (p *MinerRewardParam) Verify(address types.Address) (bool, error) {
	if !p.Coinbase.IsZero() && !p.Beneficial.IsZero() {
		return true, nil
		/*
			if data, err := MinerABI.PackMethod(MethodNameMinerReward, p.Coinbase, p.Beneficial, p.RewardBlocks, p.RewardHeight); err == nil {
				h := types.HashData(data)
				if address.Verify(h[:], p.Signature[:]) {
					return true, nil
				} else {
					return false, fmt.Errorf("invalid signature[%s] of hash[%s]", p.Signature.String(), h.String())
				}
			} else {
				return false, err
			}
		*/
	} else {
		return false, fmt.Errorf("invalid reward param")
	}
}

type MinerInfo struct {
	Beneficial   types.Address `json:"beneficial"`
	RewardBlocks uint64        `json:"rewardBlocks"`
	RewardHeight uint64        `json:"rewardHeight"`
}

func GetMinerKey(addr types.Address) []byte {
	result := []byte(nil)
	result = append(result, addr[:]...)
	return result
}

func GetMinerInfoByCoinbase(ctx *vmstore.VMContext, coinbase types.Address) (*MinerInfo, error) {
	key := GetMinerKey(coinbase)
	oldMinerData, err := ctx.GetStorage(types.MinerAddress.Bytes(), key)
	if err != nil {
		return nil, err
	}
	if len(oldMinerData) <= 0 {
		return nil, errors.New("miner data length is zero")
	}

	oldMinerInfo := new(MinerInfo)
	err = MinerABI.UnpackVariable(oldMinerInfo, VariableNameMiner, oldMinerData)
	if err != nil {
		return nil, errors.New("failed to unpack variable for miner")
	}

	return oldMinerInfo, nil
}

func MinerCheckRewardHeight(rewardHeight uint64) error {
	if rewardHeight < common.PovMinerRewardHeightStart {
		return fmt.Errorf("reward height %d should greater than or equal %d", rewardHeight, common.PovMinerRewardHeightStart)
	}

	if MinerRoundPovHeight(rewardHeight, uint64(common.PovMinerRewardHeightRound)) != rewardHeight {
		return fmt.Errorf("reward height plus one %d should be integral multiple of %d", rewardHeight+1, common.PovMinerRewardHeightRound)
	}

	return nil
}

func MinerCalcRewardEndHeight(startHeight uint64, maxEndHeight uint64) uint64 {
	if startHeight < common.PovMinerRewardHeightStart {
		startHeight = common.PovMinerRewardHeightStart
	}
	if maxEndHeight < common.PovMinerRewardHeightStart {
		maxEndHeight = common.PovMinerRewardHeightStart
	}

	endHeight := startHeight + common.PovMinerMaxRewardHeightPerCall - 1
	if endHeight > maxEndHeight {
		endHeight = maxEndHeight
	}
	return MinerRoundPovHeight(endHeight, uint64(common.PovMinerRewardHeightRound))
}

// height begin from 0, so height + 1 == blocks count
func MinerPovHeightToCount(height uint64) uint64 {
	return height + 1
}

func MinerRoundPovHeight(height uint64, round uint64) uint64 {
	roundCount := (MinerPovHeightToCount(height) / round) * round
	if roundCount == 0 {
		return roundCount
	}
	return roundCount - 1
}
