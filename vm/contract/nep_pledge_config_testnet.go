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

	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

var (
	minNetworkPledgeTime = 3  //  3 minutes
	minVotePledgeTime    = 10 //  10 minutes
	config               = map[cabi.PledgeType]pledgeInfo{
		cabi.Network: {
			pledgeTime:   time.Unix(0, 0).Add(time.Minute * time.Duration(minNetworkPledgeTime)).UTC().Unix(),
			pledgeAmount: big.NewInt(10),
		},
		cabi.Vote: {
			pledgeTime:   time.Unix(0, 0).Add(time.Minute * time.Duration(minVotePledgeTime)).UTC().Unix(),
			pledgeAmount: big.NewInt(1),
		},
	}
)
