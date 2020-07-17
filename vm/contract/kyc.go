package contract

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var KYCContract = NewChainContract(
	map[string]Contract{
		abi.MethodNameKYCAdminHandOver: &KYCAdminHandOver{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
					povState:  true,
				},
			},
		},
		abi.MethodNameKYCStatusUpdate: &KYCStatusUpdate{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
					povState:  true,
				},
			},
		},
		abi.MethodNameKYCTradeAddressUpdate: &KYCTradeAddressUpdate{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
					povState:  true,
				},
			},
		},
		abi.MethodNameKYCOperatorUpdate: &KYCOperatorUpdate{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
					povState:  true,
				},
			},
		},
	},
	abi.KYCStatusABI,
	abi.JsonKYCStatus,
)

type KYCAdminHandOver struct {
	BaseContract
}

func (a *KYCAdminHandOver) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	admin := new(abi.KYCAdminAccount)
	err := abi.KYCStatusABI.UnpackMethod(admin, abi.MethodNameKYCAdminHandOver, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	if len(admin.Comment) > abi.KYCCommentMaxLen {
		return nil, nil, ErrInvalidLen
	}

	// check admin if the block is not synced
	if !block.IsFromSync() {
		if !abi.KYCIsAdmin(ctx, block.Address) {
			return nil, nil, ErrInvalidAdmin
		}
	}

	block.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCAdminHandOver, admin.Account, admin.Comment)

	return nil, nil, nil
}

func (a *KYCAdminHandOver) removeAdmin(csdb *statedb.PovContractStateDB, addr types.Address) error {
	trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataAdmin, addr.Bytes())

	// genesis admin is not in db
	if addr == cfg.GenesisAddress() {
		return nil
	}

	valBytes, err := csdb.GetValue(trieKey)
	if err != nil || len(valBytes) == 0 {
		return errors.New("get admin err")
	}

	admin := new(abi.KYCAdminAccount)
	_, err = admin.UnmarshalMsg(valBytes)
	if err != nil {
		return err
	}

	admin.Valid = false
	data, err := admin.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, data)
}

func (a *KYCAdminHandOver) updateAdmin(csdb *statedb.PovContractStateDB, addr types.Address, comment string) error {
	trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataAdmin, addr.Bytes())

	// genesis admin is not in db
	if addr == cfg.GenesisAddress() {
		return nil
	}

	admin := new(abi.KYCAdminAccount)
	admin.Comment = comment
	admin.Valid = true
	data, err := admin.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, data)
}

func (a *KYCAdminHandOver) DoSendOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock) error {
	admin := new(abi.KYCAdminAccount)
	err := abi.KYCStatusABI.UnpackMethod(admin, abi.MethodNameKYCAdminHandOver, block.Data)
	if err != nil {
		return err
	}

	err = a.removeAdmin(csdb, block.Address)
	if err != nil {
		return err
	}

	return a.updateAdmin(csdb, admin.Account, admin.Comment)
}

type KYCStatusUpdate struct {
	BaseContract
}

func (n *KYCStatusUpdate) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	status := new(abi.KYCStatus)
	err := abi.KYCStatusABI.UnpackMethod(status, abi.MethodNameKYCStatusUpdate, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	err = n.VerifyParam(ctx, status)
	if err != nil {
		return nil, nil, ErrCheckParam
	}

	// check operator right if the block is not synced
	if !block.IsFromSync() {
		if !abi.KYCIsOperator(ctx, block.Address) {
			return nil, nil, ErrInvalidOperator
		}
	}

	block.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCStatusUpdate, status.ChainAddress, status.Status)

	return nil, nil, nil
}

func (n *KYCStatusUpdate) VerifyParam(ctx *vmstore.VMContext, status *abi.KYCStatus) error {
	if _, ok := abi.KYCStatusMap[status.Status]; !ok {
		return fmt.Errorf("invalid status %s", status.Status)
	}

	return nil
}

func (n *KYCStatusUpdate) DoSendOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock) error {
	status := new(abi.KYCStatus)
	err := abi.KYCStatusABI.UnpackMethod(status, abi.MethodNameKYCStatusUpdate, block.Data)
	if err != nil {
		return err
	}

	trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataStatus, status.ChainAddress.Bytes())

	status.Valid = true
	data, err := status.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, data)
}

type KYCTradeAddressUpdate struct {
	BaseContract
}

