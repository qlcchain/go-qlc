package common

import "github.com/qlcchain/go-qlc/common/types"

var (
	PovMinerPledgeAmountMin   = types.NewBalance(50000000000000)
	PovMinerVerifyHeightStart = uint64(3600 * 24 * 7 / 30)
)
