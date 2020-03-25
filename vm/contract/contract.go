/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"errors"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var (
	ErrToken            = errors.New("token err")
	ErrUnpackMethod     = errors.New("unpack method err")
	ErrPackMethod       = errors.New("pack method err")
	ErrNotEnoughPledge  = errors.New("not enough pledge")
	ErrCheckParam       = errors.New("check param err")
	ErrSetStorage       = errors.New("set storage err")
	ErrCalcAmount       = errors.New("calc amount err")
	ErrNotEnoughFee     = errors.New("not enough fee")
	ErrGetVerifier      = errors.New("get verifier err")
	ErrAccountInvalid   = errors.New("invalid account")
	ErrAccountNotExist  = errors.New("account not exist")
	ErrGetNodeHeight    = errors.New("get node height err")
	ErrEndHeightInvalid = errors.New("invalid claim end height")
	ErrClaimRepeat      = errors.New("claim reward repeatedly")
	ErrGetRewardHistory = errors.New("get reward history err")
	ErrVerifierNum      = errors.New("verifier num err")
	ErrPledgeNotReady   = errors.New("pledge is not ready")
	ErrGetAdmin         = errors.New("get admin err")
	ErrInvalidAdmin     = errors.New("invalid admin")
	ErrInvalidLen       = errors.New("invalid len")
)

type BaseContract struct {
	vmcontract.Describe
}

func (c *BaseContract) GetDescribe() vmcontract.Describe {
	return c.Describe
}

func (c *BaseContract) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	return types.ZeroAddress, nil
}

func (c *BaseContract) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

// refund data at receive error
func (c *BaseContract) GetRefundData() []byte {
	return []byte{1}
}

// DoPending generate pending info from send block
func (c *BaseContract) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return nil, nil, nil
}

// ProcessSend verify or update StateBlock.Data
func (c *BaseContract) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	return errors.New("not implemented")
}

// check status, update state
func (c *BaseContract) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*vmcontract.ContractBlock, error) {
	return nil, errors.New("not implemented")
}

// ProcessSend verify or update StateBlock.Data
func (c *BaseContract) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return nil, nil, errors.New("not implemented")
}

func (c *BaseContract) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	return common.ContractNoGap, nil, nil
}

func (c *BaseContract) DoSendOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock) error {
	return errors.New("not implemented")
}

func (c *BaseContract) DoReceiveOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock, input *types.StateBlock) error {
	return errors.New("not implemented")
}

func (c *BaseContract) EventNotify(eb event.EventBus, ctx *vmstore.VMContext, block *types.StateBlock) error {
	return nil
}