func (n *KYCTradeAddressUpdate) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	ka := new(abi.KYCAddress)
	err := abi.KYCStatusABI.UnpackMethod(ka, abi.MethodNameKYCTradeAddressUpdate, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	err = n.VerifyParam(ctx, ka)
	if err != nil {
		return nil, nil, ErrCheckParam
	}

	// check operator right if the block is not synced
	if !block.IsFromSync() {
		if !abi.KYCIsOperator(ctx, block.Address) {
			return nil, nil, ErrInvalidOperator
		}
	}

	block.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCTradeAddressUpdate, ka.ChainAddress, ka.Action, ka.TradeAddress, ka.Comment)

	return nil, nil, nil
}

func (n *KYCTradeAddressUpdate) VerifyParam(ctx *vmstore.VMContext, ka *abi.KYCAddress) error {
	switch ka.Action {
	case abi.KYCActionAdd:
	case abi.KYCActionRemove:
		_, err := abi.KYCGetStatusByTradeAddress(ctx, ka.TradeAddress)
		if err != nil {
			return fmt.Errorf("trade address %s not exist", ka.TradeAddress)
		}
	default:
		return fmt.Errorf("invalid action")
	}

	if len(ka.Comment) > abi.KYCCommentMaxLen {
		return fmt.Errorf("invalid comment len (max %d)", abi.KYCCommentMaxLen)
	}

	if len(ka.TradeAddress) == 0 {
		return fmt.Errorf("trade address is null")
	}

	return nil
}

func (n *KYCTradeAddressUpdate) DoSendOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock) error {
	ka := new(abi.KYCAddress)
	err := abi.KYCStatusABI.UnpackMethod(ka, abi.MethodNameKYCTradeAddressUpdate, block.Data)
	if err != nil {
		return err
	}

	switch ka.Action {
	case abi.KYCActionAdd:
		trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataAddress, ka.GetMixKey())

		ka.Valid = true
		data, err := ka.MarshalMsg(nil)
		if err != nil {
			return err
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			return err
		}

		tradeTrieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataTradeAddress, ka.GetKey())
		err = csdb.SetValue(tradeTrieKey, trieKey)
		if err != nil {
			return err
		}
	case abi.KYCActionRemove:
		trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataAddress, ka.GetMixKey())
		data, err := csdb.GetValue(trieKey)
		if err != nil {
			return err
		}

		oka := new(abi.KYCAddress)
		_, err = oka.UnmarshalMsg(data)
		if err != nil {
			return err
		}

		oka.Valid = false
		data, err = oka.MarshalMsg(nil)
		if err != nil {
			return err
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			return err
		}
	}

	return nil
}

type KYCOperatorUpdate struct {
	BaseContract
}

func (a *KYCOperatorUpdate) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	oa := new(abi.KYCOperatorAccount)
	err := abi.KYCStatusABI.UnpackMethod(oa, abi.MethodNameKYCOperatorUpdate, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	err = a.VerifyParam(ctx, oa)
	if err != nil {
		return nil, nil, ErrCheckParam
	}

	// check admin if the block is not synced
	if !block.IsFromSync() {
		if !abi.KYCIsAdmin(ctx, block.Address) {
			return nil, nil, ErrInvalidAdmin
		}
	}

	block.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCOperatorUpdate, oa.Account, oa.Action, oa.Comment)

	return nil, nil, nil
}

func (a *KYCOperatorUpdate) VerifyParam(ctx *vmstore.VMContext, oa *abi.KYCOperatorAccount) error {
	switch oa.Action {
	case abi.KYCActionAdd:
	case abi.KYCActionRemove:
		if !abi.KYCIsOperator(ctx, oa.Account) {
			return fmt.Errorf("operator %s not exist", oa.Account)
		}
	default:
		return fmt.Errorf("invalid action")
	}

	if len(oa.Comment) > abi.KYCCommentMaxLen {
		return fmt.Errorf("invalid comment len (max %d)", abi.KYCCommentMaxLen)
	}

	return nil
}

func (a *KYCOperatorUpdate) DoSendOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock) error {
	oa := new(abi.KYCOperatorAccount)
	err := abi.KYCStatusABI.UnpackMethod(oa, abi.MethodNameKYCOperatorUpdate, block.Data)
	if err != nil {
		return err
	}

	switch oa.Action {
	case abi.KYCActionAdd:
		trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataOperator, oa.Account.Bytes())

		oa.Valid = true
		data, err := oa.MarshalMsg(nil)
		if err != nil {
			return err
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			return err
		}
	case abi.KYCActionRemove:
		trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataOperator, oa.Account.Bytes())

		data, err := csdb.GetValue(trieKey)
		if err != nil {
			return err
		}

		noa := new(abi.KYCOperatorAccount)
		_, err = noa.UnmarshalMsg(data)
		if err != nil {
			return err
		}

		noa.Valid = false
		data, err = noa.MarshalMsg(nil)
		if err != nil {
			return err
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			return err
		}
	}

	return nil
}
