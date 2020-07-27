package abi

import (
	"strings"

	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	JsonKYCStatus = `[
		{"type":"function","name":"KYCAdminHandOver","inputs":[
			{"name":"account","type":"address"},
			{"name":"comment","type":"string"}
		]},
		{"type":"function","name":"KYCStatusUpdate","inputs":[
			{"name":"chainAddress","type":"address"},
			{"name":"status","type":"string"}
		]},
		{"type":"function","name":"KYCTradeAddressUpdate","inputs":[
			{"name":"chainAddress","type":"address"},
			{"name":"action","type":"uint8"},
			{"name":"tradeAddress","type":"string"},
			{"name":"comment","type":"string"}
		]},
		{"type":"function","name":"KYCOperatorUpdate","inputs":[
			{"name":"account","type":"address"},
			{"name":"action","type":"uint8"},
			{"name":"comment","type":"string"}
		]}
	]`

	MethodNameKYCAdminHandOver      = "KYCAdminHandOver"
	MethodNameKYCStatusUpdate       = "KYCStatusUpdate"
	MethodNameKYCTradeAddressUpdate = "KYCTradeAddressUpdate"
	MethodNameKYCOperatorUpdate     = "KYCOperatorUpdate"
)

var (
	KYCStatusABI, _ = abi.JSONToABIContract(strings.NewReader(JsonKYCStatus))
	KYCStatusMap    = map[string]bool{
		"KYC_STATUS_NOT_STARTED":            true,
		"KYC_STATUS_IN_PROGRESS":            true,
		"KYC_STATUS_PROCESSING":             true,
		"KYC_STATUS_FAILED_JUMIO":           true,
		"KYC_STATUS_FAILED_COMPLYADVANTAGE": true,
		"KYC_STATUS_DENIED":                 true,
		"KYC_STATUS_PENDING":                true,
		"KYC_STATUS_PENDING_INSTITUTION":    true,
		"KYC_STATUS_APPROVED":               true,
		"KYC_STATUS_SET_FOR_CLOSURE":        true,
	}
)

func KYCIsAdmin(ctx *vmstore.VMContext, addr types.Address) bool {
	csdb, err := ctx.PoVContractState()
	if err != nil {
		return false
	}

	// if there is no admin, genesis is admin
	if addr == cfg.GenesisAddress() {
		itor := csdb.NewCurTireIterator(statedb.PovCreateContractLocalStateKey(KYCDataAdmin, nil))

		for _, val, ok := itor.Next(); ok; _, val, ok = itor.Next() {
			admin := new(KYCAdminAccount)
			_, err := admin.UnmarshalMsg(val)
			if err != nil {
				continue
			}

			if admin.Valid {
				return false
			}
		}

		return true
	} else {
		trieKey := statedb.PovCreateContractLocalStateKey(KYCDataAdmin, addr.Bytes())

		valBytes, err := csdb.GetValue(trieKey)
		if err != nil || len(valBytes) == 0 {
			return false
		}

		admin := new(KYCAdminAccount)
		_, err = admin.UnmarshalMsg(valBytes)
		if err != nil {
			return false
		}

		return admin.Valid
	}
}

func KYCGetAdmin(ctx *vmstore.VMContext) ([]*KYCAdminAccount, error) {
	admins := make([]*KYCAdminAccount, 0)
	csdb, err := ctx.PoVContractState()
	if err != nil {
		return nil, err
	}

	itor := csdb.NewCurTireIterator(statedb.PovCreateContractLocalStateKey(KYCDataAdmin, nil))
	for key, val, ok := itor.Next(); ok; key, val, ok = itor.Next() {
		admin := new(KYCAdminAccount)
		_, err := admin.UnmarshalMsg(val)
		if err != nil {
			return nil, err
		}

		addr, err := types.BytesToAddress(key[2:])
		if err != nil {
			return nil, err
		}

		if admin.Valid {
			admin.Account = addr
			admins = append(admins, admin)
		}
	}

	if len(admins) == 0 {
		genesisAdmin := &KYCAdminAccount{
			Account: cfg.GenesisAddress(),
			Comment: "Initial admin",
		}
		admins = append(admins, genesisAdmin)
	}

	return admins, nil
}

