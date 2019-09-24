/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"go.uber.org/zap"
)

type BlackHoleApi struct {
	logger            *zap.SugaredLogger
	l                 *ledger.Ledger
	blackHoleContract *contract.BlackHole
}

func NewBlackHoleApi(l *ledger.Ledger) *BlackHoleApi {
	return &BlackHoleApi{logger: log.NewLogger("rpc/black_hole"), l: l, blackHoleContract: &contract.BlackHole{}}
}

func (b *BlackHoleApi) GetSendBlackHoleBlock(param *cabi.DestroyParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if isVerified, err := param.Verify(); err != nil {
		return nil, err
	} else if !isVerified {
		return nil, errors.New("invalid sign of param")
	}

	if tm, err := b.l.GetTokenMeta(param.Owner, param.Token); err != nil {
		return nil, err
	} else {
		if tm.Balance.Compare(types.Balance{Int: param.Amount}) == types.BalanceCompSmaller {
			return nil, fmt.Errorf("not enough balance, [%s] of [%s]", param.Amount.String(), tm.Balance.String())
		}

		if singedData, err := cabi.BlackHoleABI.PackMethod(cabi.MethodNameDestroy, param.Owner, param.Previous, param.Token,
			param.Amount, param.Sign); err == nil {

			return &types.StateBlock{
				Type:           types.ContractSend,
				Token:          tm.Type,
				Address:        param.Owner,
				Balance:        tm.Balance.Sub(types.Balance{Int: param.Amount}),
				Vote:           types.ZeroBalance,
				Network:        types.ZeroBalance,
				Oracle:         types.ZeroBalance,
				Storage:        types.ZeroBalance,
				Previous:       param.Previous,
				Link:           types.Hash(types.BlackHoleAddress),
				Representative: tm.Representative,
				Data:           singedData,
				Timestamp:      common.TimeNow().Unix(),
			}, nil
		} else {
			return nil, err
		}
	}
}

func (b *BlackHoleApi) GetReceiveBlackHoleBlock(send *types.Hash) (*types.StateBlock, error) {
	if send == nil {
		return nil, ErrParameterNil
	}
	blk, err := b.l.GetStateBlock(*send)
	if err != nil {
		return nil, err
	}

	rev := &types.StateBlock{
		Timestamp: common.TimeNow().Unix(),
	}
	vmContext := vmstore.NewVMContext(b.l)
	if r, err := b.blackHoleContract.DoReceive(vmContext, rev, blk); err == nil {
		if len(r) > 0 {
			return r[0].Block, nil
		} else {
			return nil, errors.New("fail to generate black hole reward block")
		}
	} else {
		return nil, err
	}
}

func (b *BlackHoleApi) GetTotalDestroyInfo(addr *types.Address) (types.Balance, error) {
	if addr == nil || addr.IsZero() {
		return types.ZeroBalance, ErrParameterNil
	}

	vmContext := vmstore.NewVMContext(b.l)
	return cabi.GetTotalDestroyInfo(vmContext, addr)
}

func (b *BlackHoleApi) GetDestroyInfoDetail(addr *types.Address) ([]*cabi.DestroyInfo, error) {
	if addr == nil || addr.IsZero() {
		return nil, ErrParameterNil
	}

	vmContext := vmstore.NewVMContext(b.l)
	return cabi.GetDestroyInfoDetail(vmContext, addr)
}
