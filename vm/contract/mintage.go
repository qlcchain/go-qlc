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
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/common/vmcontract/mintage"
	mintage2 "github.com/qlcchain/go-qlc/common/vmcontract/mintage"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type Mintage struct {
	BaseContract
}

var MintageContract = NewChainContract(
	map[string]Contract{
		mintage2.MethodNameMintage: &Mintage{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer1,
					signature: true,
					work:      true,
				},
			},
		},
		mintage2.MethodNameMintageWithdraw: &WithdrawMintage{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer1,
					signature: true,
					work:      true,
				},
			},
		},
	},
	mintage2.MintageABI,
	mintage2.JsonMintage,
)

func (m *Mintage) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	param := new(mintage.ParamMintage)
	err := mintage.MintageABI.UnpackMethod(param, mintage.MethodNameMintage, block.Data)
	if err != nil {
		return err
	}
	if err = verifyToken(*param); err != nil {
		return err
	}

	tokenId := mintage.NewTokenHash(block.Address, block.Previous, param.TokenName)
	if _, err = ctx.GetTokenById(tokenId); err == nil {
		return fmt.Errorf("token Id[%s] already exist", tokenId.String())
	}

	if infos, err := ctx.ListTokens(); err == nil {
		for _, v := range infos {
			if v.TokenName == param.TokenName || v.TokenSymbol == param.TokenSymbol {
				return fmt.Errorf("invalid token name(%s) or token symbol(%s)", param.TokenName, param.TokenSymbol)
			}
		}
	}

	if block.Data, err = mintage.MintageABI.PackMethod(
		mintage.MethodNameMintage,
		tokenId,
		param.TokenName,
		param.TokenSymbol,
		param.TotalSupply,
		param.Decimals,
		param.Beneficial,
		param.NEP5TxId); err != nil {
		return err
	}
	return nil
}

func verifyToken(param mintage.ParamMintage) error {
	if param.TotalSupply.Cmp(util.Tt256m1) > 0 ||
		//param.TotalSupply.Cmp(new(big.Int).Exp(util.Big10, new(big.Int).SetUint64(uint64(param.Decimals)), nil)) < 0 ||
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

func (m *Mintage) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(mintage.ParamMintage)
	err := mintage.MintageABI.UnpackMethod(param, mintage.MethodNameMintage, block.Data)
	if err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: param.Beneficial,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Amount: types.Balance{Int: param.TotalSupply},
			Type:   block.Token,
		}, nil
}

func (m *Mintage) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(mintage.ParamMintage)
	_ = mintage.MintageABI.UnpackMethod(param, mintage.MethodNameMintage, input.Data)
	var tokenInfo []byte
	amount, _ := ctx.CalculateAmount(input)
	if amount.Sign() > 0 &&
		amount.Compare(types.Balance{Int: MinPledgeAmount}) != types.BalanceCompSmaller &&
		input.Token == cfg.ChainToken() {
		var err error
		tokenInfo, err = mintage.MintageABI.PackVariable(
			mintage.VariableNameToken,
			param.TokenId,
			param.TokenName,
			param.TokenSymbol,
			param.TotalSupply,
			param.Decimals,
			param.Beneficial,
			amount.Int,
			minMintageTime.Calculate(time.Unix(input.Timestamp, 0)).Unix(),
			input.Address,
			param.NEP5TxId)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid block amount %d", amount.Int)
	}

	if _, err := ctx.GetStorage(contractaddress.MintageAddress[:], []byte(param.NEP5TxId)); err == nil {
		return nil, fmt.Errorf("invalid mintage nep5 tx id %s", param.NEP5TxId)
	} else {
		if err := ctx.SetStorage(contractaddress.MintageAddress[:], []byte(param.NEP5TxId), nil); err != nil {
			return nil, err
		}
	}

	exp := new(big.Int).Exp(util.Big10, new(big.Int).SetUint64(uint64(param.Decimals)), nil)
	totalSupply := types.Balance{Int: new(big.Int).Mul(param.TotalSupply, exp)}

	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Representative = param.Beneficial
	block.Token = param.TokenId
	block.Link = input.GetHash()
	block.Data = tokenInfo
	block.Previous = types.ZeroHash
	block.Balance = totalSupply
	//block.Vote = types.ZeroBalance
	//block.Storage = types.ZeroBalance
	//block.Network = types.ZeroBalance
	//block.Oracle = types.ZeroBalance

	if _, err := ctx.GetStorage(contractaddress.MintageAddress[:], param.TokenId[:]); err == nil {
		return nil, fmt.Errorf("invalid token")
	} else {
		if err := ctx.SetStorage(contractaddress.MintageAddress[:], param.TokenId[:], tokenInfo); err != nil {
			return nil, err
		}
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: param.Beneficial,
			BlockType: types.ContractReward,
			Amount:    totalSupply,
			Token:     param.TokenId,
			Data:      tokenInfo,
		},
	}, nil
}

func (m *Mintage) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	data := block.GetData()
	tr := types.ZeroAddress

	if method, err := mintage.MintageABI.MethodById(data[0:4]); err == nil {
		if method.Name == mintage.MethodNameMintage {
			param := new(mintage.ParamMintage)
			if err = method.Inputs.Unpack(param, data[4:]); err == nil {
				tr = param.Beneficial
				return tr, nil
			} else {
				return tr, err
			}
		} else {
			return tr, errors.New("method err")
		}
	} else {
		return tr, err
	}
}

