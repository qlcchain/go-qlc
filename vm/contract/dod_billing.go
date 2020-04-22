package contract

import (
	"errors"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type DoDSetAccount struct {
	BaseContract
}

func (ca *DoDSetAccount) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	account := new(abi.DoDAccount)
	err := abi.DoDBillingABI.UnpackMethod(account, abi.MethodNameDoDSetAccount, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	err = ca.setStorage(ctx, account)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (ca *DoDSetAccount) setStorage(ctx *vmstore.VMContext, account *abi.DoDAccount) error {
	var key []byte
	ah := types.Sha256DHashData([]byte(account.AccountName))
	key = append(key, abi.DoDDataTypeAccount)
	key = append(key, ah.Bytes()...)

	retVal, _ := ctx.GetStorage(contractaddress.DoDBillingAddress.Bytes(), key)
	if len(retVal) > 0 {
		oa := new(abi.DoDAccount)
		_, err := oa.UnmarshalMsg(retVal)
		if err != nil {
			return err
		}

		if len(account.AccountType) == 0 {
			account.AccountType = oa.AccountType
		}

		if len(account.UUID) == 0 {
			account.UUID = oa.UUID
		}

		if len(account.AccountInfo) == 0 {
			account.AccountInfo = oa.AccountInfo
		}

		account.Connections = oa.Connections
	}

	val, err := account.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return ctx.SetStorage(contractaddress.DoDBillingAddress.Bytes(), key, val)
}

type DoDSetService struct {
	BaseContract
}

func (ss *DoDSetService) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	conn := new(abi.DoDConnection)
	err := abi.DoDBillingABI.UnpackMethod(conn, abi.MethodNameDoDSetService, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	err = ss.setStorage(ctx, conn)
	if err != nil {
		return nil, nil, err
	}

	err = ss.updateAccountService(ctx, conn)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (ss *DoDSetService) setStorage(ctx *vmstore.VMContext, conn *abi.DoDConnection) error {
	var key []byte
	key = append(key, abi.DoDDataTypeConnection)
	key = append(key, conn.ConnectionID...)

	val, err := conn.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return ctx.SetStorage(contractaddress.DoDBillingAddress.Bytes(), key, val)
}

func (ss *DoDSetService) updateAccountService(ctx *vmstore.VMContext, conn *abi.DoDConnection) error {
	account := abi.GetAccount(ctx, conn.AccountName)
	if account == nil {
		return errors.New("account not exist")
	}

	for _, c := range account.Connections {
		if c == conn.ConnectionID {
			return nil
		}
	}

	account.Connections = append(account.Connections, conn.ConnectionID)

	var key []byte
	ah := types.Sha256DHashData([]byte(account.AccountName))
	key = append(key, abi.DoDDataTypeAccount)
	key = append(key, ah.Bytes()...)

	val, err := account.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return ctx.SetStorage(contractaddress.DoDBillingAddress.Bytes(), key, val)
}

type DoDUpdateUsage struct {
	BaseContract
}

func (uu *DoDUpdateUsage) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	usage := new(abi.DoDUsage)
	err := abi.DoDBillingABI.UnpackMethod(usage, abi.MethodNameDoDUpdateUsage, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	err = uu.setStorage(ctx, usage)
	if err != nil {
		return nil, nil, err
	}

	err = uu.updateService(ctx, usage)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (uu *DoDUpdateUsage) setStorage(ctx *vmstore.VMContext, usage *abi.DoDUsage) error {
	var key []byte
	key = append(key, abi.DoDDataTypeUsage)
	key = append(key, usage.ConnectionID...)
	key = append(key, util.BE_Uint64ToBytes(usage.Timestamp)...)

	val, err := usage.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return ctx.SetStorage(contractaddress.DoDBillingAddress.Bytes(), key, val)
}

func (uu *DoDUpdateUsage) updateService(ctx *vmstore.VMContext, usage *abi.DoDUsage) error {
	conn := abi.GetConnection(ctx, usage.ConnectionID)
	if conn == nil {
		return errors.New("connection not exist")
	}

	if conn.ChargeType == abi.DoDChargeTypeBandwidth && conn.PaidRule == abi.DoDPaidRulePrePaid && conn.BuyMode == abi.DoDBuyModeCommon {
		var expense float64

		if usage.Timestamp > conn.TempStartTime && usage.Timestamp <= conn.TempEndTime {
			expense = conn.TempPrice * usage.Usage
		} else {
			expense = conn.Price * usage.Usage
		}

		conn.Balance -= expense
	}

	if conn.ChargeType == abi.DoDChargeTypeUsage && conn.PaidRule == abi.DoDPaidRulePrePaid {
		if conn.BuyMode == abi.DoDBuyModeCommon {
			conn.Quota -= usage.Usage
		} else {
			conn.Balance -= usage.Usage * conn.Price
		}
	}

	var key []byte
	key = append(key, abi.DoDDataTypeConnection)
	key = append(key, conn.ConnectionID...)

	val, err := conn.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return ctx.SetStorage(contractaddress.DoDBillingAddress.Bytes(), key, val)
}
