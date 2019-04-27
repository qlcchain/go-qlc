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
			pledgeAmount: big.NewInt(10),
		},
		cabi.Vote: {
			pledgeTime:   &timeSpan{minutes: 10}, //  10 minutes
			pledgeAmount: big.NewInt(1),
		},
	}
)
