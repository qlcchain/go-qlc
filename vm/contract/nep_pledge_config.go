// +build !testnet

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
			pledgeTime:   &timeSpan{days: 90}, //90 days
			pledgeAmount: big.NewInt(2000 * 1e8),
		},
		cabi.Vote: {
			pledgeTime:   &timeSpan{days: 10}, //10 days
			pledgeAmount: big.NewInt(1 * 1e8),
		},
	}
)
