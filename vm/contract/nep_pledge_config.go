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
	"time"

	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

var (
	minNetworkPledgeTime = 3  // minMintageWithdrawTime 3 months
	minVotePledgeTime    = 10 // minMintageWithdrawTime 10 days
	config               = map[cabi.PledgeType]pledgeInfo{
		cabi.Network: {
			pledgeTime:   time.Unix(0, 0).AddDate(0, minNetworkPledgeTime, 0).UTC().Unix(),
			pledgeAmount: big.NewInt(2000),
		},
		cabi.Vote: {
			pledgeTime:   time.Unix(0, 0).AddDate(0, 0, minVotePledgeTime).UTC().Unix(),
			pledgeAmount: big.NewInt(1),
		},
	}
)
