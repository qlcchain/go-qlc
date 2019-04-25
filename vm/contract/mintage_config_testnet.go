// +build testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"math/big"
	"time"
)

var (
	MinPledgeAmount        = big.NewInt(1 * 1e12)                                              // 10K QLC
	tokenNameLengthMax     = 40                                                                // Maximum length of a token name(include)
	tokenSymbolLengthMax   = 10                                                                // Maximum length of a token symbol(include)
	minMintageWithdrawTime = time.Unix(0, 0).Add(time.Minute * time.Duration(10)).UTC().Unix() //10 minutes
)
