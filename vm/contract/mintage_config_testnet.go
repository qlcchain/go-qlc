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
)

var (
	MinPledgeAmount      = big.NewInt(10 * 1e8)   // 10 QLC
	tokenNameLengthMax   = 40                     // Maximum length of a token name(include)
	tokenSymbolLengthMax = 10                     // Maximum length of a token symbol(include)
	minMintageTime       = &timeSpan{minutes: 10} //10 minutes
)
