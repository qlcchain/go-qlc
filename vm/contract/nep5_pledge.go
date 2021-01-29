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
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var Nep5Contract = NewChainContract(
	map[string]Contract{
		cabi.MethodNEP5Pledge: &Nep5Pledge{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer1,
					signature: true,
					work:      true,
				},
			},
		},
		cabi.MethodWithdrawNEP5Pledge: &WithdrawNep5Pledge{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer1,
					signature: true,
					work:      true,
				},
			},
		},
	},
	cabi.NEP5PledgeABI,
	cabi.JsonNEP5Pledge,
)

type pledgeInfo struct {
	pledgeTime   *timeSpan
	pledgeAmount *big.Int
}

type Nep5Pledge struct {
	BaseContract
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

	if block.Token != cfg.ChainToken() || amount.IsZero() || !b {
		return fmt.Errorf("invalid block data, %t, %t, %t (chain token %s)", block.Token != cfg.ChainToken(), amount.IsZero(), !b, cfg.ChainToken())
	}

	param, err := cabi.ParsePledgeParam(block.GetData())
	if err != nil {
		return errors.New("invalid beneficial address")
	}

	pt := cabi.PledgeType(param.PType)
	if info, b := config[pt]; !b {
		return fmt.Errorf("unsupport type %s", pt.String())
	} else if amount.Compare(types.Balance{Int: info.pledgeAmount}) == types.BalanceCompSmaller {
		return fmt.Errorf("not enough pledge amount %s, expect %s", amount.String(), info.pledgeAmount)
	}

	if param.PledgeAddress != block.Address {
		return fmt.Errorf("invalid pledge address[%s],expect %s",
			param.PledgeAddress.String(), block.Address.String())
	}

	block.Data, err = cabi.NEP5PledgeABI.PackMethod(cabi.MethodNEP5Pledge, param.Beneficial,
		param.PledgeAddress, param.PType, param.NEP5TxId)
	if err != nil {
		return err
	}
	return nil
}

func (*Nep5Pledge) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return nil, nil, errors.New("not implemented")
	//param, err := cabi.ParsePledgeParam(block.GetData())
	//if err != nil {
	//	return nil, nil, fmt.Errorf("invalid beneficial address, %s", err)
	//}
	//
	//return &types.PendingKey{
	//		Address: param.Beneficial,
	//		Hash:    block.GetHash(),
	//	}, &types.PendingInfo{
	//		Source: types.Address(block.GetLink()),
	//		Amount: types.ZeroBalance,
	//		Type:   block.Token,
	//	}, nil
}

