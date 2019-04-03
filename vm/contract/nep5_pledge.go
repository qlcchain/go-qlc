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
	"sort"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

var (
	minPledgeTime = time.Duration(24 * 30 * 6) // minWithdrawTime 6 months)
	config        = map[cabi.PledgeType]pledgeInfo{
		cabi.Netowrk: {
			pledgeTime:   minPledgeTime,
			pledgeAmount: big.NewInt(2300),
		},
		cabi.Vote: {
			pledgeTime:   minPledgeTime,
			pledgeAmount: big.NewInt(1),
		},
	}
)

type pledgeInfo struct {
	pledgeTime   time.Duration
	pledgeAmount *big.Int
}

type Nep5Pledge struct {
}

func (p *Nep5Pledge) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

// check pledge chain coin
// - address is normal user address
// - small than min pledge amount
// transfer quota to beneficial address
func (*Nep5Pledge) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	// check pledge amount
	amount, err := ctx.CalculateAmount(block)
	if err != nil {
		return err
	}

	// check send account is user account
	b, err := ctx.IsUserAccount(block.Address)
	if err != nil {
		return err
	}

	if block.Token != common.ChainToken() || amount.IsZero() || !b {
		return errors.New("invalid block data")
	}

	param := new(cabi.PledgeParam)
	if err := cabi.NEP5PledgeABI.UnpackMethod(param, cabi.MethodNEP5Pledge, block.Data); err != nil {
		return errors.New("invalid beneficial address")
	}

	if info, b := config[param.PType]; !b {
		return fmt.Errorf("unsupport type %s", param.PType.String())
	} else if amount.Compare(types.Balance{Int: info.pledgeAmount}) == types.BalanceCompSmaller {
		return fmt.Errorf("not enough pledge amount %s, expect %s", amount.String(), info.pledgeAmount)
	}

	if param.PledgeAddress != block.Address {
		return fmt.Errorf("invalid pledge address[%s],expect %s",
			param.PledgeAddress.String(), block.Address.String())
	}

	block.Data, err = cabi.NEP5PledgeABI.PackMethod(cabi.MethodNEP5Pledge, param.Beneficial,
		param.PledgeAddress, uint8(param.PType))
	if err != nil {
		return err
	}
	return nil
}

func (*Nep5Pledge) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.PledgeParam)
	_ = cabi.NEP5PledgeABI.UnpackMethod(param, cabi.MethodNEP5Pledge, input.Data)
	amount, _ := ctx.CalculateAmount(input)

	var withdrawTime int64
	if info, b := config[param.PType]; b {
		withdrawTime = time.Unix(input.Timestamp, 0).Add(info.pledgeTime).UTC().Unix()
	} else {
		return nil, fmt.Errorf("unsupport type %s", param.PType.String())
	}

	info := cabi.NEP5PledgeInfo{
		PType:         uint8(param.PType),
		Amount:        amount.Int,
		WithdrawTime:  withdrawTime,
		Beneficial:    param.Beneficial,
		PledgeAddress: param.PledgeAddress,
	}

	pledgeKey := cabi.GetPledgeKey(input.Address, param.Beneficial, input.Timestamp)

	if oldPledgeData, err := ctx.GetStorage(types.NEP5PledgeAddress[:], pledgeKey); err != nil && err != vmstore.ErrStorageNotFound {
		return nil, err
	} else {
		if len(oldPledgeData) > 0 {
			oldPledge := new(cabi.NEP5PledgeInfo)
			err := cabi.NEP5PledgeABI.UnpackVariable(oldPledge, cabi.VariableNEP5PledgeInfo, oldPledgeData)
			if err != nil {
				return nil, err
			}
			if oldPledge.PledgeAddress != info.PledgeAddress || oldPledge.WithdrawTime != info.WithdrawTime ||
				oldPledge.Beneficial != info.Beneficial || oldPledge.PType != info.PType {
				return nil, errors.New("invalid saved pledge info")
			}
		}
	}

	newPledgeInfo, _ := cabi.NEP5PledgeABI.PackVariable(cabi.VariableNEP5PledgeInfo, info.PType, info.Amount,
		info.Beneficial, info.PledgeAddress)
	err := ctx.SetStorage(types.NEP5PledgeAddress[:], pledgeKey, newPledgeInfo)
	if err != nil {
		return nil, err
	}
	//tm, err := ctx.GetTokenMeta(param.Beneficial, common.ChainToken())
	//
	//if err != nil {
	//	return nil, err
	//}
	//
	//if tm != nil {
	//	block.Previous = tm.Header
	//	block.Balance = tm.Balance
	//} else {
	//	block.Previous = types.ZeroHash
	//	block.Balance = types.ZeroBalance
	//}
	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = input.Token
	block.Link = input.GetHash()
	block.Data = newPledgeInfo
	block.Vote = types.ZeroBalance
	block.Network = types.ZeroBalance
	block.Oracle = types.ZeroBalance
	block.Storage = types.ZeroBalance
	//block.Timestamp = time.Now().UTC().Unix()

	switch param.PType {
	case cabi.Netowrk:
		block.Network = amount
	case cabi.Oracle:
		block.Oracle = amount
	case cabi.Storage:
		block.Storage = amount
	case cabi.Vote:
		block.Vote = amount
	default:
		break
	}

	return []*ContractBlock{
		{
			Block:     block,
			ToAddress: input.Address,
			BlockType: types.ContractReward,
			Amount:    amount,
			Token:     input.Token,
			Data:      newPledgeInfo,
		},
	}, nil
}

