/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/ledger"
	"math/big"
	"strings"

	"github.com/qlcchain/go-qlc/vm/abi"
)

const (
	JsonQGasSwap = `
	[
		{
		  "type": "function",
		  "name": "QGasPledge",
		  "inputs": [
		    {
		      "name": "pledgeAddress",
		      "type": "address"
		    },
		    {
		      "name": "amount",
		      "type": "uint256"
		    },
		    {
		      "name": "toAddress",
		      "type": "address"
		    }
		  ]
		},
		{
		  "type": "function",
		  "name": "QGasWithdraw",
		  "inputs": [
		    {
		      "name": "withdrawAddress",
		      "type": "address"
		    },
		    {
		      "name": "amount",
		      "type": "uint256"
		    },
		    {
		      "name": "fromAddress",
		      "type": "address"
		    },
		    {
		      "name": "linkHash",
		      "type": "hash"
		    }
		  ]
		},
		{
		  "type": "variable",
		  "name": "QGasSwapInfo",
		  "inputs": [
		    {
		      "name": "amount",
		      "type": "uint256"
		    },
		    {
		      "name": "fromAddress",
		      "type": "address"
		    },
		    {
		      "name": "toAddress",
		      "type": "address"
		    },
		    {
		      "name": "sendHash",
		      "type": "hash"
		    },
		    {
		      "name": "linkHash",
		      "type": "hash"
		    },
		    {
		      "name": "swapType",
		      "type": "uint32"
		    },
		    {
		      "name": "time",
		      "type": "int64"
		    }
		  ]
		}
	]`

	MethodQGasPledge     = "QGasPledge"
	MethodQGasWithdraw   = "QGasWithdraw"
	VariableQGasSwapInfo = "QGasSwapInfo"
)

var (
	QGasSwapABI, _ = abi.JSONToABIContract(strings.NewReader(JsonQGasSwap))
)

type QGasPledgeParam struct {
	PledgeAddress types.Address `msg:"pledgeAddress" json:"pledgeAddress"`
	Amount        *big.Int      `msg:"amount" json:"amount"`
	ToAddress     types.Address `msg:"toAddress" json:"toAddress"`
}

func (q *QGasPledgeParam) ToABI() ([]byte, error) {
	return QGasSwapABI.PackMethod(MethodQGasPledge, q.PledgeAddress, q.Amount, q.ToAddress)
}

func ParseQGasPledgeParam(data []byte) (*QGasPledgeParam, error) {
	if len(data) == 0 {
		return nil, errors.New("qgas pledge data is nil")
	}

	info := new(QGasPledgeParam)
	if err := QGasSwapABI.UnpackMethod(info, MethodQGasPledge, data); err == nil {
		return info, nil
	} else {
		return nil, err
	}
}

func (q *QGasPledgeParam) Verify() (bool, error) {
	if q.Amount == nil || q.Amount.Cmp(big.NewInt(0)) <= 0 {
		return false, errors.New("invalid amount to pledge")
	}
	return true, nil
}

type QGasWithdrawParam struct {
	WithdrawAddress types.Address `msg:"withdrawAddress" json:"withdrawAddress"`
	Amount          *big.Int      `msg:"amount" json:"amount"`
	FromAddress     types.Address `msg:"fromAddress" json:"fromAddress"`
	LinkHash        types.Hash    `msg:"linkHash" json:"linkHash"`
}

func (q *QGasWithdrawParam) ToABI() ([]byte, error) {
	return QGasSwapABI.PackMethod(MethodQGasWithdraw, q.WithdrawAddress, q.Amount, q.FromAddress, q.LinkHash)
}

func ParseQGasWithdrawParam(data []byte) (*QGasWithdrawParam, error) {
	if len(data) == 0 {
		return nil, errors.New("qgas withdraw data is nil")
	}

	info := new(QGasWithdrawParam)
	if err := QGasSwapABI.UnpackMethod(info, MethodQGasWithdraw, data); err == nil {
		return info, nil
	} else {
		return nil, err
	}
}

