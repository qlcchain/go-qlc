package contract

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type VerifierRegister struct {
	WithSignNoPending
}

func (vr *VerifierRegister) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	reg := new(cabi.VerifierRegInfo)
	err := cabi.VerifierABI.UnpackMethod(reg, cabi.MethodNameVerifierRegister, block.Data)
	if err != nil {
		return nil, nil, err
	}

	err = cabi.VerifierPledgeCheck(ctx, reg.Account)
	if err != nil {
		return nil, nil, err
	}

	err = cabi.VerifierRegInfoCheck(ctx, reg.Account, reg.VType, reg.VInfo)
	if err != nil {
		return nil, nil, err
	}

	block.Data, err = cabi.VerifierABI.PackMethod(cabi.MethodNameVerifierRegister, reg.Account, reg.VType, reg.VInfo)
	if err != nil {
		return nil, nil, err
	}

	err = vr.SetStorage(ctx, reg.Account, reg.VType, reg.VInfo)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (vr *VerifierRegister) SetStorage(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) error {
	data, err := cabi.VerifierABI.PackVariable(cabi.VariableNameVerifierInfo, vInfo)
	if err != nil {
		return err
	}

	vrKey := append(util.BE_Uint32ToBytes(vType), account[:]...)
	err = ctx.SetStorage(types.VerifierAddress[:], vrKey, data)
	if err != nil {
		return err
	}

	return nil
}

func (vr *VerifierRegister) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.NewBalance(0), nil
}

func (vr *VerifierRegister) GetRefundData() []byte {
	return []byte{1}
}

func (vr *VerifierRegister) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	return 0, nil
}

func (vr *VerifierRegister) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

type VerifierUnregister struct {
	WithSignNoPending
}

func (vu *VerifierUnregister) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	reg := new(cabi.VerifierRegInfo)
	err := cabi.VerifierABI.UnpackMethod(reg, cabi.MethodNameVerifierUnregister, block.Data)
	if err != nil {
		return nil, nil, err
	}

	err = cabi.VerifierUnRegInfoCheck(ctx, reg.Account, reg.VType)
	if err != nil {
		return nil, nil, err
	}

	block.Data, err = cabi.VerifierABI.PackMethod(cabi.MethodNameVerifierUnregister, reg.Account, reg.VType)
	if err != nil {
		return nil, nil, err
	}

	err = vu.SetStorage(ctx, reg.Account, reg.VType)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (vu *VerifierUnregister) SetStorage(ctx *vmstore.VMContext, account types.Address, vType uint32) error {
	vuKey := append(util.BE_Uint32ToBytes(vType), account[:]...)
	ctx.DelStorage(types.VerifierAddress[:], vuKey)
	return nil
}

func (vu *VerifierUnregister) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.NewBalance(0), nil
}

func (vu *VerifierUnregister) GetRefundData() []byte {
	return []byte{1}
}

func (vu *VerifierUnregister) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	return 0, nil
}

func (vu *VerifierUnregister) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}
