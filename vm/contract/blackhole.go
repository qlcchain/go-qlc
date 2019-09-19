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

func (b *BlackHole) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.DestroyParam)
	err := cabi.RewardsABI.UnpackMethod(param, cabi.MethodNameDestroy, block.Data)
	if err != nil {
		return nil, nil, err
	}

	if err := b.verify(ctx, param, block); err == nil {
		//TODO: update database...

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

func (b *BlackHole) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	panic("implement me")
}

func (b *BlackHole) GetRefundData() []byte {
	return []byte{1}
}
