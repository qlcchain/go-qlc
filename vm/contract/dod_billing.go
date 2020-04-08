package contract

import (
	"github.com/qlcchain/go-qlc/common/types"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type DoDCreateAccount struct {
	BaseContract
}

func (ca *DoDCreateAccount) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	admin := new(abi.DoDAccount)
	err := abi.DoDBillingABI.UnpackMethod(admin, abi.MethodNameDoDCreateAccount, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	return nil, nil, nil
}

type DoDCoupleAccount struct {
	BaseContract
}

func (ca *DoDCoupleAccount) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	admin := new(abi.DoDAccount)
	err := abi.DoDBillingABI.UnpackMethod(admin, abi.MethodNameDoDCoupleAccount, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	return nil, nil, nil
}

type DoDServiceSet struct {
	BaseContract
}

func (ss *DoDServiceSet) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	admin := new(abi.DoDService)
	err := abi.DoDBillingABI.UnpackMethod(admin, abi.MethodNameDoDServiceSet, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	return nil, nil, nil
}

type DoDUserServiceSet struct {
	BaseContract
}

func (us *DoDUserServiceSet) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	admin := new(abi.DoDAccount)
	err := abi.DoDBillingABI.UnpackMethod(admin, abi.MethodNameDoDUserServiceSet, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	return nil, nil, nil
}

type DoDUsageUpdate struct {
	BaseContract
}

func (uu *DoDUsageUpdate) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	admin := new(abi.DoDAccount)
	err := abi.DoDBillingABI.UnpackMethod(admin, abi.MethodNameDoDUsageUpdate, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	return nil, nil, nil
}
