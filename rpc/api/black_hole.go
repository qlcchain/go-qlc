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

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type BlackHoleAPI struct {
	logger            *zap.SugaredLogger
	l                 *ledger.Ledger
	blackHoleContract *contract.BlackHole
	cc                *context.ChainContext
}

func NewBlackHoleApi(l *ledger.Ledger, cc *context.ChainContext) *BlackHoleAPI {
	api := &BlackHoleAPI{
		logger:            log.NewLogger("rpc/black_hole"),
		l:                 l,
		blackHoleContract: &contract.BlackHole{},
		cc:                cc,
	}
	return api
}

func (b *BlackHoleAPI) GetSendBlock(param *cabi.DestroyParam) (*types.StateBlock, error) {
	if b.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	vmContext := vmstore.NewVMContext(b.l)
	stateBlock, err := cabi.PackSendBlock(vmContext, param)
	if err != nil {
		return nil, err
	}
	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		povHeader, err := b.l.GetLatestPovHeader()
		if err != nil {
			return nil, fmt.Errorf("get pov header error: %s", err)
		}
		stateBlock.PoVHeight = povHeader.GetHeight()
		stateBlock.Extra = *h
	}
	return stateBlock, nil
}

func (b *BlackHoleAPI) GetRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	if send == nil {
		return nil, ErrParameterNil
	}
	if b.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
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
			povHeader, err := b.l.GetLatestPovHeader()
			if err != nil {
				return nil, fmt.Errorf("get pov header error: %s", err)
			}
			r[0].Block.PoVHeight = povHeader.GetHeight()
			return r[0].Block, nil
		} else {
			return nil, errors.New("fail to generate black hole reward block")
		}
	} else {
		return nil, err
	}
}

func (b *BlackHoleAPI) GetTotalDestroyInfo(addr *types.Address) (types.Balance, error) {
	if addr == nil || addr.IsZero() {
		return types.ZeroBalance, ErrParameterNil
	}

	vmContext := vmstore.NewVMContext(b.l)
	return cabi.GetTotalDestroyInfo(vmContext, addr)
}

func (b *BlackHoleAPI) GetDestroyInfoDetail(addr *types.Address) ([]*cabi.DestroyInfo, error) {
	if addr == nil || addr.IsZero() {
		return nil, ErrParameterNil
	}

	vmContext := vmstore.NewVMContext(b.l)
	return cabi.GetDestroyInfoDetail(vmContext, addr)
}
