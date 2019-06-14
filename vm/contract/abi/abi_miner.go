package abi

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/abi"
	"math/big"
	"strings"
)

const (
	jsonMiner = `
	[
		{"type":"function","name":"MinerReward","inputs":[{"name":"coinbase","type":"address"},{"name":"beneficial","type":"address"}]},
		{"type":"variable","name":"minerInfo","inputs":[{"name":"beneficial","type":"address"},{"name":"rewardHeight","type":"uint64"}]}
	]`

	MethodNameMinerReward = "MinerReward"
	VariableNameMiner     = "minerInfo"
)

var (
	MinerABI, _ = abi.JSONToABIContract(strings.NewReader(jsonMiner))

	// Reward per block, rewardPerBlock * blockNumPerYear / gasTotalSupply = 3%
	// 100000000 * 10e8 * 0.03 / (3600 * 24 * 30 * 365 / 30)
	RewardPerBlockInt     = big.NewInt(95129375)
	RewardPerBlockBalance = types.NewBalance(95129375)

	RewardTimeLimit   = int64(3600 * 24)
	RewardHeightLimit = uint64(3600 * 24 / 30)
)

type MinerRewardParam struct {
	Coinbase   types.Address
	Beneficial types.Address
}

type MinerInfo struct {
	Beneficial   types.Address
	RewardHeight uint64
}

func GetMinerKey(addr types.Address) []byte {
	result := []byte(nil)
	result = append(result, addr[:]...)
	return result
}
