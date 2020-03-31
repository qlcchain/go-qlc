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
	"sync"

	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
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
	tokenCache    = &sync.Map{}
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

func ListTokens(ctx *vmstore.VMContext) ([]*types.TokenInfo, error) {
	logger := log.NewLogger("ListTokens")
	defer func() {
		logger.Sync()
	}()
	var infos []*types.TokenInfo
	if err := ctx.Iterator(contractaddress.MintageAddress[:], func(key []byte, value []byte) error {
		if len(value) > 0 {
			tokenId, _ := types.BytesToHash(key[(types.AddressSize + 1):])
			if config.IsGenesisToken(tokenId) {
				if info, err := ParseGenesisTokenInfo(value); err == nil {
					infos = append(infos, info)
				} else {
					logger.Error(err)
				}
			} else {
				if info, err := ParseTokenInfo(value); err == nil {
					exp := new(big.Int).Exp(util.Big10, new(big.Int).SetUint64(uint64(info.Decimals)), nil)
					info.TotalSupply = info.TotalSupply.Mul(info.TotalSupply, exp)
					infos = append(infos, info)
				} else {
					logger.Error(err)
				}
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
	if _, ok := tokenCache.Load(tokenId); !ok {
		if err := saveCache(ctx); err != nil {
			return nil, err
		}
	}
	if ti, ok := tokenCache.Load(tokenId); ok {
		return ti.(*types.TokenInfo), nil
	}

	return nil, fmt.Errorf("can not find token %s", tokenId.String())
}

func GetTokenByName(ctx *vmstore.VMContext, tokenName string) (*types.TokenInfo, error) {
	if _, ok := tokenCache.Load(tokenName); !ok {
		if err := saveCache(ctx); err != nil {
			return nil, err
		}
	}
	if ti, ok := tokenCache.Load(tokenName); ok {
		return ti.(*types.TokenInfo), nil
	}

	return nil, fmt.Errorf("can not find token %s", tokenName)
}

func saveCache(ctx *vmstore.VMContext) error {
	if infos, err := ListTokens(ctx); err == nil {
		for _, v := range infos {
			tokenCache.Store(v.TokenId, v)
			tokenCache.Store(v.TokenName, v)
		}
		return nil
	} else {
		return err
	}
}