type WithdrawMintage struct {
	BaseContract
}

func (m *WithdrawMintage) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	if amount, err := ctx.CalculateAmount(block); block.Type != types.ContractSend || err != nil ||
		amount.Compare(types.ZeroBalance) != types.BalanceCompEqual {
		return errors.New("invalid block ")
	}
	tokenId := new(types.Hash)
	if err := mintage.MintageABI.UnpackMethod(tokenId, mintage.MethodNameMintageWithdraw, block.Data); err != nil {
		return errors.New("invalid input data")
	}
	return nil
}

func (m *WithdrawMintage) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	tokenId := new(types.Hash)
	err := mintage.MintageABI.UnpackMethod(tokenId, mintage.MethodNameMintageWithdraw, block.Data)

	if err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: block.Address,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Amount: types.ZeroBalance,
			Type:   block.Token,
		}, nil
}

func (m *WithdrawMintage) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	tokenId := new(types.Hash)
	err := mintage.MintageABI.UnpackMethod(tokenId, mintage.MethodNameMintageWithdraw, input.Data)
	if err != nil {
		return nil, err
	}
	tokenInfoData, err := ctx.GetStorage(contractaddress.MintageAddress[:], tokenId[:])
	if err != nil {
		return nil, err
	}

	tokenInfo := new(types.TokenInfo)
	err = mintage.MintageABI.UnpackVariable(tokenInfo, mintage.VariableNameToken, tokenInfoData)
	if err != nil {
		return nil, err
	}

	now := common.TimeNow().Unix()
	if tokenInfo.PledgeAddress != input.Address ||
		tokenInfo.PledgeAmount.Sign() == 0 ||
		now < tokenInfo.WithdrawTime {
		return nil, errors.New("cannot withdraw mintage pledge, status error")
	}

	newTokenInfo, err := mintage.MintageABI.PackVariable(
		mintage.VariableNameToken,
		tokenInfo.TokenId,
		tokenInfo.TokenName,
		tokenInfo.TokenSymbol,
		tokenInfo.TotalSupply,
		tokenInfo.Decimals,
		tokenInfo.Owner,
		big.NewInt(0),
		int64(0),
		tokenInfo.PledgeAddress,
		tokenInfo.NEP5TxId)

	if err != nil {
		return nil, err
	}
	var tm *types.TokenMeta
	am, _ := ctx.GetAccountMeta(tokenInfo.PledgeAddress)
	if am != nil {
		tm = am.Token(cfg.ChainToken())
		if tm == nil {
			return nil, fmt.Errorf("chain token %s do not found", cfg.ChainToken().String())
		}
	} else {
		return nil, errors.New("accountMeta not found")
	}

	block.Type = types.ContractReward
	block.Address = tokenInfo.PledgeAddress
	block.Representative = tm.Representative
	block.Token = tm.Type
	block.Link = input.GetHash()
	block.Data = newTokenInfo
	block.Balance = tm.Balance.Add(types.Balance{Int: tokenInfo.PledgeAmount})
	//block.Vote = am.CoinVote
	//block.Oracle = am.CoinOracle
	//block.Storage = am.CoinStorage
	//block.Network = am.CoinNetwork
	block.Previous = tm.Header

	var pledgeData []byte
	if pledgeData, err = ctx.GetStorage(contractaddress.MintageAddress[:], tokenId[:]); err != nil && err != vmstore.ErrStorageNotFound {
		return nil, err
	} else {
		// already exist,verify data
		if len(pledgeData) > 0 {
			oldPledge := new(types.TokenInfo)
			err := mintage.MintageABI.UnpackVariable(oldPledge, mintage.VariableNameToken, pledgeData)
			if err != nil {
				return nil, err
			}
			if oldPledge.PledgeAddress != tokenInfo.PledgeAddress || oldPledge.WithdrawTime != tokenInfo.WithdrawTime ||
				oldPledge.TokenId != tokenInfo.TokenId || oldPledge.Owner != tokenInfo.Owner || oldPledge.Decimals != tokenInfo.Decimals ||
				oldPledge.TotalSupply.String() != tokenInfo.TotalSupply.String() || oldPledge.TokenSymbol != tokenInfo.TokenSymbol ||
				oldPledge.TokenName != tokenInfo.TokenName || oldPledge.PledgeAmount.String() != tokenInfo.PledgeAmount.String() {
				return nil, errors.New("invalid saved mine info")
			}
			if err := ctx.SetStorage(contractaddress.MintageAddress[:], tokenId[:], newTokenInfo); err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("invalid saved mine data")
		}
	}

	if tokenInfo.PledgeAmount.Sign() > 0 {
		return []*ContractBlock{
			{
				VMContext: ctx,
				Block:     block,
				ToAddress: tokenInfo.PledgeAddress,
				BlockType: types.ContractReward,
				Amount:    types.Balance{Int: tokenInfo.PledgeAmount},
				Token:     cfg.ChainToken(),
				Data:      newTokenInfo,
			},
		}, nil
	}
	return nil, nil
}

// SetMinMintageTime only for test
func SetMinMintageTime(y, m, d, h, i, s int) {
	minMintageTime = &timeSpan{
		years:   y,
		months:  m,
		days:    d,
		hours:   h,
		minutes: i,
		seconds: s,
	}
}
