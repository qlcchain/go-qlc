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

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type BlackHole struct {
}

// TODO: save contract data
func (b *BlackHole) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.DestroyParam)
	err := cabi.RewardsABI.UnpackMethod(param, cabi.MethodNameDestroy, block.Data)
	if err != nil {
		return nil, nil, err
	}

	if err := b.verify(ctx, param, block); err == nil {
		// make sure that the same block only process once
		if b, err := ctx.GetStorage(block.Address[:], block.Previous[:]); err == nil && len(b) > 0 {
			return nil, nil, fmt.Errorf("block already processed, %s of %s",
				block.Previous.String(), block.Address.String())
		} else if err != vmstore.ErrStorageNotFound {
			return nil, nil, err
		}

		if data, err := cabi.RewardsABI.PackVariable(cabi.VariableDestroyInfo, block.Address, block.Previous,
			block.Token, param.Amount, common.TimeNow()); err == nil {
			if err := ctx.SetStorage(block.Address[:], block.Previous[:], data); err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, err
		}

		return &types.PendingKey{
				Address: param.Owner,
				Hash:    block.GetHash(),
			}, &types.PendingInfo{
				Source: types.Address(block.Link),
				Amount: types.ZeroBalance,
				Type:   block.Token,
			}, nil
	} else {
		return nil, nil, err
	}
}

func (b *BlackHole) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (b *BlackHole) verify(ctx *vmstore.VMContext, param *cabi.DestroyParam, block *types.StateBlock) error {
	var data []byte

	data = append(data, param.Owner[:]...)
	data = append(data, param.Amount.Bytes()...)
	data = append(data, param.Token[:]...)

	verify := param.Owner.Verify(data, param.Sign[:])
	if !verify {
		return errors.New("invalid sign")
	}
	if block.Token != common.GasToken() {
		return fmt.Errorf("invalid token: %s", block.Token.String())
	}
	if amount, err := ctx.CalculateAmount(block); err == nil {
		if amount.Compare(types.Balance{Int: param.Amount}) != types.BalanceCompEqual {
			return fmt.Errorf("amount mistmatch, exp: %s,act:%s", param.Amount.String(), amount.String())
		}
	} else {
		return fmt.Errorf("can not calcuate transfer amount, %s", err.Error())
	}

	return nil
}

func (b *BlackHole) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock,
	input *types.StateBlock) ([]*ContractBlock, error) {
	// verify send block data
	param := new(cabi.DestroyParam)
	err := cabi.RewardsABI.UnpackMethod(param, cabi.MethodNameDestroy, input.Data)
	if err != nil {
		return nil, err
	}

	rxMeta, _ := ctx.GetAccountMeta(input.Address)
	// qgas token should be exist
	rxToken := rxMeta.Token(input.Token)
	txHash := input.GetHash()

	block.Type = types.ContractReward
	block.Address = input.Address
	block.Link = txHash
	block.Token = input.Token
	block.Extra = types.ZeroHash
	block.Vote = types.ZeroBalance
	block.Network = types.ZeroBalance
	block.Oracle = types.ZeroBalance
	block.Storage = types.ZeroBalance

	block.Balance = rxToken.Balance
	block.Previous = rxToken.Header
	block.Representative = input.Representative

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: input.Address,
			BlockType: types.ContractReward,
			Amount:    types.ZeroBalance,
			Token:     input.Token,
			Data:      []byte{},
		},
	}, nil
}

func (b *BlackHole) GetRefundData() []byte {
	return []byte{1}
}
