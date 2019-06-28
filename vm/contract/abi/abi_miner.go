package abi

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"math/big"
	"strings"
)

const (
	jsonMiner = `
	[
		{"type":"function","name":"MinerReward","inputs":[{"name":"coinbase","type":"address"},{"name":"beneficial","type":"address"},{"name":"rewardHeight","type":"uint64"}]},
		{"type":"variable","name":"minerInfo","inputs":[{"name":"beneficial","type":"address"},{"name":"rewardHeight","type":"uint64"},{"name":"rewardBlocks","type":"uint64"}]}
	]`

	MethodNameMinerReward = "MinerReward"
	VariableNameMiner     = "minerInfo"
)

var (
	MinerABI, _ = abi.JSONToABIContract(strings.NewReader(jsonMiner))

	// Reward per block, rewardPerBlock * blockNumPerYear / gasTotalSupply = 3%
	// 10000000000000000 * 0.03 / (3600 * 24 * 365 / 30)
	RewardPerBlockInt     = big.NewInt(285388127)
	RewardPerBlockBalance = types.NewBalance(285388127)

	RewardHeightGapToLatest = uint64(common.POVChainBlocksPerDay * 1)
	MaxRewardHeightPerCall  = uint64(common.POVChainBlocksPerDay * 90)
)

type MinerRewardParam struct {
	Coinbase     types.Address
	Beneficial   types.Address
	RewardHeight uint64
}

type MinerInfo struct {
	Beneficial   types.Address
	RewardBlocks uint64
	RewardHeight uint64
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

	if MinerRoundPovHeightByDay(rewardHeight) != rewardHeight {
		return fmt.Errorf("reward height plus one %d should be integral multiple of %d", rewardHeight+1, common.POVChainBlocksPerDay)
	}

	return nil
}

func MinerCalcRewardEndHeight(startHeight uint64, maxEndHeight uint64) uint64 {
	if startHeight < common.PovMinerRewardHeightStart {
		startHeight = common.PovMinerRewardHeightStart
	}
	endHeight := startHeight + MaxRewardHeightPerCall - 1
	if endHeight > maxEndHeight {
		endHeight = maxEndHeight
	}
	return MinerRoundPovHeightByDay(endHeight)
}

// height begin from 0, so height + 1 == blocks count
func MinerPovHeightToCount(height uint64) uint64 {
	return height + 1
}

func MinerRoundPovHeightByDay(height uint64) uint64 {
	roundCount := (MinerPovHeightToCount(height) / uint64(common.POVChainBlocksPerDay)) * uint64(common.POVChainBlocksPerDay)
	if roundCount == 0 {
		return roundCount
	}
	return roundCount - 1
}
