/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type CreateContract struct {
}

func (c *CreateContract) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	panic("implement me")
}

func (c *CreateContract) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	panic("implement me")
}

func (c *CreateContract) GetRefundData() []byte {
	panic("implement me")
}

func (c *CreateContract) GetDescribe() Describe {
	panic("implement me")
}

func (c *CreateContract) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	panic("implement me")
}

func (c *CreateContract) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	panic("implement me")
}

type SignContract struct {
}

func (s *SignContract) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	panic("implement me")
}

func (s *SignContract) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	panic("implement me")
}

func (s *SignContract) GetRefundData() []byte {
	panic("implement me")
}

func (s *SignContract) GetDescribe() Describe {
	panic("implement me")
}

func (s *SignContract) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	panic("implement me")
}

func (s *SignContract) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	panic("implement me")
}

type ProcessCDR struct {
}

func (p *ProcessCDR) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	panic("implement me")
}

func (p *ProcessCDR) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	panic("implement me")
}

func (p *ProcessCDR) GetRefundData() []byte {
	panic("implement me")
}

func (p *ProcessCDR) GetDescribe() Describe {
	panic("implement me")
}

func (p *ProcessCDR) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	panic("implement me")
}

func (p *ProcessCDR) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	panic("implement me")
}
