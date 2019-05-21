/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"math/big"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/abi"
)

const (
	jsonRewards = `
	[
		{"type":"function","name":"AirdropRewards","inputs":[{"name":"header","type":"bytes32"},{"name":"amount","type":"uint256"},{"name":"NEP5TxId","type":"string"}]},
		{"type":"function","name":"ConfidantRewards","inputs":[{"name":"id","type":"string"},{"name":"header","type":"bytes32"},{"name":"amount","type":"uint256"}]},
		{"type":"variable","name":"rewardsInfo","inputs":[{"name":"header","type":"bytes32"},{"name":"amount","type":"uint256"},{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"type","type":"uint8"}]}	
	]`

	MethodNameAirdropRewards   = "AirdropRewards"
	MethodNameConfidantRewards = "ConfidantRewards"
	VariableNameRewards        = "rewardsInfo"
)

var (
	RewardsABI, _ = abi.JSONToABIContract(strings.NewReader(jsonRewards))
)

type ParamRewards struct {
	Header types.Hash
	Amount *big.Int
	From   *types.Address
	Type   uint8
}