func (q *QGasWithdrawParam) Verify() (bool, error) {
	if q.Amount == nil || q.Amount.Cmp(big.NewInt(0)) <= 0 {
		return false, errors.New("invalid amount to withdraw")
	}
	return true, nil
}

func GetQGasSwapKey(user types.Address, txHash types.Hash) []byte {
	result := user[:]
	result = append(result, txHash[:]...)
	return result
}

const (
	QGasPledge uint32 = iota
	QGasWithdraw
	QGasAll
)

func QGasSwapTypeToString(typ uint32) string {
	switch typ {
	case QGasPledge:
		return "Pledge"
	case QGasWithdraw:
		return "Withdraw"
	case QGasAll:
		return "All"
	}
	return ""
}

type QGasSwapInfo struct {
	Amount      *big.Int
	FromAddress types.Address
	ToAddress   types.Address
	SendHash    types.Hash
	LinkHash    types.Hash
	SwapType    uint32
	Time        int64
}

func (info *QGasSwapInfo) ToABI() ([]byte, error) {
	return QGasSwapABI.PackVariable(VariableQGasSwapInfo, info.Amount, info.FromAddress,
		info.ToAddress, info.SendHash, info.LinkHash, info.SwapType, info.Time)
}

func ParseQGasSwapInfo(data []byte) (*QGasSwapInfo, error) {
	if len(data) == 0 {
		return nil, errors.New("swap info data is nil")
	}

	info := new(QGasSwapInfo)
	if err := QGasSwapABI.UnpackVariable(info, VariableQGasSwapInfo, data); err == nil {
		return info, nil
	} else {
		return nil, err
	}
}

func GetQGasSwapInfos(store ledger.Store, address types.Address, typ uint32) ([]*QGasSwapInfo, error) {
	var piList []*QGasSwapInfo
	iterator := store.NewVMIterator(&contractaddress.QGasSwapAddress)
	prefixKey := make([]byte, 0)
	if !address.IsZero() {
		prefixKey = append(prefixKey, address.Bytes()...)
	}
	if err := iterator.Next(prefixKey, func(key []byte, value []byte) error {
		if len(key) >= keySize && len(value) > 0 {
			if pledgeInfo, err := ParseQGasSwapInfo(value); err == nil {
				switch typ {
				case QGasPledge:
					if pledgeInfo.SwapType == QGasPledge {
						piList = append(piList, pledgeInfo)
					}
				case QGasWithdraw:
					if pledgeInfo.SwapType == QGasWithdraw {
						piList = append(piList, pledgeInfo)
					}
				case QGasAll:
					piList = append(piList, pledgeInfo)
				}
			} else {
				return fmt.Errorf("parse qgas swap info: %s", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return piList, nil
}

func GetQGasSwapAmount(store ledger.Store, address types.Address) (map[string]*big.Int, error) {
	iterator := store.NewVMIterator(&contractaddress.QGasSwapAddress)
	prefixKey := make([]byte, 0)
	if !address.IsZero() {
		prefixKey = append(prefixKey, address.Bytes()...)
	}
	var pledgeAmount uint64
	var withdrawAmount uint64
	//var allAmount uint64
	if err := iterator.Next(prefixKey, func(key []byte, value []byte) error {
		if len(key) >= keySize && len(value) > 0 {
			if swapInfo, err := ParseQGasSwapInfo(value); err == nil {
				if swapInfo.SwapType == QGasPledge {
					pledgeAmount, _ = util.SafeAdd(swapInfo.Amount.Uint64(), pledgeAmount)
				} else {
					withdrawAmount, _ = util.SafeAdd(swapInfo.Amount.Uint64(), withdrawAmount)
				}
				//allAmount, _ = util.SafeAdd(swapInfo.Amount.Uint64(), allAmount)
			} else {
				return fmt.Errorf("parse qgas swap info: %s", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]*big.Int{
		QGasSwapTypeToString(QGasPledge):   new(big.Int).SetUint64(pledgeAmount),
		QGasSwapTypeToString(QGasWithdraw): new(big.Int).SetUint64(withdrawAmount),
		//QGasSwapTypeToString(QGasAll):      new(big.Int).SetUint64(allAmount),
	}, nil
}