func (*Nep5Pledge) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param, err := cabi.ParsePledgeParam(input.Data)
	if err != nil {
		return nil, err
	}
	a, _ := ctx.CalculateAmount(input)
	amount := a.Copy()

	var withdrawTime int64
	pt := cabi.PledgeType(param.PType)
	if info, b := config[pt]; b {
		withdrawTime = info.pledgeTime.Calculate(time.Unix(input.Timestamp, 0)).Unix()
	} else {
		return nil, fmt.Errorf("unsupport type %s", pt.String())
	}

	info := cabi.NEP5PledgeInfo{
		PType:         param.PType,
		Amount:        amount.Int,
		WithdrawTime:  withdrawTime,
		Beneficial:    param.Beneficial,
		PledgeAddress: param.PledgeAddress,
		NEP5TxId:      param.NEP5TxId,
	}

	if _, err := ctx.GetStorage(contractaddress.NEP5PledgeAddress[:], []byte(param.NEP5TxId)); err == nil {
		return nil, fmt.Errorf("invalid pledge nep5 tx id, %s", param.NEP5TxId)
	} else {
		if err := ctx.SetStorage(contractaddress.NEP5PledgeAddress[:], []byte(param.NEP5TxId), nil); err != nil {
			return nil, err
		}
	}

	// save data
	pledgeKey := cabi.GetPledgeKey(input.Address, param.Beneficial, param.NEP5TxId)
	pledgeData, err := info.ToABI()
	if err != nil {
		return nil, err
	}
	err = ctx.SetStorage(contractaddress.NEP5PledgeAddress[:], pledgeKey, pledgeData)
	if err != nil {
		return nil, err
	}

	//var pledgeData []byte
	//if pledgeData, err = ctx.GetStorage(contractaddress.NEP5PledgeAddress[:], pledgeKey); err != nil && err != vmstore.ErrStorageNotFound {
	//	return nil, err
	//} else {
	//	// already exist,verify data
	//	if len(pledgeData) > 0 {
	//		oldPledge, err := cabi.ParsePledgeInfo(pledgeData)
	//		if err != nil {
	//			return nil, err
	//		}
	//		if oldPledge.PledgeAddress != info.PledgeAddress || oldPledge.WithdrawTime != info.WithdrawTime ||
	//			oldPledge.Beneficial != info.Beneficial || oldPledge.PType != info.PType ||
	//			oldPledge.NEP5TxId != info.NEP5TxId {
	//			return nil, errors.New("invalid saved pledge info")
	//		}
	//	} else {
	//		// save data
	//		pledgeData, err = info.ToABI()
	//		if err != nil {
	//			return nil, err
	//		}
	//		err = ctx.SetStorage(contractaddress.NEP5PledgeAddress[:], pledgeKey, pledgeData)
	//		if err != nil {
	//			return nil, err
	//		}
	//	}
	//}

	am, _ := ctx.GetAccountMeta(param.Beneficial)
	if am != nil {
		tm := am.Token(cfg.ChainToken())
		block.Balance = am.CoinBalance
		block.Vote = &am.CoinVote
		block.Network = &am.CoinNetwork
		block.Oracle = &am.CoinOracle
		block.Storage = &am.CoinStorage
		if tm != nil {
			block.Previous = tm.Header
			block.Representative = tm.Representative
		} else {
			block.Previous = types.ZeroHash
			block.Representative = input.Representative
		}
	} else {
		//block.Vote = types.ZeroBalance
		//block.Network = types.ZeroBalance
		//block.Oracle = types.ZeroBalance
		//block.Storage = types.ZeroBalance
		block.Previous = types.ZeroHash
		block.Representative = input.Representative
	}

	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = input.Token
	block.Link = input.GetHash()
	block.Data = pledgeData

	//TODO: query snapshot balance
	switch cabi.PledgeType(param.PType) {
	case cabi.Network:
		block.Network = types.ToBalance(block.GetNetwork().Add(amount))
	case cabi.Oracle:
		block.Oracle = types.ToBalance(block.GetOracle().Add(amount))
	case cabi.Storage:
		block.Storage = types.ToBalance(block.GetStorage().Add(amount))
	case cabi.Vote:
		block.Vote = types.ToBalance(block.GetVote().Add(amount))
	default:
		break
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: param.Beneficial,
			BlockType: types.ContractReward,
			Amount:    amount,
			Token:     input.Token,
			Data:      pledgeData,
		},
	}, nil
}

func (*Nep5Pledge) GetRefundData() []byte {
	return []byte{1}
}

func (*Nep5Pledge) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	data := block.GetData()
	tr := types.ZeroAddress

	if method, err := cabi.NEP5PledgeABI.MethodById(data[0:4]); err == nil {
		if method.Name == cabi.MethodNEP5Pledge {
			param := new(cabi.PledgeParam)
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

type WithdrawNep5Pledge struct {
	BaseContract
}

func (*WithdrawNep5Pledge) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	if amount, err := ctx.CalculateAmount(block); err != nil {
		return err
	} else {
		if block.Type != types.ContractSend || amount.Compare(types.ZeroBalance) == types.BalanceCompEqual {
			return errors.New("invalid block ")
		}
	}

	param, err := cabi.ParseWithdrawPledgeParam(block.GetData())
	if err != nil {
		return errors.New("invalid input data")
	}

	if block.Data, err = param.ToABI(); err != nil {
		return err
	}

	return nil
}

func (m *WithdrawNep5Pledge) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return nil, nil, errors.New("not implemented")
}

