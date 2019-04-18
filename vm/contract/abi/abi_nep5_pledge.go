/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"bytes"
	"encoding/json"
	"math/big"
	"strings"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	jsonNEP5Pledge = `
	[
		{"type":"function","name":"NEP5Pledge", "inputs":[{"name":"beneficial","type":"address"},{"name":"pledgeAddress","type":"address"},{"name":"pType","type":"uint8"},{"name":"NEP5TxId","type":"string"}]},
		{"type":"function","name":"WithdrawNEP5Pledge","inputs":[{"name":"beneficial","type":"address"},{"name":"amount","type":"uint256"},{"name":"pType","type":"uint8"}]},
		{"type":"variable","name":"nep5PledgeInfo","inputs":[{"name":"pType","type":"uint8"},{"name":"amount","type":"uint256"},{"name":"withdrawTime","type":"int64"},{"name":"beneficial","type":"address"},{"name":"pledgeAddress","type":"address"},{"name":"NEP5TxId","type":"string"}]}
	]`

	MethodNEP5Pledge         = "NEP5Pledge"
	MethodWithdrawNEP5Pledge = "WithdrawNEP5Pledge"
	VariableNEP5PledgeInfo   = "nep5PledgeInfo"
)

var (
	NEP5PledgeABI, _ = abi.JSONToABIContract(strings.NewReader(jsonNEP5Pledge))
)

type PledgeType uint8

const (
	Network PledgeType = iota
	Vote
	Storage
	Oracle
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
}

type NEP5PledgeInfo struct {
	PType         uint8
	Amount        *big.Int
	WithdrawTime  int64
	Beneficial    types.Address
	PledgeAddress types.Address
	NEP5TxId      string
}

func GetPledgeKey(addr types.Address, beneficial types.Address, time int64) []byte {
	result := []byte(beneficial[:])
	result = append(result, addr[:]...)
	result = append(result, util.Int2Bytes(time)...)
	return result
}

func GetPledgeBeneficialAmount(ctx *vmstore.VMContext, beneficial types.Address, pType PledgeType) *big.Int {
	var result uint64
	logger := log.NewLogger("GetPledgeBeneficialAmount")
	defer func() {
		_ = logger.Sync()
	}()

	tmp := make(map[types.Hash]NEP5PledgeInfo)
	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key, beneficial[:]) {
			if err := json.Unmarshal(value, tmp); err == nil {
				for _, v := range tmp {
					if v.PType == uint8(pType) {
						result, _ = util.SafeAdd(v.Amount.Uint64(), result)
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

	return new(big.Int).SetUint64(result)
}

func GetPledgeBeneficialTotalAmount(ctx *vmstore.VMContext, beneficial types.Address) *big.Int {
	var result uint64
	logger := log.NewLogger("GetPledgeBeneficialTotalAmount")
	defer func() {
		_ = logger.Sync()
	}()

	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key, beneficial[:]) {
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

// GetPledgeInfos get pledge info list by pledge address
func GetPledgeInfos(ctx *vmstore.VMContext, addr types.Address) ([]*NEP5PledgeInfo, *big.Int) {
	var result uint64
	var piList []*NEP5PledgeInfo

	logger := log.NewLogger("GetPledgeInfos")
	defer func() {
		_ = logger.Sync()
	}()

	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasSuffix(key[types.AddressSize:], addr[:]) && len(value) > 0 {
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

	return piList, new(big.Int).SetUint64(result)
}

//GetBeneficialPledgeInfos get pledge info by beneficial address and pledge type
func GetBeneficialPledgeInfos(ctx *vmstore.VMContext, beneficial types.Address, pType uint8) ([]*NEP5PledgeInfo, *big.Int) {
	var result uint64
	var piList []*NEP5PledgeInfo

	logger := log.NewLogger("GetBeneficialPledgeInfos")
	defer func() {
		_ = logger.Sync()
	}()

	err := ctx.Iterator(types.NEP5PledgeAddress[:], func(key []byte, value []byte) error {
		if len(key) > 2*types.AddressSize && bytes.HasPrefix(key, beneficial[:]) && len(value) > 0 {
			pledgeInfo := new(NEP5PledgeInfo)
			if err := NEP5PledgeABI.UnpackVariable(pledgeInfo, VariableNEP5PledgeInfo, value); err == nil {
				if pledgeInfo.PType == pType {
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
	now := time.Now().UTC().Unix()
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

	return result
}
