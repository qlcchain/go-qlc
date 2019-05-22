/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type AirdropRewords struct {
}

func (ar *AirdropRewords) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (ar *AirdropRewords) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	param := new(cabi.RewardsParam)
	err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameAirdropRewards, block.Data)
	if err != nil {
		return err
	}

	if _, err := param.Verify(block.Address); err != nil {
		return err
	}

	return nil
}

func (ar *AirdropRewords) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.RewardsParam)
	err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameAirdropRewards, block.Data)
	if err != nil {
		return nil, nil, err
	}

	if _, err := param.Verify(block.Address); err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: *param.Beneficial,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: block.Address,
			Amount: types.Balance{Int: param.Amount},
			Type:   block.Token,
		}, nil
}

func (ar *AirdropRewords) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return generate(ctx, block, input, func(param *cabi.RewardsParam) []byte {
		return cabi.GetRewardsKey(param.Id, param.TxHeader, param.RxHeader)
	})
}

func (*AirdropRewords) GetRefundData() []byte {
	return []byte{1}
}

type ConfidantRewards struct {
}

func (*ConfidantRewards) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (*ConfidantRewards) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	param := new(cabi.RewardsParam)
	err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameAirdropRewards, block.Data)
	if err != nil {
		return err
	}

	if _, err := param.Verify(block.Address); err != nil {
		return err
	}

	return nil
}

func (ar *ConfidantRewards) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.RewardsParam)
	err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameAirdropRewards, block.Data)
	if err != nil {
		return nil, nil, err
	}

	if _, err := param.Verify(block.Address); err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: *param.Beneficial,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: block.Address,
			Amount: types.Balance{Int: param.Amount},
			Type:   block.Token,
		}, nil
}

func (*ConfidantRewards) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return generate(ctx, block, input, func(param *cabi.RewardsParam) []byte {
		return cabi.GetConfidantKey(*param.Beneficial, param.Id, param.TxHeader, param.RxHeader)
	})
}

func (*ConfidantRewards) GetRefundData() []byte {
	return []byte{2}
}

func generate(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock, fn func(param *cabi.RewardsParam) []byte) ([]*ContractBlock, error) {
	param := new(cabi.RewardsParam)
	err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameAirdropRewards, input.Data)
	if err != nil {
		return nil, err
	}

	if _, err := param.Verify(input.Address); err != nil {
		return nil, err
	}

	//verify is QGAS
	amount, _ := ctx.CalculateAmount(input)
	if amount.Sign() > 0 && input.Token != common.GasToken() {
		txHash := input.GetHash()
		txAddress := input.Address
		txMeta, err := ctx.GetAccountMeta(txAddress)
		if err != nil {
			return nil, err
		}
		txToken := txMeta.Token(input.Token)
		rxAddress := *param.Beneficial

		rxMeta, _ := ctx.GetAccountMeta(rxAddress)

		block.Type = types.ContractReward
		block.Address = rxAddress
		block.Link = txHash
		block.Token = input.Token
		block.Extra = types.ZeroHash
		//block.Timestamp = common.TimeNow().UTC().Unix()

		// already have account
		if rxMeta != nil && len(rxMeta.Tokens) > 0 {
			block.Vote = rxMeta.CoinVote
			block.Oracle = rxMeta.CoinOracle
			block.Network = rxMeta.CoinNetwork
			block.Storage = rxMeta.CoinStorage
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
			block.Vote = types.ZeroBalance
			block.Network = types.ZeroBalance
			block.Oracle = types.ZeroBalance
			block.Storage = types.ZeroBalance
			block.Previous = types.ZeroHash
			block.Representative = input.Representative
		}

		info := &cabi.RewardsInfo{
			Type:     uint8(cabi.Confidant),
			From:     &input.Address,
			To:       &rxAddress,
			TxHeader: txToken.Header[:],
			RxHeader: block.Previous[:],
			Amount:   amount.Int,
		}

		//key := cabi.GetConfidantKey(rxAddress, param.Id, param.TxHeader, param.RxHeader)

		key := fn(param)
		if data, err := ctx.GetStorage(types.RewardsAddress[:], key); err != nil && err != vmstore.ErrStorageNotFound {
			return nil, err
		} else {
			//already exist
			if len(data) > 0 {
				if rewardsInfo, err := cabi.ParseRewardsInfo(data); err == nil {
					if !bytes.EqualFold(rewardsInfo.TxHeader, info.TxHeader) || !bytes.EqualFold(rewardsInfo.RxHeader, info.RxHeader) ||
						rewardsInfo.Amount.Cmp(info.Amount) != 0 || rewardsInfo.Type != info.Type ||
						rewardsInfo.From != info.From || rewardsInfo.To != info.To {
						return nil, errors.New("invalid saved confidant data")
					}
				} else {
					return nil, err
				}
			} else {
				if data, err := cabi.RewardsABI.PackVariable(cabi.VariableNameRewards, info); err == nil {
					if err := ctx.SetStorage(types.RewardsAddress[:], key, data); err != nil {
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
		return nil, fmt.Errorf("invalid token hash %s", input.Token.String())
	}
}
