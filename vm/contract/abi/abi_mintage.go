/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/abi"
	"math/big"
	"strings"
)

const (
	jsonMintage = `
	[
		{"type":"function","name":"Mintage","inputs":[{"name":"tokenId","type":"tokenId"},{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"}]},
		{"type":"function","name":"Withdraw","inputs":[{"name":"tokenId","type":"tokenId"}]},
		{"type":"variable","name":"token","inputs":[{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"owner","type":"address"},{"name":"pledgeAmount","type":"uint256"},{"name":"withdrawTime","type":"uint64"}]}
	]`

	MethodNameMintage         = "Mintage"
	MethodNameMintageWithdraw = "Withdraw"
	VariableNameToken         = "token"
)

var (
	ABIMintage, _ = abi.JSONToABIContract(strings.NewReader(jsonMintage))
)

type ParamMintage struct {
	Token       types.Hash
	TokenName   string
	TokenSymbol string
	TotalSupply *big.Int
	Decimals    uint8
}

func ParseTokenInfo(data []byte) (*types.TokenInfo, error) {
	if len(data) == 0 {
		return nil, errors.New("token info data is nil")
	}
	tokenInfo := new(types.TokenInfo)
	err := ABIMintage.UnpackVariable(tokenInfo, VariableNameToken, data)
	return tokenInfo, err
}

func NewTokenHash(address types.Address, previous types.Hash, tokenName string) types.Hash {
	h, _ := types.HashBytes(address[:], previous[:], util.String2Bytes(tokenName))
	return h
}

func GetStorageKey(key []byte) []byte {
	var tmp []byte
	tmp = append(tmp, types.MintageAddress[:]...)
	tmp = append(tmp, key...)

	return tmp
}