func (*WithdrawNep5Pledge) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param, err := cabi.ParseWithdrawPledgeParam(input.GetData())
	if err != nil {
		return nil, err
	}

	pledgeResult := cabi.SearchBeneficialPledgeInfoByTxId(ctx, param)

	if pledgeResult == nil {
		return nil, ErrPledgeNotReady
	}

	//if len(pledgeResults) > 2 {
	//	sort.Slice(pledgeResults, func(i, j int) bool {
	//		return pledgeResults[i].PledgeInfo.WithdrawTime > pledgeResults[j].PledgeInfo.WithdrawTime
	//	})
	//}

	pledgeInfo := pledgeResult

	amount, _ := ctx.CalculateAmount(input)

	var pledgeData []byte
	if pledgeData, err = ctx.GetStorage(nil, pledgeInfo.Key); err != nil && err != vmstore.ErrStorageNotFound {
		return nil, err
	} else {
		// already exist,verify data
		if len(pledgeData) > 0 {
			oldPledge, err := cabi.ParsePledgeInfo(pledgeData)
			if err != nil {
				return nil, err
			}
			if oldPledge.PledgeAddress != pledgeInfo.PledgeInfo.PledgeAddress || oldPledge.WithdrawTime != pledgeInfo.PledgeInfo.WithdrawTime ||
				oldPledge.Beneficial != pledgeInfo.PledgeInfo.Beneficial || oldPledge.PType != pledgeInfo.PledgeInfo.PType ||
				oldPledge.NEP5TxId != pledgeInfo.PledgeInfo.NEP5TxId {
				return nil, errors.New("invalid saved pledge info")
			}

			// TODO: save data or change pledge info state
			err = ctx.SetStorage(nil, pledgeInfo.Key, nil)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("invalid pledge data")
		}
	}

	am, _ := ctx.GetAccountMeta(pledgeInfo.PledgeInfo.PledgeAddress)
	if am == nil {
		return nil, fmt.Errorf("can not find %s", pledgeInfo.PledgeInfo.PledgeAddress.String())
	}
	tm := am.Token(cfg.ChainToken())
	if tm != nil {
		block.Type = types.ContractReward
		block.Address = pledgeInfo.PledgeInfo.PledgeAddress
		block.Token = input.Token
		block.Link = input.GetHash()
		block.Data = pledgeData
		block.Vote = &am.CoinVote
		block.Network = &am.CoinNetwork
		block.Oracle = &am.CoinOracle
		block.Storage = &am.CoinStorage
		block.Previous = tm.Header
		block.Representative = tm.Representative
		block.Balance = am.CoinBalance.Add(amount)

		return []*ContractBlock{
			{
				VMContext: ctx,
				Block:     block,
				ToAddress: pledgeInfo.PledgeInfo.PledgeAddress,
				BlockType: types.ContractReward,
				Amount:    amount,
				Token:     input.Token,
				Data:      pledgeData,
			},
		}, nil
	} else {
		return nil, fmt.Errorf("chain token %s do not found", cfg.ChainToken().String())
	}
}

func (*WithdrawNep5Pledge) GetRefundData() []byte {
	return []byte{2}
}

func (*WithdrawNep5Pledge) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	data := block.GetData()
	tr := types.ZeroAddress

	if method, err := cabi.NEP5PledgeABI.MethodById(data[0:4]); err == nil {
		if method.Name == cabi.MethodWithdrawNEP5Pledge {
			param := new(cabi.WithdrawPledgeParam)
			if err = method.Inputs.Unpack(param, data[4:]); err == nil {
				pledgeResult := cabi.SearchBeneficialPledgeInfoByTxId(ctx, param)
				if pledgeResult != nil {
					tr = pledgeResult.PledgeInfo.PledgeAddress
					return tr, nil
				} else {
					return tr, errors.New("find pledge err")
				}
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

// SetPledgeTime only for test
func SetPledgeTime(y, m, d, h, i, s int) {
	t := &timeSpan{
		years:   y,
		months:  m,
		days:    d,
		hours:   h,
		minutes: i,
		seconds: s,
	}
	config = map[cabi.PledgeType]pledgeInfo{
		cabi.Network: {
			pledgeTime:   t, // 3 minutes
			pledgeAmount: big.NewInt(10 * 1e8),
		},
		cabi.Vote: {
			pledgeTime:   t, //  10 minutes
			pledgeAmount: big.NewInt(1 * 1e8),
		},
		cabi.Oracle: {
			pledgeAmount: big.NewInt(3 * 1e8), // 3M
		},
	}
}
