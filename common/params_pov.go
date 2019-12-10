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

	// Miner core parameters
	PovMinerPledgeAmountMin         = types.NewBalance(10000000000000)
	PovMinerVerifyHeightStart       = uint64(POVChainBlocksPerDay * 1)
	PovMinerRewardHeightStart       = uint64(POVChainBlocksPerDay * 3)
	PovMinerRewardHeightGapToLatest = uint64(POVChainBlocksPerDay * 1)
	PovMinerMaxRewardBlocksPerCall  = uint64(POVChainBlocksPerDay * 7)
	PovMinerRewardHeightRound       = uint64(POVChainBlocksPerDay * 1)
	PovMinerMaxFindNonceTimeSec     = PovChainBlockInterval * PovChainTargetCycle

	// Reward per block, rewardPerBlock * blockNumPerYear / gasTotalSupply = 3%
	// 10000000000000000 * 0.03 / (365 * 24 * 60)
	PovMinerRewardPerDay          = uint64(821917808219)
	PovMinerRewardPerBlock        = uint64(570776255)
	PovMinerRewardPerBlockInt     = big.NewInt(int64(PovMinerRewardPerBlock))
	PovMinerRewardPerBlockBalance = types.NewBalance(int64(PovMinerRewardPerBlock))
	PovMinerMinRewardPerBlock     = uint64(570776) // 0.001 * 570776255
	PovMinerRewardRatioMiner      = 60             // 60%
	PovMinerRewardRatioRep        = 40             // 40%

	// Bonus per block, bonusPerBlock * blockNumPerYear / stakingRewardPerYear = 90%(max)
	PovStakingRewardPerDay     = uint64(830200000000)
	PovMinerBonusMaxPerDay     = uint64(747180000000) // 830200000000 * 90 / 100
	PovMinerBonusPerKiloPerDay = uint64(747180)       // 747180000000 / (1G / 1M)
	PovMinerBonusDiffRatioMax  = uint64(1000000000)   // 1G
	PovMinerBonusDiffRatioMin  = 1000                 // 1K

	PoVMaxForkHeight = uint64(POVChainBlocksPerHour * 23)

	PovGenesisPowHex    = "00000000ffff0000000000000000000000000000000000000000000000000000"
	PovGenesisPowInt, _ = new(big.Int).SetString(PovGenesisPowHex, 16)
	PovGenesisPowBits   = types.BigToCompact(PovGenesisPowInt) //0x1d00ffff

	// PowLimit is the highest proof of work value a Bitcoin block
	// can have for the test network.
	PovPowLimitHex    = "00000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	PovPowLimitInt, _ = new(big.Int).SetString(PovPowLimitHex, 16)
	PovPowLimitBits   = types.BigToCompact(PovPowLimitInt) //0x1d00ffff

	// maximum number of seconds a block time is allowed to be ahead of the now time.
	PovMaxAllowedFutureTimeSec = 300

	PovMinCoinbaseExtraSize = 2
	PovMaxCoinbaseExtraSize = 100
)
