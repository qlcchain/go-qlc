/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"errors"
	"math/big"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/abi"
)

const (
	jsonMintage = `
	[
		{"type":"function","name":"Mintage","inputs":[{"name":"tokenId","type":"tokenId"},{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"beneficial","type":"address"}]},
		{"type":"function","name":"Withdraw","inputs":[{"name":"tokenId","type":"tokenId"}]},
		{"type":"variable","name":"token","inputs":[{"name":"tokenId","type":"tokenId"},{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"owner","type":"address"},{"name":"pledgeAmount","type":"uint256"},{"name":"withdrawTime","type":"int64"},{"name":"pledgeAddress","type":"address"}]}
	]`

	MethodNameMintage         = "Mintage"
	MethodNameMintageWithdraw = "Withdraw"
	VariableNameToken         = "token"
)

var (
	ABIMintage, _ = abi.JSONToABIContract(strings.NewReader(jsonMintage))
)

type ParamMintage struct {
	TokenId     types.Hash
	TokenName   string
	TokenSymbol string
	TotalSupply *big.Int
	Decimals    uint8
	Beneficial  types.Address
}

func ParseTokenInfo(data []byte) (*types.TokenInfo, error) {
	if len(data) == 0 {
		return nil, errors.New("token info data is nil")
	}
	tokenInfo := new(types.TokenInfo)
	err := ABIMintage.UnpackVariable(tokenInfo, VariableNameToken, data)
	if err == nil && tokenInfo.PledgeAddress.IsZero() {
		tokenInfo.PledgeAddress = tokenInfo.Owner
	}
	return tokenInfo, err
}

func NewTokenHash(address types.Address, previous types.Hash, tokenName string) types.Hash {
	h, _ := types.HashBytes(address[:], previous[:], util.String2Bytes(tokenName))
	return h
}