func KYCIsOperator(ctx *vmstore.VMContext, addr types.Address) bool {
	csdb, err := ctx.PoVContractState()
	if err != nil {
		return false
	}

	trieKey := statedb.PovCreateContractLocalStateKey(KYCDataOperator, addr.Bytes())

	valBytes, err := csdb.GetValue(trieKey)
	if err != nil || len(valBytes) == 0 {
		return false
	}

	oa := new(KYCOperatorAccount)
	_, err = oa.UnmarshalMsg(valBytes)
	if err != nil {
		return false
	}

	return oa.Valid
}

func KYCGetOperator(ctx *vmstore.VMContext) ([]*KYCOperatorAccount, error) {
	oas := make([]*KYCOperatorAccount, 0)
	csdb, err := ctx.PoVContractState()
	if err != nil {
		return nil, err
	}

	itor := csdb.NewCurTireIterator(statedb.PovCreateContractLocalStateKey(KYCDataOperator, nil))
	for key, val, ok := itor.Next(); ok; key, val, ok = itor.Next() {
		oa := new(KYCOperatorAccount)
		_, err := oa.UnmarshalMsg(val)
		if err != nil {
			return nil, err
		}

		addr, err := types.BytesToAddress(key[2:])
		if err != nil {
			return nil, err
		}

		if oa.Valid {
			oa.Account = addr
			oas = append(oas, oa)
		}
	}

	return oas, nil
}

func KYCGetAllStatus(ctx *vmstore.VMContext) ([]*KYCStatus, error) {
	status := make([]*KYCStatus, 0)

	csdb, err := ctx.PoVContractState()
	if err != nil {
		return nil, err
	}

	itor := csdb.NewCurTireIterator(statedb.PovCreateContractLocalStateKey(KYCDataStatus, nil))
	for key, val, ok := itor.Next(); ok; key, val, ok = itor.Next() {
		s := new(KYCStatus)
		_, err := s.UnmarshalMsg(val)
		if err != nil {
			return nil, err
		}

		addr, err := types.BytesToAddress(key[2 : 2+types.AddressSize])
		if err != nil {
			return nil, err
		}

		if s.Valid {
			s.ChainAddress = addr
			status = append(status, s)
		}
	}

	return status, nil
}

func KYCGetStatusByChainAddress(ctx *vmstore.VMContext, address types.Address) (*KYCStatus, error) {
	csdb, err := ctx.PoVContractState()
	if err != nil {
		return nil, err
	}

	trieKey := statedb.PovCreateContractLocalStateKey(KYCDataStatus, address.Bytes())
	data, err := csdb.GetValue(trieKey)
	if err != nil {
		return nil, err
	}

	s := new(KYCStatus)
	_, err = s.UnmarshalMsg(data)
	if err != nil {
		return nil, err
	}

	if s.Valid {
		s.ChainAddress = address
		return s, nil
	}

	return nil, nil
}

func KYCGetStatusByTradeAddress(ctx *vmstore.VMContext, address string) (*KYCStatus, error) {
	csdb, err := ctx.PoVContractState()
	if err != nil {
		return nil, err
	}

	ka := &KYCAddress{TradeAddress: address}

	trieKey := statedb.PovCreateContractLocalStateKey(KYCDataTradeAddress, ka.GetKey())
	data, err := csdb.GetValue(trieKey)
	if err != nil {
		return nil, err
	}

	data2, err := csdb.GetValue(data)
	ks := new(KYCAddress)
	_, err = ks.UnmarshalMsg(data2)
	if err != nil {
		return nil, err
	}

	if !ks.Valid {
		return nil, nil
	}

	chainAddress, err := types.BytesToAddress(data[2 : 2+types.AddressSize])
	if err != nil {
		return nil, err
	}

	return KYCGetStatusByChainAddress(ctx, chainAddress)
}

func KYCGetTradeAddress(ctx *vmstore.VMContext, address types.Address) ([]*KYCAddress, error) {
	kas := make([]*KYCAddress, 0)

	csdb, err := ctx.PoVContractState()
	if err != nil {
		return nil, err
	}

	itor := csdb.NewCurTireIterator(statedb.PovCreateContractLocalStateKey(KYCDataAddress, address.Bytes()))
	for _, val, ok := itor.Next(); ok; _, val, ok = itor.Next() {
		ka := new(KYCAddress)
		_, err := ka.UnmarshalMsg(val)
		if err != nil {
			return nil, err
		}

		if ka.Valid {
			kas = append(kas, ka)
		}
	}

	return kas, nil
}
