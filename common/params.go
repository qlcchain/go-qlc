package common

import (
	"math/big"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	PovChainBlockInterval = 30
	PovChainTargetCycle   = 20
	PovChainBlockSize     = 4 * 1024 * 1024

	PovMinerPledgeAmountMin   = types.NewBalance(100000000000000)
	PovMinerVerifyHeightStart = uint64(3600 * 24 * 1 / PovChainBlockInterval)

	PoVMaxForkHeight = uint64(3600 * 24 * 7 / PovChainBlockInterval)

	PovGenesisTargetHex = "0000007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	PovMinimumTargetHex = "0000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	PovMaximumTargetHex = "000003ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

	PovGenesisTargetInt, _ = new(big.Int).SetString(PovGenesisTargetHex, 16)
	PovMinimumTargetInt, _ = new(big.Int).SetString(PovMinimumTargetHex, 16)
	PovMaximumTargetInt, _ = new(big.Int).SetString(PovMaximumTargetHex, 16)

	// maximum number of seconds a block time is allowed to be ahead of the now time.
	PovMaxAllowedFutureTimeSec = 15

	//vote right divisor
	VoteDivisor = int64(200)
)
