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
	PovMinerMaxFindNonceTimeSec     = PovChainBlockInterval * PovChainTargetCycle

	// Reward per block, rewardPerBlock * blockNumPerYear / gasTotalSupply = 3%
	// 10000000000000000 * 0.03 / (3600 * 24 * 365 / 30)
	PovMinerRewardPerBlock        = 285388127
	PovMinerRewardPerBlockInt     = big.NewInt(int64(PovMinerRewardPerBlock))
	PovMinerRewardPerBlockBalance = types.NewBalance(int64(PovMinerRewardPerBlock))
	PovMinerMinRewardPerBlock     = 285388 // 0.001 * 285388127

	PoVMaxForkHeight = uint64(POVChainBlocksPerHour * 12)

	PovGenesisPowHex    = "00000ffff0000000000000000000000000000000000000000000000000000000"
	PovGenesisPowInt, _ = new(big.Int).SetString(PovGenesisPowHex, 16)
	PovGenesisPowBits   = types.BigToCompact(PovGenesisPowInt)

	// PowLimit is the highest proof of work value a Bitcoin block
	// can have for the test network.
	PovPowLimitHex    = "00000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	PovPowLimitInt, _ = new(big.Int).SetString(PovPowLimitHex, 16)
	PovPowLimitBits   = types.BigToCompact(PovPowLimitInt)

	// maximum number of seconds a block time is allowed to be ahead of the now time.
	PovMaxAllowedFutureTimeSec = 300

	PovMaxNonce = ^uint32(0) // 2^32 - 1
)
