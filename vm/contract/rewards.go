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

type AirdropRewords struct {
}

func (*AirdropRewords) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	panic("implement me")
}

func (*AirdropRewords) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	panic("implement me")
}

func (*AirdropRewords) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	panic("implement me")
}

func (*AirdropRewords) GetRefundData() []byte {
	panic("implement me")
}

type ConfidantRewards struct {
}

func (*ConfidantRewards) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	panic("implement me")
}

func (*ConfidantRewards) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	panic("implement me")
}

func (*ConfidantRewards) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	panic("implement me")
}

func (*ConfidantRewards) GetRefundData() []byte {
	panic("implement me")
}
