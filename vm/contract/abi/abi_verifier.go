package abi

import (
	"bytes"
	"fmt"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"strings"
)

const (
	JsonVerifier = `
	[
		{"type":"function","name":"VerifierRegister","inputs":[
			{"name":"account","type":"address"},
			{"name":"vType","type":"uint32"},
			{"name":"vInfo","type":"string"}
		]},
		{"type":"function","name":"VerifierUnregister","inputs":[
			{"name":"account","type":"address"},
			{"name":"vType","type":"uint32"}
		]},
		{"type":"variable","name":"VerifierInfo","inputs":[
			{"name":"vInfo","type":"string"}
		]}
	]`

	MethodNameVerifierRegister   = "VerifierRegister"
	MethodNameVerifierUnregister = "VerifierUnregister"
	VariableNameVerifierInfo     = "VerifierInfo"
)

var (
	VerifierABI, _ = abi.JSONToABIContract(strings.NewReader(JsonVerifier))
)

const (
	// prefix + contractAddr + type + account => info
	VerifierTypeIndexS = 1 + types.AddressSize
	VerifierTypeIndexE = VerifierTypeIndexS + 4
	VerifierAccIndexS  = VerifierTypeIndexE
	VerifierAccIndexE  = VerifierAccIndexS + 32
)

type VerifierRegInfo struct {
	Account types.Address
	VType   uint32
	VInfo   string
}

type VerifierUnRegInfo struct {
	Account types.Address
	VType   uint32
}

func VerifierRegInfoCheck(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) error {
	switch vType {
	case types.OracleTypeEmail:
		if !util.VerifyEmailFormat(vInfo) {
			return fmt.Errorf("invalid email format (%s)", vInfo)
		}

		if CheckVerifierInfoExist(ctx, account, vType, vInfo) {
			return fmt.Errorf("email has been registered (%s)", vInfo)
		}
	case types.OracleTypeWeChat:
		if CheckVerifierInfoExist(ctx, account, vType, vInfo) {
			return fmt.Errorf("weChat id has been registered (%s)", vInfo)
		}
	default:
		return fmt.Errorf("invalid verifier type(%s)", types.OracleTypeToString(vType))
	}

	return nil
}

func VerifierUnRegInfoCheck(ctx *vmstore.VMContext, account types.Address, vType uint32) error {
	switch vType {
	case types.OracleTypeEmail, types.OracleTypeWeChat:
		if !CheckVerifierExist(ctx, account, vType) {
			return fmt.Errorf("there is no valid registered info")
		}
	default:
		return fmt.Errorf("invalid verifier type(%s)", types.OracleTypeToString(vType))
	}

	return nil
}

func VerifierPledgeCheck(ctx *vmstore.VMContext, account types.Address) error {
	am, err := ctx.Ledger.GetAccountMeta(account)
	if err != nil {
		return err
	}

	minPledgeAmount := types.Balance{Int: types.MinVerifierPledgeAmount}
	if am.CoinOracle.Compare(minPledgeAmount) == types.BalanceCompSmaller {
		return fmt.Errorf("%s have not enough oracle pledge %s", account, am.CoinOracle)
	}

	return nil
}

func CheckVerifierInfoExist(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) bool {
	key := append(util.BE_Uint32ToBytes(vType), account[:]...)
	val, err := ctx.GetStorage(types.VerifierAddress[:], key)
	if err != nil {
		return false
	}

	var info string
	err = VerifierABI.UnpackVariable(&info, VariableNameVerifierInfo, val)
	if err != nil {
		return false
	}

	if info == vInfo {
		return true
	}

	return false
}

func CheckVerifierExist(ctx *vmstore.VMContext, account types.Address, vType uint32) bool {
	key := append(util.BE_Uint32ToBytes(vType), account[:]...)
	val, err := ctx.GetStorage(types.VerifierAddress[:], key)
	if err != nil {
		return false
	}

	var info string
	err = VerifierABI.UnpackVariable(&info, VariableNameVerifierInfo, val)
	if err != nil {
		return false
	}

	return true
}

func GetVerifierInfoByAccountAndType(ctx *vmstore.VMContext, account types.Address, vType uint32) (string, error) {
	key := append(util.BE_Uint32ToBytes(vType), account[:]...)
	val, err := ctx.GetStorage(types.VerifierAddress[:], key)
	if err != nil {
		return "", err
	}

	var info string
	err = VerifierABI.UnpackVariable(&info, VariableNameVerifierInfo, val)
	if err != nil {
		return "", err
	}

	return info, nil
}

func GetAllVerifiers(ctx *vmstore.VMContext) ([]*VerifierRegInfo, error) {
	vrs := make([]*VerifierRegInfo, 0)
	err := ctx.Iterator(types.VerifierAddress[:], func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[VerifierAccIndexS:VerifierAccIndexE])
		if err != nil {
			return err
		}

		var info string
		err = VerifierABI.UnpackVariable(&info, VariableNameVerifierInfo, value)
		if err != nil {
			return err
		}

		vr := &VerifierRegInfo{
			Account: addr,
			VType:   util.BE_BytesToUint32(key[VerifierTypeIndexS:VerifierTypeIndexE]),
			VInfo:   info,
		}

		vrs = append(vrs, vr)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return vrs, nil
}

func GetVerifiersByType(ctx *vmstore.VMContext, vType uint32) ([]*VerifierRegInfo, error) {
	vrs := make([]*VerifierRegInfo, 0)
	itKey := append(types.VerifierAddress[:], util.BE_Uint32ToBytes(vType)...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[VerifierAccIndexS:VerifierAccIndexE])
		if err != nil {
			return err
		}

		var info string
		err = VerifierABI.UnpackVariable(&info, VariableNameVerifierInfo, value)
		if err != nil {
			return err
		}

		vr := &VerifierRegInfo{
			Account: addr,
			VType:   util.BE_BytesToUint32(key[VerifierTypeIndexS:VerifierTypeIndexE]),
			VInfo:   info,
		}

		vrs = append(vrs, vr)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return vrs, nil
}

func GetVerifiersByAccount(ctx *vmstore.VMContext, account types.Address) ([]*VerifierRegInfo, error) {
	vrs := make([]*VerifierRegInfo, 0)
	err := ctx.Iterator(types.VerifierAddress[:], func(key []byte, value []byte) error {
		addrBytes := key[VerifierAccIndexS:VerifierAccIndexE]
		if bytes.Equal(addrBytes, account[:]) {
			addr, err := types.BytesToAddress(addrBytes)
			if err != nil {
				return err
			}

			var info string
			err = VerifierABI.UnpackVariable(&info, VariableNameVerifierInfo, value)
			if err != nil {
				return err
			}

			vr := &VerifierRegInfo{
				Account: addr,
				VType:   util.BE_BytesToUint32(key[VerifierTypeIndexS:VerifierTypeIndexE]),
				VInfo:   info,
			}

			vrs = append(vrs, vr)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return vrs, nil
}
