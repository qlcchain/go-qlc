/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/qlcchain/go-qlc/log"

	"github.com/qlcchain/go-qlc/vm/vmstore"

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
	MintageABI, _ = abi.JSONToABIContract(strings.NewReader(jsonMintage))
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
	if err := MintageABI.UnpackVariable(tokenInfo, VariableNameToken, data); err == nil {
		return tokenInfo, nil
	} else {
		return nil, err
	}
}

func NewTokenHash(address types.Address, previous types.Hash, tokenName string) types.Hash {
	h, _ := types.HashBytes(address[:], previous[:], util.String2Bytes(tokenName))
	return h
}

func ListTokens(ctx *vmstore.VMContext) ([]*types.TokenInfo, error) {
	logger := log.NewLogger("ListTokens")
	defer func() {
		logger.Sync()
	}()
	var infos []*types.TokenInfo
	if err := ctx.Iterator(types.MintageAddress[:], func(key []byte, value []byte) error {
		if len(value) > 0 {
			if info, err := ParseTokenInfo(value); err == nil {
				infos = append(infos, info)
			} else {
				logger.Error(err)
			}
		}
		return nil
	}); err == nil {
		return infos, nil
	} else {
		return nil, err
	}
}

func GetTokenById(ctx *vmstore.VMContext, tokenId types.Hash) (*types.TokenInfo, error) {
	if infos, err := ListTokens(ctx); err == nil {
		for _, v := range infos {
			if v.TokenId == tokenId {
				return v, nil
			}
		}
	} else {
		return nil, err
	}

	return nil, fmt.Errorf("can not find token %s", tokenId.String())
}

func GetTokenByName(ctx *vmstore.VMContext, tokenName string) (*types.TokenInfo, error) {
	if infos, err := ListTokens(ctx); err == nil {
		for _, v := range infos {
			if v.TokenName == tokenName {
				return v, nil
			}
		}
	} else {
		return nil, err
	}

	return nil, fmt.Errorf("can not find token %s", tokenName)
}
