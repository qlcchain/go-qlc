/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/contract"
	"go.uber.org/zap"
)

type ContractApi struct {
	logger *zap.SugaredLogger
	ledger *ledger.Ledger
}

func NewContractApi(ledger *ledger.Ledger) *ContractApi {
	return &ContractApi{logger: log.NewLogger("api_contract"), ledger: ledger}
}

func (c *ContractApi) GetAbiByContractAddress(address types.Address) (string, error) {
	return contract.GetAbiByContractAddress(address)
}

func (c *ContractApi) PackContractData(abiStr string, methodName string, params []string) ([]byte, error) {
	abiContract, err := abi.JSONToABIContract(strings.NewReader(abiStr))
	if err != nil {
		return nil, err
	}
	method, ok := abiContract.Methods[methodName]
	if !ok {
		return nil, errors.New("method name not found")
	}
	arguments, err := convert(params, method.Inputs)
	if err != nil {
		return nil, err
	}
	return abiContract.PackMethod(methodName, arguments...)
}

func (c *ContractApi) ContractAddressList() []types.Address {
	return types.ChainContractAddressList
}

func convert(params []string, arguments abi.Arguments) ([]interface{}, error) {
	if len(params) != len(arguments) {
		return nil, errors.New("argument size not match")
	}
	resultList := make([]interface{}, len(params))
	for i, argument := range arguments {
		result, err := convertOne(params[i], argument.Type)
		if err != nil {
			return nil, err
		}
		resultList[i] = result
	}
	return resultList, nil
}

func convertOne(param string, t abi.Type) (interface{}, error) {
	typeString := t.String()
	if strings.Contains(typeString, "[") {
		return convertToArray(param, t)
	} else if typeString == "bool" {
		return convertToBool(param)
	} else if strings.HasPrefix(typeString, "int") {
		return convertToInt(param, t.Size)
	} else if strings.HasPrefix(typeString, "uint") {
		return convertToUint(param, t.Size)
	} else if typeString == "address" {
		return types.HexToAddress(param)
	} else if typeString == "tokenId" {
		return types.NewHash(param)
	} else if typeString == "string" {
		return param, nil
	} else if typeString == "bytes" {
		return convertToDynamicBytes(param)
	} else if strings.HasPrefix(typeString, "bytes") {
		return convertToFixedBytes(param, t.Size)
	}
	return nil, errors.New("unknown type " + typeString)
}

func convertToArray(param string, t abi.Type) (interface{}, error) {
	if t.Elem.Elem != nil {
		return nil, errors.New(t.String() + " type not supported")
	}
	typeString := t.Elem.String()
	if typeString == "bool" {
		return convertToBoolArray(param)
	} else if strings.HasPrefix(typeString, "int") {
		return convertToIntArray(param, *t.Elem)
	} else if strings.HasPrefix(typeString, "uint") {
		return convertToUintArray(param, *t.Elem)
	} else if typeString == "address" {
		return convertToAddressArray(param)
	} else if typeString == "tokenId" {
		return convertToTokenIdArray(param)
	} else if typeString == "string" {
		return convertToStringArray(param)
	}
	return nil, errors.New(typeString + " array type not supported")
}

func convertToBoolArray(param string) (interface{}, error) {
	resultList := make([]bool, 0)
	if err := json.Unmarshal([]byte(param), &resultList); err != nil {
		return nil, err
	}
	return resultList, nil
}

func convertToIntArray(param string, t abi.Type) (interface{}, error) {
	size := t.Size
	if size == 8 {
		resultList := make([]int8, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 16 {
		resultList := make([]int16, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 32 {
		resultList := make([]int32, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 64 {
		resultList := make([]int64, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else {
		resultList := make([]*big.Int, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	}
}
func convertToUintArray(param string, t abi.Type) (interface{}, error) {
	size := t.Size
	if size == 8 {
		resultList := make([]uint8, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 16 {
		resultList := make([]uint16, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 32 {
		resultList := make([]uint32, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 64 {
		resultList := make([]uint64, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else {
		resultList := make([]*big.Int, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	}
}
func convertToAddressArray(param string) (interface{}, error) {
	resultList := make([]types.Address, 0)
	if err := json.Unmarshal([]byte(param), &resultList); err != nil {
		return nil, err
	}
	return resultList, nil
}
func convertToTokenIdArray(param string) (interface{}, error) {
	resultList := types.ZeroHash
	if err := json.Unmarshal([]byte(param), &resultList); err != nil {
		return nil, err
	}
	return resultList, nil
}

func convertToStringArray(param string) (interface{}, error) {
	resultList := make([]string, 0)
	if err := json.Unmarshal([]byte(param), &resultList); err != nil {
		return nil, err
	}
	return resultList, nil
}

func convertToBool(param string) (interface{}, error) {
	if param == "true" {
		return true, nil
	} else {
		return false, nil
	}
}

func convertToInt(param string, size int) (interface{}, error) {
	bigInt, ok := new(big.Int).SetString(param, 0)
	if !ok || bigInt.BitLen() > size-1 {
		return nil, errors.New(param + " convert to int failed")
	}
	if size == 8 {
		return int8(bigInt.Int64()), nil
	} else if size == 16 {
		return int16(bigInt.Int64()), nil
	} else if size == 32 {
		return int32(bigInt.Int64()), nil
	} else if size == 64 {
		return int64(bigInt.Int64()), nil
	} else {
		return bigInt, nil
	}
}

func convertToUint(param string, size int) (interface{}, error) {
	bigInt, ok := new(big.Int).SetString(param, 0)
	if !ok || bigInt.BitLen() > size {
		return nil, errors.New(param + " convert to uint failed")
	}
	if size == 8 {
		return uint8(bigInt.Uint64()), nil
	} else if size == 16 {
		return uint16(bigInt.Uint64()), nil
	} else if size == 32 {
		return uint32(bigInt.Uint64()), nil
	} else if size == 64 {
		return uint64(bigInt.Uint64()), nil
	} else {
		return bigInt, nil
	}
}

func convertToBytes(param string, size int) (interface{}, error) {
	if size == 0 {
		return convertToDynamicBytes(param)
	} else {
		return convertToFixedBytes(param, size)
	}
}

func convertToFixedBytes(param string, size int) (interface{}, error) {
	if len(param) != size*2 {
		return nil, errors.New(param + " is not valid bytes")
	}
	return hex.DecodeString(param)
}
func convertToDynamicBytes(param string) (interface{}, error) {
	return hex.DecodeString(param)
}
