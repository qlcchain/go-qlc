/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"encoding/json"
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

	block.Data, err = cabi.NEP5PledgeABI.PackMethod(cabi.MethodNEP5Pledge, param.Beneficial, uint8(param.PType))
	if err != nil {
		return err
	}
	return nil
}

func (*Nep5Pledge) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
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
		PledgeAddress: input.Address,
	}

	bytes, _ := json.Marshal(info)
	key := types.HashData(bytes)

	pledgeKey := cabi.GetPledgeKey(input.Address, param.Beneficial)

	oldPledgeData, err := ctx.GetStorage(types.NEP5PledgeAddress[:], pledgeKey)
	if err != nil {
		return nil, err
	}

	cache := make(map[types.Hash]cabi.NEP5PledgeInfo)

	if len(oldPledgeData) > 0 {
		err := json.Unmarshal(oldPledgeData, cache)
		if err != nil {
			return nil, err
		}
	}

	if _, ok := cache[key]; !ok {
		cache[key] = info
	}

	newPledgeInfo, err := json.Marshal(cache)
	if err != nil {
		return nil, err
	}

	err = ctx.SetStorage(types.NEP5PledgeAddress[:], pledgeKey, newPledgeInfo)
	if err != nil {
		return nil, err
	}

	pledgeInfo, _ := cabi.NEP5PledgeABI.PackVariable(cabi.VariableNEP5PledgeInfo, info.PType, info.Amount, info.Beneficial, info.PledgeAddress)

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
	block.Data = pledgeInfo
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
			Data:      pledgeInfo,
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

	if block.Data, err = cabi.NEP5PledgeABI.PackMethod(cabi.MethodWithdrawNEP5Pledge, param.Beneficial, param.Amount, param.PType); err != nil {
		return
	}

	return nil
}

func (*WithdrawNep5Pledge) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.WithdrawPledgeParam)
	err := cabi.NEP5PledgeABI.UnpackMethod(param, cabi.MethodWithdrawNEP5Pledge, input.Data)
	if err != nil {
		return nil, err
	}

	infos, _ := cabi.GetBeneficialPledgeInfos(ctx, param.Beneficial, param.PType)
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].WithdrawTime > infos[j].WithdrawTime
	})

	now := time.Now().UTC().Unix()
	var key types.Hash
	var info cabi.NEP5PledgeInfo
	for _, v := range infos {
		if v.Amount == param.Amount && now < v.WithdrawTime {
			bytes, _ := json.Marshal(v)
			key = types.HashData(bytes)
			info = *v
			break
		}
	}

	if key.IsZero() {
		return nil, errors.New("can not pledge info")
	}

	amount, _ := ctx.CalculateAmount(input)

	pledgeKey := cabi.GetPledgeKey(input.Address, param.Beneficial)
	oldPledgeData, err := ctx.GetStorage(types.NEP5PledgeAddress[:], pledgeKey)
	if err != nil {
		return nil, err
	}
	cache := make(map[types.Hash]cabi.NEP5PledgeInfo)

	if len(oldPledgeData) > 0 {
		err := json.Unmarshal(oldPledgeData, cache)
		if err != nil {
			return nil, err
		}
	}

	if _, ok := cache[key]; ok {
		delete(cache, key)
	}

	newPledgeInfo, err := json.Marshal(cache)
	if err != nil {
		return nil, err
	}

	err = ctx.SetStorage(types.NEP5PledgeAddress[:], pledgeKey, newPledgeInfo)
	if err != nil {
		return nil, err
	}

	pledgeInfo, _ := cabi.NEP5PledgeABI.PackVariable(cabi.VariableNEP5PledgeInfo, info.PType, info.Amount, info.Beneficial, info.PledgeAddress)

	block.Type = types.ContractReward
	block.Address = info.PledgeAddress
	block.Token = input.Token
	block.Link = input.GetHash()
	block.Data = pledgeInfo
	block.Vote = types.ZeroBalance
	block.Network = types.ZeroBalance
	block.Oracle = types.ZeroBalance
	block.Storage = types.ZeroBalance
	//block.Timestamp = time.Now().UTC().Unix()

	return []*ContractBlock{
		{
			Block:     block,
			ToAddress: info.PledgeAddress,
			BlockType: types.ContractReward,
			Amount:    amount,
			Token:     input.Token,
			Data:      pledgeInfo,
		},
	}, nil
}

func (*WithdrawNep5Pledge) GetRefundData() []byte {
	return []byte{2}
}
