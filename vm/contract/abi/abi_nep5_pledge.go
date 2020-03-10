/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	JsonNEP5Pledge = `
	[
		{"type":"function","name":"NEP5Pledge", "inputs":[{"name":"beneficial","type":"address"},{"name":"pledgeAddress","type":"address"},{"name":"pType","type":"uint8"},{"name":"NEP5TxId","type":"string"}]},
		{"type":"function","name":"WithdrawNEP5Pledge","inputs":[{"name":"beneficial","type":"address"},{"name":"amount","type":"uint256"},{"name":"pType","type":"uint8"},{"name":"NEP5TxId","type":"string"}]},
		{"type":"variable","name":"nep5PledgeInfo","inputs":[{"name":"pType","type":"uint8"},{"name":"amount","type":"uint256"},{"name":"withdrawTime","type":"int64"},{"name":"beneficial","type":"address"},{"name":"pledgeAddress","type":"address"},{"name":"NEP5TxId","type":"string"}]}
	]`

	MethodNEP5Pledge         = "NEP5Pledge"
	MethodWithdrawNEP5Pledge = "WithdrawNEP5Pledge"
	VariableNEP5PledgeInfo   = "nep5PledgeInfo"
)

var (
	NEP5PledgeABI, _ = abi.JSONToABIContract(strings.NewReader(JsonNEP5Pledge))
)

//go:generate stringer -type=PledgeType
type PledgeType uint8

const (
	Network PledgeType = iota
	Vote
	Storage
	Oracle
	Invalid
)

type PledgeParam struct {
	Beneficial    types.Address
	PledgeAddress types.Address
	PType         uint8
	NEP5TxId      string
}

type VariablePledgeBeneficial struct {
	Amount *big.Int
	PType  PledgeType
}

type WithdrawPledgeParam struct {
	Beneficial types.Address `json:"beneficial"`
	Amount     *big.Int      `json:"amount"`
	PType      uint8         `json:"pType"`
	NEP5TxId   string        `json:"nep5TxId"`
}

type NEP5PledgeInfo struct {
	PType         uint8
	Amount        *big.Int
	WithdrawTime  int64
	Beneficial    types.Address
	PledgeAddress types.Address
	NEP5TxId      string
}

// string to pledge type
func StringToPledgeType(sType string) (PledgeType, error) {
	switch strings.ToLower(sType) {
	case "network", "confidant":
		return Network, nil
	case "vote":
		return Vote, nil
	case "oracle":
		return Oracle, nil
	default:
		return Invalid, fmt.Errorf("unsupport type: %s", sType)
	}
}

// ParsePledgeParam convert data to PledgeParam
func ParsePledgeParam(data []byte) (*PledgeParam, error) {
	if len(data) == 0 {
		return nil, errors.New("pledge param data is nil")
	}
	param := new(PledgeParam)
	if err := NEP5PledgeABI.UnpackMethod(param, MethodNEP5Pledge, data); err == nil {
		return param, nil
	} else {
		return nil, err
	}
}

// ParsePledgeInfo convert data to NEP5PledgeInfo
func ParsePledgeInfo(data []byte) (*NEP5PledgeInfo, error) {
	if len(data) == 0 {
		return nil, errors.New("pledge info data is nil")
	}

	info := new(NEP5PledgeInfo)
	if err := NEP5PledgeABI.UnpackVariable(info, VariableNEP5PledgeInfo, data); err == nil {
		return info, nil
	} else {
		return nil, err
	}
}

func GetPledgeKey(addr types.Address, beneficial types.Address, neoTxId string) []byte {
	result := []byte(beneficial[:])
	result = append(result, addr[:]...)
	result = append(result, []byte(neoTxId)...)
	return result
}

func GetTotalPledgeAmount(ctx *vmstore.VMContext) *big.Int {
	var result uint64
	logger := log.NewLogger("GetTotalPledgeAmount")
	defer func() {
		_ = logger.Sync()
	}()

	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				result, _ = util.SafeAdd(pledgeInfo.Amount.Uint64(), result)
			} else {
				logger.Error(err)
			}
		}
		return nil
	})
	if err != nil {
		logger.Error(err)
	}

	return new(big.Int).SetUint64(result)
}

