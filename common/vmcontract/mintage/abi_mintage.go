/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package mintage

import (
	"errors"
	"math/big"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/abi"
)

const (
	JsonMintage = `
	[
		{"type":"function","name":"Mintage","inputs":[{"name":"tokenId","type":"tokenId"},{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"beneficial","type":"address"},{"name":"NEP5TxId","type":"string"}]},
		{"type":"function","name":"Withdraw","inputs":[{"name":"tokenId","type":"tokenId"}]},
		{"type":"variable","name":"token","inputs":[{"name":"tokenId","type":"tokenId"},{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"owner","type":"address"},{"name":"pledgeAmount","type":"uint256"},{"name":"withdrawTime","type":"int64"},{"name":"pledgeAddress","type":"address"},{"name":"NEP5TxId","type":"string"}]},
		{"type":"variable","name":"genesisToken","inputs":[{"name":"tokenId","type":"tokenId"},{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"owner","type":"address"},{"name":"pledgeAmount","type":"uint256"},{"name":"withdrawTime","type":"int64"},{"name":"pledgeAddress","type":"address"}]}
	]`

	MethodNameMintage         = "Mintage"
	MethodNameMintageWithdraw = "Withdraw"
	VariableNameToken         = "token"
	VariableNameGenesisToken  = "genesisToken"
)

var (
	MintageABI, _ = abi.JSONToABIContract(strings.NewReader(JsonMintage))
)

type ParamMintage struct {
	TokenId     types.Hash
	TokenName   string
	TokenSymbol string
	TotalSupply *big.Int
	Decimals    uint8
	Beneficial  types.Address
	NEP5TxId    string
}

func ParseTokenInfo(data []byte) (*types.TokenInfo, error) {
	if len(data) == 0 {
		return nil, errors.New("token info data is nil")
	}
	tokenInfo := new(types.TokenInfo)
	if err := MintageABI.UnpackVariable(tokenInfo, VariableNameToken, data); err == nil {
		return tokenInfo, nil
	} else {
		return nil, err
	}
}

func ParseGenesisTokenInfo(data []byte) (*types.TokenInfo, error) {
	if len(data) == 0 {
		return nil, errors.New("token info data is nil")
	}
	genesisTokenInfo := new(types.TokenInfo)

	if err := MintageABI.UnpackVariable(genesisTokenInfo, VariableNameGenesisToken, data); err == nil {
		tokenInfo := &types.TokenInfo{
			TokenId:       genesisTokenInfo.TokenId,
			TokenName:     genesisTokenInfo.TokenName,
			TokenSymbol:   genesisTokenInfo.TokenSymbol,
			TotalSupply:   genesisTokenInfo.TotalSupply,
			Decimals:      genesisTokenInfo.Decimals,
			Owner:         genesisTokenInfo.Owner,
			PledgeAmount:  genesisTokenInfo.PledgeAmount,
			WithdrawTime:  genesisTokenInfo.WithdrawTime,
			PledgeAddress: genesisTokenInfo.PledgeAddress,
			NEP5TxId:      "",
		}
		return tokenInfo, nil
	} else {
		return nil, err
	}
}

func NewTokenHash(address types.Address, previous types.Hash, tokenName string) types.Hash {
	h, _ := types.HashBytes(address[:], previous[:], util.String2Bytes(tokenName))
	return h
}
