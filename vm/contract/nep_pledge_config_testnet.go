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

	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

var (
	config = map[cabi.PledgeType]pledgeInfo{
		cabi.Network: {
			pledgeTime:   &timeSpan{minutes: 3}, // 3 minutes
			pledgeAmount: big.NewInt(10 * 1e8),
		},
		cabi.Vote: {
			pledgeTime:   &timeSpan{minutes: 10}, //  10 minutes
			pledgeAmount: big.NewInt(1 * 1e8),
		},
		cabi.Oracle: {
			pledgeTime:   &timeSpan{minutes: 1},     // 1 minutes
			pledgeAmount: big.NewInt(3000000 * 1e8), // 3M
		},
	}
)
