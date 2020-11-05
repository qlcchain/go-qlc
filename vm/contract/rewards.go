/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"fmt"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var RewardsContract = NewChainContract(
	map[string]Contract{
		cabi.MethodNameAirdropRewards: &AirdropRewards{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer: SpecVer1,
					pending: true,
				},
			},
		},
		cabi.MethodNameConfidantRewards: &ConfidantRewards{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer: SpecVer1,
					pending: true,
				},
			},
		},
	},
	cabi.RewardsABI,
	cabi.JsonRewards,
)

type AirdropRewards struct {
	BaseContract
}

func (ar *AirdropRewards) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	param, err := cabi.ParseRewardsParam(cabi.MethodNameAirdropRewards, block.Data)
	if err != nil {
		return err
	}

	if _, err := param.Verify(block.Address, cabi.MethodNameUnsignedAirdropRewards); err != nil {
		return err
	}

	return nil
}

func (ar *AirdropRewards) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return doPending(block, cabi.MethodNameAirdropRewards, cabi.MethodNameUnsignedAirdropRewards)
}

func (ar *AirdropRewards) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return generate(ctx, cabi.MethodNameAirdropRewards, cabi.MethodNameUnsignedAirdropRewards,
		block, input, func(param *cabi.RewardsParam) []byte {
			return cabi.GetRewardsKey(param.Id[:], param.TxHeader[:], param.RxHeader[:])
		})
}

func (*AirdropRewards) GetRefundData() []byte {
	return []byte{1}
}

func (*AirdropRewards) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	data := block.GetData()
	tr := types.ZeroAddress

	if method, err := cabi.RewardsABI.MethodById(data[0:4]); err == nil {
		param := new(cabi.RewardsParam)
		if err = method.Inputs.Unpack(param, data[4:]); err == nil {
			tr = param.Beneficial
			return tr, nil
		} else {
			return tr, err
		}
	} else {
		return tr, err
	}
}

type ConfidantRewards struct {
	BaseContract
}

func (*ConfidantRewards) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	param, err := cabi.ParseRewardsParam(cabi.MethodNameConfidantRewards, block.Data)
	if err != nil {
		return err
	}

	if _, err := param.Verify(block.Address, cabi.MethodNameUnsignedConfidantRewards); err != nil {
		return err
	}

	return nil
}

func (ar *ConfidantRewards) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return doPending(block, cabi.MethodNameConfidantRewards, cabi.MethodNameUnsignedConfidantRewards)
}

func doPending(block *types.StateBlock, signed, unsigned string) (*types.PendingKey, *types.PendingInfo, error) {
	param, err := cabi.ParseRewardsParam(signed, block.Data)
	if err != nil {
		return nil, nil, err
	}

	if _, err := param.Verify(block.Address, unsigned); err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: param.Beneficial,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Amount: types.Balance{Int: param.Amount},
			Type:   block.Token,
		}, nil
}

func (*ConfidantRewards) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock,
	input *types.StateBlock) ([]*ContractBlock, error) {
	return generate(ctx, cabi.MethodNameConfidantRewards, cabi.MethodNameUnsignedConfidantRewards,
		block, input, func(param *cabi.RewardsParam) []byte {
			return cabi.GetConfidantKey(param.Beneficial, param.Id[:], param.TxHeader[:], param.RxHeader[:])
		})
}

func (*ConfidantRewards) GetRefundData() []byte {
	return []byte{2}
}

func (*ConfidantRewards) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	data := block.GetData()
	tr := types.ZeroAddress

	if method, err := cabi.RewardsABI.MethodById(data[0:4]); err == nil {
		param := new(cabi.RewardsParam)
		if err = method.Inputs.Unpack(param, data[4:]); err == nil {
			tr = param.Beneficial
			return tr, nil
		} else {
			return tr, err
		}
	} else {
		return tr, err
	}
}

