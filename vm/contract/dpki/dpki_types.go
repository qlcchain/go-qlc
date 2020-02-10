package dpki

import (
	"math/big"

	"github.com/qlcchain/go-qlc/common/types"
)

type PKDRewardParam struct {
	Account      types.Address `json:"account"`
	Beneficial   types.Address `json:"beneficial"`
	EndHeight    uint64        `json:"endHeight"`
	RewardAmount *big.Int      `json:"rewardAmount"`
}

func NewPKDRewardParam() *PKDRewardParam {
	i := new(PKDRewardParam)
	i.RewardAmount = big.NewInt(0)
	return i
}

type PKDRewardInfo struct {
	Beneficial   types.Address `json:"beneficial"`
	EndHeight    uint64        `json:"endHeight"`
	RewardAmount *big.Int      `json:"rewardAmount"`
	Timestamp    int64         `json:"timestamp"`
}

func NewPKDRewardInfo() *PKDRewardInfo {
	i := new(PKDRewardInfo)
	i.RewardAmount = big.NewInt(0)
	return i
}
