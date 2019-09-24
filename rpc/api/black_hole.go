/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"errors"

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
	vmContext := vmstore.NewVMContext(b.l)
	stateBlock, err := cabi.PackSendBlock(vmContext, param)
	if err != nil {
		return nil, err
	}
	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		stateBlock.Extra = *h
	}
	return stateBlock, nil
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