func GetPledgeBeneficialAmount(ctx *vmstore.VMContext, beneficial types.Address, pType uint8) *big.Int {
	var result uint64
	logger := log.NewLogger("GetPledgeBeneficialAmount")
	defer func() {
		_ = logger.Sync()
	}()

	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key[(types.AddressSize+1):], beneficial[:]) && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				if pledgeInfo.PType == pType {
					result, _ = util.SafeAdd(pledgeInfo.Amount.Uint64(), result)
				}
			} else {
				logger.Error(err)
			}
		}
		return nil
	})
	if err != nil {
		logger.Error(err)
	}

	return new(big.Int).SetUint64(result)
}

func GetPledgeBeneficialTotalAmount(ctx *vmstore.VMContext, beneficial types.Address) (*big.Int, error) {
	var result uint64
	logger := log.NewLogger("GetPledgeBeneficialTotalAmount")
	defer func() {
		_ = logger.Sync()
	}()

	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key[(types.AddressSize+1):], beneficial[:]) && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				result, _ = util.SafeAdd(pledgeInfo.Amount.Uint64(), result)
			} else {
				logger.Error(err)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return new(big.Int).SetUint64(result), nil
}

// GetPledgeInfos get pledge info list by pledge address
func GetPledgeInfos(ctx *vmstore.VMContext, addr types.Address) ([]*NEP5PledgeInfo, *big.Int) {
	var result uint64
	var piList []*NEP5PledgeInfo

	logger := log.NewLogger("GetPledgeInfos")
	defer func() {
		_ = logger.Sync()
	}()

	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key[(types.AddressSize*2+1):], addr[:]) && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				piList = append(piList, pledgeInfo)
				result, _ = util.SafeAdd(pledgeInfo.Amount.Uint64(), result)
			} else {
				logger.Error(err)
			}
		}
		return nil
	})

	if err != nil {
		logger.Error(err)
	}
	sort.Slice(piList, func(i, j int) bool { return piList[i].WithdrawTime < piList[j].WithdrawTime })
	return piList, new(big.Int).SetUint64(result)
}

// GetPledgeInfos get pledge info list by pledge address
func GetBeneficialInfos(ctx *vmstore.VMContext, addr types.Address) ([]*NEP5PledgeInfo, *big.Int) {
	var result uint64
	var piList []*NEP5PledgeInfo

	logger := log.NewLogger("GetPledgeInfos")
	defer func() {
		_ = logger.Sync()
	}()

	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key[(types.AddressSize+1):], addr[:]) && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				piList = append(piList, pledgeInfo)
				result, _ = util.SafeAdd(pledgeInfo.Amount.Uint64(), result)
			} else {
				logger.Error(err)
			}
		}
		return nil
	})

	if err != nil {
		logger.Error(err)
	}
	sort.Slice(piList, func(i, j int) bool { return piList[i].WithdrawTime < piList[j].WithdrawTime })
	return piList, new(big.Int).SetUint64(result)
}

//GetBeneficialPledgeInfos get pledge info by beneficial address and pledge type
func GetBeneficialPledgeInfos(ctx *vmstore.VMContext, beneficial types.Address, pType PledgeType) ([]*NEP5PledgeInfo, *big.Int) {
	var result uint64
	var piList []*NEP5PledgeInfo

	logger := log.NewLogger("GetBeneficialPledgeInfos")
	defer func() {
		_ = logger.Sync()
	}()

	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key[(types.AddressSize+1):], beneficial[:]) && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				if pledgeInfo.PType == uint8(pType) {
					piList = append(piList, pledgeInfo)
					result, _ = util.SafeAdd(pledgeInfo.Amount.Uint64(), result)
				}
			} else {
				logger.Error(err)
			}
		}
		return nil
	})

	if err != nil {
		logger.Error(err)
	}

	return piList, new(big.Int).SetUint64(result)
}

type PledgeResult struct {
	Key        []byte
	PledgeInfo *NEP5PledgeInfo
}

