/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"sort"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestSortNEP5PledgeInfo(t *testing.T) {
	infos := []*NEP5PledgeInfo{
		{
			PType:         0,
			Amount:        nil,
			WithdrawTime:  1111,
			Beneficial:    types.Address{},
			PledgeAddress: types.Address{},
		},
		{
			PType:         0,
			Amount:        nil,
			WithdrawTime:  222,
			Beneficial:    types.Address{},
			PledgeAddress: types.Address{},
		},
		{
			PType:         0,
			Amount:        nil,
			WithdrawTime:  32,
			Beneficial:    types.Address{},
			PledgeAddress: types.Address{},
		},
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].WithdrawTime < infos[j].WithdrawTime
	})

	for _, v := range infos {
		t.Log(v)
	}
}
