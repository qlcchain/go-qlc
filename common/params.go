package common

import (
	"github.com/qlcchain/go-qlc/common/types"
	"math/big"
)

var (
	PovMinerPledgeAmountMin   = types.NewBalance(50000000000000)
	PovMinerVerifyHeightStart = uint64(3600 * 24 * 7 / 30)

	PovGenesisTargetHex = "0000007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	PovMinimumTargetHex = "0000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	PovMaximumTargetHex = "000003ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

	PovGenesisTargetInt, _ = new(big.Int).SetString(PovGenesisTargetHex, 16)
	PovMinimumTargetInt, _ = new(big.Int).SetString(PovMinimumTargetHex, 16)
	PovMaximumTargetInt, _ = new(big.Int).SetString(PovMaximumTargetHex, 16)

	// maximum number of seconds a block time is allowed to be ahead of the now time.
	PovMaxAllowedFutureTimeSec = 15
)