func SearchBeneficialPledgeInfo(ctx *vmstore.VMContext, param *WithdrawPledgeParam) []*PledgeResult {
	logger := log.NewLogger("GetBeneficialPledgeInfos")
	defer func() {
		_ = logger.Sync()
	}()
	var result []*PledgeResult
	now := common.TimeNow().Unix()
	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key[(types.AddressSize+1):], param.Beneficial[:]) && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				if pledgeInfo.PType == param.PType &&
					pledgeInfo.Amount.String() == param.Amount.String() && now >= pledgeInfo.WithdrawTime {
					result = append(result, &PledgeResult{Key: key, PledgeInfo: pledgeInfo})
				}
			} else {
				logger.Error(err)
			}
		}
		return nil
	})

	if err != nil {
		logger.Error(err)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].PledgeInfo.WithdrawTime < result[j].PledgeInfo.WithdrawTime })
	return result
}

func SearchPledgeInfoWithNEP5TxId(ctx *vmstore.VMContext, param *WithdrawPledgeParam) *PledgeResult {
	logger := log.NewLogger("GetBeneficialPledgeInfos")
	defer func() {
		_ = logger.Sync()
	}()
	var result *PledgeResult
	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key[(types.AddressSize+1):], param.Beneficial[:]) && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				if pledgeInfo.PType == param.PType &&
					pledgeInfo.Amount.String() == param.Amount.String() && pledgeInfo.NEP5TxId == param.NEP5TxId {
					result = &PledgeResult{Key: key, PledgeInfo: pledgeInfo}
				}
			} else {
				logger.Error(err)
			}
		}
		return nil
	})

	if err != nil {
		logger.Error(err)
	}

	return result
}

// FIXME: can't sleep to waiting...
func SearchBeneficialPledgeInfoByTxId(ctx *vmstore.VMContext, param *WithdrawPledgeParam) *PledgeResult {
	result := searchBeneficialPledgeInfoByTxId(ctx, param)
	if result != nil {
		return result
	}
	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		result := searchBeneficialPledgeInfoByTxId(ctx, param)
		if result != nil {
			return result
		}
	}
	return nil
}

func searchBeneficialPledgeInfoByTxId(ctx *vmstore.VMContext, param *WithdrawPledgeParam) *PledgeResult {
	logger := log.NewLogger("GetBeneficialPledgeInfos")
	defer func() {
		_ = logger.Sync()
	}()
	result := new(PledgeResult)
	now := common.TimeNow().Unix()
	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key[(types.AddressSize+1):], param.Beneficial[:]) && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				if pledgeInfo.PType == param.PType && pledgeInfo.Amount.String() == param.Amount.String() &&
					now >= pledgeInfo.WithdrawTime && pledgeInfo.NEP5TxId == param.NEP5TxId {
					result.Key = append(result.Key, key...)
					result.PledgeInfo = pledgeInfo
				} else {
					if pledgeInfo.NEP5TxId == param.NEP5TxId {
						logger.Errorf("data from param, %s, %s, %s, %s, data from pledgeInfo, %s, %s, %s, %s, now is %s, withdraw time is %s",
							param.PType, param.Amount.String(), param.Beneficial.String(), param.NEP5TxId,
							pledgeInfo.PType, pledgeInfo.Amount.String(), pledgeInfo.Beneficial.String(), pledgeInfo.NEP5TxId,
							now, pledgeInfo.WithdrawTime,
						)
					}
				}
			} else {
				logger.Error(err)
			}
		}
		return nil
	})

	if err != nil {
		logger.Error(err)
	}
	if result.Key == nil {
		return nil
	}
	return result
}

func SearchBeneficialPledgeInfoIgnoreWithdrawTime(ctx *vmstore.VMContext, param *WithdrawPledgeParam) []*PledgeResult {
	logger := log.NewLogger("SearchBeneficialPledgeInfoIgnoreWithdrawTime")
	defer func() {
		_ = logger.Sync()
	}()
	var result []*PledgeResult
	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key[(types.AddressSize+1):], param.Beneficial[:]) && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				if pledgeInfo.PType == param.PType && pledgeInfo.Amount.String() == param.Amount.String() {
					result = append(result, &PledgeResult{Key: key, PledgeInfo: pledgeInfo})
				}
			} else {
				logger.Error(err)
			}
		}
		return nil
	})

	if err != nil {
		logger.Error(err)
	}

	return result
}

func SearchAllPledgeInfos(ctx *vmstore.VMContext) ([]*NEP5PledgeInfo, error) {
	var result []*NEP5PledgeInfo
	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)

			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				result = append(result, pledgeInfo)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}
