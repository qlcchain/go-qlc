/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"errors"
	"math/big"
	"regexp"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	l "github.com/qlcchain/go-qlc/ledger"
	cabi "github.com/qlcchain/go-qlc/vm/abi/contract"
)

var (
	tokenNameLengthMax   = 40                         // Maximum length of a token name(include)
	tokenSymbolLengthMax = 10                         // Maximum length of a token symbol(include)
	minWithdrawTime      = time.Duration(24 * 30 * 3) // minWithdrawTime 3 months
)

type Mintage struct{}

//TODO: implement
func (m *Mintage) GetFee(ledger *l.Ledger, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (m *Mintage) DoSend(ledger *l.Ledger, block *types.StateBlock) error {
	param := new(cabi.ParamMintage)
	err := cabi.ABIMintage.UnpackMethod(param, cabi.MethodNameMintage, block.Data)
	if err != nil {
		return err
	}
	if err = verifyToken(*param); err != nil {
		return err
	}

	tokenId := cabi.NewTokenHash(param)
	if _, err = ledger.GetTokenById(types.Hash(tokenId)); err != nil {
		return errors.New("invalid token Id")
	}

	if block.Data, err = cabi.ABIMintage.PackMethod(
		cabi.VariableNameToken,
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
	storage, err := ledger.GetStorage(&block.Address, param.Token[:])
	if err != nil {
		return nil, err
	}
	if len(storage) > 0 {
		return nil, errors.New("invalid token")
	}

	var tokenInfo []byte
	_, amount := ledger.CalculateAmount(input)
	if amount.Sign() == 0 {
		tokenInfo, _ = cabi.ABIMintage.PackVariable(
			cabi.VariableNameToken,
			param.TokenName,
			param.TokenSymbol,
			param.TotalSupply,
			param.Decimals,
			input.Address,
			amount.Int,
			time.Now().UTC().Unix())
	} else {
		withdrawTime := time.Now().UTC().Add(time.Hour * minWithdrawTime).Unix()
		tokenInfo, _ = cabi.ABIMintage.PackVariable(
			cabi.VariableNameToken,
			param.TokenName,
			param.TokenSymbol,
			param.TotalSupply,
			param.Decimals,
			input.Address,
			amount.Int,
			withdrawTime)
	}
	_ = ledger.SetStorage(param.Token[:], tokenInfo)
	return []*ContractBlock{
		{
			block,
			input.Address,
			types.ContractReward,
			types.Balance{Int: param.TotalSupply},
			param.Token,
			[]byte{},
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
	if isSend, amount := ledger.CalculateAmount(block); amount.Compare(types.ZeroBalance) != types.BalanceCompEqual || !isSend {
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
	tokenInfo := new(types.TokenInfo)
	storage, err := ledger.GetStorage(&block.Address, tokenId[:])
	if err != nil {
		return nil, err
	}
	_ = cabi.ABIMintage.UnpackVariable(tokenInfo, cabi.VariableNameToken, storage)

	now := time.Now().UTC().Unix()

	if tokenInfo.Owner != input.Address ||
		tokenInfo.PledgeAmount.Sign() == 0 ||
		now > input.Timestamp {
		return nil, errors.New("cannot withdraw mintage pledge, status error")
	}

	newTokenInfo, _ := cabi.ABIMintage.PackVariable(
		cabi.VariableNameToken,
		tokenInfo.TokenName,
		tokenInfo.TokenSymbol,
		tokenInfo.TotalSupply,
		tokenInfo.Decimals,
		tokenInfo.Owner,
		big.NewInt(0),
		uint64(0))
	storageKey := cabi.GetStorageKey(tokenId[:])

	if err := ledger.SetStorage(storageKey, newTokenInfo); err != nil {
		return nil, err
	}

	if tokenInfo.PledgeAmount.Sign() > 0 {
		return []*ContractBlock{
			{
				block,
				tokenInfo.Owner,
				types.ContractRefund,
				types.Balance{Int: tokenInfo.PledgeAmount},
				l.QLCChainToken,
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