func generate(ctx *vmstore.VMContext, signed, unsigned string, block *types.StateBlock, input *types.StateBlock,
	fn func(param *cabi.RewardsParam) []byte) ([]*ContractBlock, error) {
	param, err := cabi.ParseRewardsParam(signed, input.Data)
	if err != nil {
		return nil, err
	}

	if _, err := param.Verify(input.Address, unsigned); err != nil {
		return nil, err
	}

	//verify is QGAS
	amount, err := ctx.CalculateAmount(input)
	if err != nil {
		return nil, err
	}
	if amount.Sign() > 0 && amount.Compare(types.ZeroBalance) == types.BalanceCompBigger && input.Token == cfg.GasToken() {
		txHash := input.GetHash()
		txAddress := input.Address
		txMeta, err := ctx.GetAccountMeta(txAddress)
		if err != nil {
			return nil, err
		}
		txToken := txMeta.Token(input.Token)
		rxAddress := param.Beneficial

		rxMeta, _ := ctx.GetAccountMeta(rxAddress)

		block.Type = types.ContractReward
		block.Address = rxAddress
		block.Link = txHash
		block.Token = input.Token
		//block.Extra = types.ZeroHash
		//block.Vote = types.ZeroBalance
		//block.Network = types.ZeroBalance
		//block.Oracle = types.ZeroBalance
		//block.Storage = types.ZeroBalance
		//block.Timestamp = common.TimeNow().UTC().Unix()

		// already have account
		if rxMeta != nil && len(rxMeta.Tokens) > 0 {
			if rxToken := rxMeta.Token(input.Token); rxToken != nil {
				//already have token
				block.Balance = rxToken.Balance.Add(amount)
				block.Previous = rxToken.Header
				block.Representative = txToken.Representative
			} else {
				block.Balance = amount
				block.Previous = types.ZeroHash
				//use other token's rep
				block.Representative = rxMeta.Tokens[0].Representative
			}
		} else {
			block.Balance = amount
			block.Previous = types.ZeroHash
			block.Representative = input.Representative
		}

		t := uint8(cabi.Rewards)
		if signed == cabi.MethodNameConfidantRewards {
			t = uint8(cabi.Confidant)
		}

		info := &cabi.RewardsInfo{
			Type:     t,
			From:     input.Address,
			To:       rxAddress,
			TxHeader: txToken.Header,
			RxHeader: block.Previous,
			Amount:   amount.Int,
		}

		key := fn(param)
		if data, err := ctx.GetStorage(contractaddress.RewardsAddress[:], key); err != nil && err != vmstore.ErrStorageNotFound {
			return nil, err
		} else {
			//already exist
			if len(data) > 0 {
				if rewardsInfo, err := cabi.ParseRewardsInfo(data); err == nil {
					if rewardsInfo.Amount.Cmp(info.Amount) != 0 || rewardsInfo.Type != info.Type ||
						//rewardsInfo.TxHeader != info.TxHeader || rewardsInfo.RxHeader != info.RxHeader ||
						rewardsInfo.From != info.From || rewardsInfo.To != info.To {
						return nil, fmt.Errorf("invalid saved confidant data: txHeader(%s,%s,%t);"+
							" rxHeader(%s,%s,%t); amount(%s,%s,%t); type(%d,%d,%t); from(%s,%s,%t); to(%s,%s,%t)",
							rewardsInfo.TxHeader, info.TxHeader, rewardsInfo.TxHeader == info.TxHeader,
							rewardsInfo.RxHeader, info.RxHeader, rewardsInfo.RxHeader == info.RxHeader,
							rewardsInfo.Amount, info.Amount, rewardsInfo.Amount.Cmp(info.Amount) == 0,
							rewardsInfo.Type, info.Type, rewardsInfo.Type == info.Type,
							rewardsInfo.From, info.From, rewardsInfo.From == info.From,
							rewardsInfo.To, info.To, rewardsInfo.To == info.To)
					}
				} else {
					return nil, err
				}
			} else {
				if data, err := info.ToABI(); err == nil {
					if err := ctx.SetStorage(contractaddress.RewardsAddress[:], key, data); err != nil {
						return nil, err
					}
				} else {
					return nil, err
				}
			}
		}

		return []*ContractBlock{
			{
				VMContext: ctx,
				Block:     block,
				ToAddress: rxAddress,
				BlockType: types.ContractReward,
				Amount:    amount,
				Token:     input.Token,
				Data:      []byte{},
			},
		}, nil
	} else {
		return nil, fmt.Errorf("invalid token hash %s or amount %s", input.Token.String(), amount.String())
	}
}
