/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/abi"
	"math/big"
	"strings"
)

const (
	jsonVotePledge = `
	[
		{"type":"function","name":"VotePledge", "inputs":[{"name":"beneficial","type":"address"}]},
		{"type":"function","name":"WithdrawVotePledge","inputs":[{"name":"beneficial","type":"address"},{"name":"amount","type":"uint256"}]},
		{"type":"variable","name":"votePledgeInfo","inputs":[{"name":"amount","type":"uint256"},{"name":"withdrawTime","type":"int64"}]},
		{"type":"variable","name":"votePledgeBeneficial","inputs":[{"name":"amount","type":"uint256"}]}
	]`

	MethodVotePledge             = "VotePledge"
	MethodWithdrawVotePledge     = "WithdrawVotePledge"
	VariableVotePledgeInfo       = "votePledgeInfo"
	VariableVotePledgeBeneficial = "votePledgeBeneficial"
)

var (
	ABIPledge, _ = abi.JSONToABIContract(strings.NewReader(jsonVotePledge))
)

type VariablePledgeBeneficial struct {
	Amount *big.Int
}

type WithdrawPledgeParam struct {
	Beneficial types.Address
	Amount     *big.Int
}

type VotePledgeInfo struct {
	Amount            *big.Int
	WithdrawTime      int64
	BeneficialAddress types.Address
}

func GetPledgeBeneficialKey(beneficial types.Address) []byte {
	return beneficial.Bytes()
}

func GetPledgeKey(addr types.Address, pledgeBeneficialKey []byte, time int64) []byte {
	result := []byte(addr[:])
	tmp := util.Int2Bytes(time)

	result = append(result, pledgeBeneficialKey...)
	result = append(result, tmp...)
	return result
}

func GetPledgeKeyPrefix(addr types.Address, pledgeBeneficialKey []byte) []byte {
	return append(addr.Bytes(), pledgeBeneficialKey...)
}
