/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	l "github.com/qlcchain/go-qlc/ledger"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

var (
	minPledgeAmount      = big.NewInt(1e12)           // 10K QLC
	tokenNameLengthMax   = 40                         // Maximum length of a token name(include)
	tokenSymbolLengthMax = 10                         // Maximum length of a token symbol(include)
	minWithdrawTime      = time.Duration(24 * 30 * 3) // minWithdrawTime 3 months
)

type Mintage struct{}

//TODO: implement
func (m *Mintage) GetFee(ledger *l.Ledger, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (m *Mintage) DoSend(l *l.Ledger, block *types.StateBlock) error {
	param := new(cabi.ParamMintage)
	err := cabi.ABIMintage.UnpackMethod(param, cabi.MethodNameMintage, block.Data)
	if err != nil {
		return err
	}
	if err = verifyToken(*param); err != nil {
		return err
	}

	tokenId := cabi.NewTokenHash(block.Address, block.Previous, param.TokenName)
	if _, err = l.GetTokenById(types.Hash(tokenId)); err == nil {
		return fmt.Errorf("token Id[%s] already exist", tokenId.String())
	}

	if infos, err := l.ListTokens(); err == nil {
		for _, v := range infos {
			if v.TokenName == param.TokenName || v.TokenSymbol == param.TokenSymbol {
				return fmt.Errorf("invalid token name(%s) or token symbol(%s)", param.TokenName, param.TokenSymbol)
			}
		}
	}

	if block.Data, err = cabi.ABIMintage.PackMethod(
		cabi.MethodNameMintage,
		tokenId,
		param.TokenName,
		param.TokenSymbol,
		param.TotalSupply,
		param.Decimals); err != nil {
		return err
	}
	return nil
}

func verifyToken(param cabi.ParamMintage) error {
	if param.TotalSupply.Cmp(util.Tt256m1) > 0 ||
		param.TotalSupply.Cmp(new(big.Int).Exp(util.Big10, new(big.Int).SetUint64(uint64(param.Decimals)), nil)) < 0 ||
		len(param.TokenName) == 0 || len(param.TokenName) > tokenNameLengthMax ||
		len(param.TokenSymbol) == 0 || len(param.TokenSymbol) > tokenSymbolLengthMax {
		return errors.New("invalid token param")
	}
	if ok, _ := regexp.MatchString("^([a-zA-Z_]+[ ]?)*[a-zA-Z_]$", param.TokenName); !ok {
		return errors.New("invalid token name")
	}
	if ok, _ := regexp.MatchString("^([a-zA-Z_]+[ ]?)*[a-zA-Z_]$", param.TokenSymbol); !ok {
		return errors.New("invalid token symbol")
	}
	return nil
}

//TODO: verify input block timestamp
func (m *Mintage) DoReceive(ledger *l.Ledger, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.ParamMintage)
	_ = cabi.ABIMintage.UnpackMethod(param, cabi.MethodNameMintage, input.Data)
	var tokenInfo []byte
	amount, _ := ledger.CalculateAmount(input)
	if amount.Sign() > 0 &&
		amount.Compare(types.Balance{Int: minPledgeAmount}) != types.BalanceCompSmaller &&
		input.Token == common.QLCChainToken {
		var err error
		tokenInfo, err = cabi.ABIMintage.PackVariable(
			cabi.VariableNameToken,
			param.TokenId,
			param.TokenName,
			param.TokenSymbol,
			param.TotalSupply,
			param.Decimals,
			input.Address,
			amount.Int,
			time.Unix(input.Timestamp, 0).Add(minWithdrawTime).UTC().Unix())
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid block amount %d", amount.Int)
	}

	exp := new(big.Int).Exp(util.Big10, new(big.Int).SetUint64(uint64(param.Decimals)), nil)
	totalSupply := types.Balance{Int: new(big.Int).Mul(param.TotalSupply, exp)}

	block.Type = types.ContractReward
	block.Token = param.TokenId
	block.Link = input.Address.ToHash()
	block.Data = tokenInfo
	block.Previous = types.ZeroHash
	block.Balance = totalSupply
	block.Timestamp = time.Now().UTC().Unix()

	if _, err := ledger.GetStorage(types.MintageAddress[:], param.TokenId[:]); err == nil {
		//return nil, fmt.Errorf("invalid token")
	} else {
		//block.Data = tokenInfo
		if err := ledger.SetStorage(types.MintageAddress[:], param.TokenId[:], tokenInfo); err != nil {
			return nil, err
		}
	}

	return []*ContractBlock{
		{
			Block:     block,
			ToAddress: input.Address,
			BlockType: types.ContractReward,
			Amount:    totalSupply,
			Token:     param.TokenId,
			Data:      tokenInfo,
		},
	}, nil
}

func (m *Mintage) GetRefundData() []byte {
	return []byte{1}
}

func (m *Mintage) GetQuota() uint64 {
	return 0
}

type WithdrawMintage struct{}

func (m *WithdrawMintage) GetFee(ledger *l.Ledger, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (m *WithdrawMintage) DoSend(ledger *l.Ledger, block *types.StateBlock) error {
	if amount, err := ledger.CalculateAmount(block); block.Type != types.Send || amount.Compare(types.ZeroBalance) != types.BalanceCompEqual || err != nil {
		return errors.New("invalid block ")
	}
	tokenId := new(types.Hash)
	if err := cabi.ABIMintage.UnpackMethod(tokenId, cabi.MethodNameMintageWithdraw, block.Data); err != nil {
		return errors.New("invalid input data")
	}
	return nil
}

func (m *WithdrawMintage) DoReceive(ledger *l.Ledger, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	tokenId := new(types.Hash)
	_ = cabi.ABIMintage.UnpackMethod(tokenId, cabi.MethodNameMintageWithdraw, input.Data)
	ti, err := ledger.GetStorage(types.MintageAddress[:], tokenId[:])
	if err != nil {
		return nil, err
	}

	tokenInfo := new(types.TokenInfo)
	_ = cabi.ABIMintage.UnpackVariable(tokenInfo, cabi.VariableNameToken, ti)

	now := time.Now().UTC().Unix()

	if tokenInfo.Owner != input.Address ||
		tokenInfo.PledgeAmount.Sign() == 0 ||
		now > input.Timestamp {
		return nil, errors.New("cannot withdraw mintage pledge, status error")
	}

	newTokenInfo, _ := cabi.ABIMintage.PackVariable(
		cabi.VariableNameToken,
		tokenInfo.TokenId,
		tokenInfo.TokenName,
		tokenInfo.TokenSymbol,
		tokenInfo.TotalSupply,
		tokenInfo.Decimals,
		tokenInfo.Owner,
		big.NewInt(0),
		uint64(0))
	if err := ledger.SetStorage(types.MintageAddress[:], tokenId[:], newTokenInfo); err != nil {
		return nil, err
	}

	if tokenInfo.PledgeAmount.Sign() > 0 {
		return []*ContractBlock{
			{
				block,
				tokenInfo.Owner,
				types.ContractRefund,
				types.Balance{Int: tokenInfo.PledgeAmount},
				common.QLCChainToken,
				[]byte{},
			},
		}, nil
	}
	return nil, nil
}

func (m *WithdrawMintage) GetRefundData() []byte {
	return []byte{2}
}

func (m *WithdrawMintage) GetQuota() uint64 {
	return 0
}
