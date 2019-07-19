// +build !testnet

package common

import (
	"math/big"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	// PoV Block Chain Params
	PovChainGenesisBlockHeight = uint64(0)

	PovChainBlockInterval = 60
	PovChainTargetCycle   = 20
	PovChainBlockSize     = 2 * 1024 * 1024

	PovChainRetargetTimespan    = PovChainBlockInterval * PovChainTargetCycle
	PovChainMinRetargetTimespan = PovChainRetargetTimespan / 4
	PovChainMaxRetargetTimespan = PovChainRetargetTimespan * 4

	POVChainBlocksPerHour = 3600 / PovChainBlockInterval
	POVChainBlocksPerDay  = POVChainBlocksPerHour * 24

	PovMinerPledgeAmountMin         = types.NewBalance(100000000000000)
	PovMinerVerifyHeightStart       = uint64(POVChainBlocksPerDay * 1)
	PovMinerRewardHeightStart       = uint64(POVChainBlocksPerDay * 30)
	PovMinerRewardHeightGapToLatest = uint64(POVChainBlocksPerDay * 1)
	PovMinerMaxRewardBlocksPerCall  = uint64(POVChainBlocksPerDay * 7)
	PovMinerRewardHeightRound       = uint64(POVChainBlocksPerDay * 1)

	// Reward per block, rewardPerBlock * blockNumPerYear / gasTotalSupply = 3%
	// 10000000000000000 * 0.03 / (3600 * 24 * 365 / 30)
	PovMinerRewardPerBlockInt     = big.NewInt(285388127)
	PovMinerRewardPerBlockBalance = types.NewBalance(285388127)

	PoVMaxForkHeight = uint64(POVChainBlocksPerHour * 12)

	PovGenesisTargetHex = "0000007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	//PovMinimumTargetHex = "0000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	//PovMaximumTargetHex = "000003ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

	PovGenesisTargetInt, _ = new(big.Int).SetString(PovGenesisTargetHex, 16)
	//PovMinimumTargetInt, _ = new(big.Int).SetString(PovMinimumTargetHex, 16)
	//PovMaximumTargetInt, _ = new(big.Int).SetString(PovMaximumTargetHex, 16)

	// maximum number of seconds a block time is allowed to be ahead of the now time.
	PovMaxAllowedFutureTimeSec = 15

	PovMaxNonce = ^uint64(0) // 2^64 - 1
)