func (*Nep5Pledge) GetRefundData() []byte {
	return []byte{1}
}

type WithdrawNep5Pledge struct {
}

func (*WithdrawNep5Pledge) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (*WithdrawNep5Pledge) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) (err error) {
	if amount, err := ctx.CalculateAmount(block); err != nil {
		return err
	} else {
		if block.Type != types.Send || amount.Compare(types.ZeroBalance) != types.BalanceCompEqual {
			return errors.New("invalid block ")
		}
	}

	param := new(cabi.WithdrawPledgeParam)
	if err := cabi.NEP5PledgeABI.UnpackMethod(param, cabi.MethodWithdrawNEP5Pledge, block.Data); err != nil {
		return errors.New("invalid input data")
	}

	if block.Data, err = cabi.NEP5PledgeABI.PackMethod(cabi.MethodWithdrawNEP5Pledge, param.Beneficial,
		param.Amount, param.PType); err != nil {
		return
	}

	return nil
}

func (*WithdrawNep5Pledge) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.WithdrawPledgeParam)
	err := cabi.NEP5PledgeABI.UnpackMethod(param, cabi.MethodWithdrawNEP5Pledge, input.Data)
	if err != nil {
		return nil, err
	}

	pledgeResults := cabi.SearchBeneficialPledgeInfo(ctx, param)

	if len(pledgeResults) == 0 {
		return nil, errors.New("pledge is not ready")
	}

	if len(pledgeResults) > 2 {
		sort.Slice(pledgeResults, func(i, j int) bool {
			return pledgeResults[i].PledgeInfo.WithdrawTime > pledgeResults[j].PledgeInfo.WithdrawTime
		})
	}

	pledgeInfo := pledgeResults[0]

	amount, _ := ctx.CalculateAmount(input)

	err = ctx.SetStorage(types.NEP5PledgeAddress[:], pledgeInfo.Key, nil)
	if err != nil {
		return nil, err
	}

	data, err := cabi.NEP5PledgeABI.PackVariable(cabi.VariableNEP5PledgeInfo, pledgeInfo)
	if err != nil {
		return nil, err
	}

	block.Type = types.ContractReward
	block.Address = pledgeInfo.PledgeInfo.PledgeAddress
	block.Token = input.Token
	block.Link = input.GetHash()
	block.Data = data
	block.Vote = types.ZeroBalance
	block.Network = types.ZeroBalance
	block.Oracle = types.ZeroBalance
	block.Storage = types.ZeroBalance

	return []*ContractBlock{
		{
			Block:     block,
			ToAddress: pledgeInfo.PledgeInfo.PledgeAddress,
			BlockType: types.ContractReward,
			Amount:    amount,
			Token:     input.Token,
			Data:      data,
		},
	}, nil
}

func (*WithdrawNep5Pledge) GetRefundData() []byte {
	return []byte{2}
}
