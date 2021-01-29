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
	"math/big"
	"strings"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	JsonDestroy = `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      { "name": "owner", "type": "address" },
      { "name": "previous", "type": "hash" },
      { "name": "token", "type": "tokenId" },
      { "name": "amount", "type": "uint256" },
      { "name": "sign", "type": "signature" }
    ]
  },
  {
    "type": "variable",
    "name": "destroyInfo",
    "inputs": [
      { "name": "owner", "type": "address" },
      { "name": "previous", "type": "hash" },
      { "name": "token", "type": "tokenId" },
      { "name": "amount", "type": "uint256" },
      { "name": "timeStamp", "type": "int64" }
    ]
  }
]
`

	MethodNameDestroy   = "Destroy"
	VariableDestroyInfo = "destroyInfo"
	KeySize             = types.AddressSize + types.HashSize
)

var (
	BlackHoleABI, _ = abi.JSONToABIContract(strings.NewReader(JsonDestroy))
)

type DestroyParam struct {
	Owner    types.Address   `json:"owner"`
	Previous types.Hash      `json:"previous"`
	Token    types.Hash      `json:"token"`
	Amount   *big.Int        `json:"amount"`
	Sign     types.Signature `json:"signature"`
}

func (param *DestroyParam) Signature(acc *types.Account) (types.Signature, error) {
	if acc.Address() == param.Owner {
		data := param.toBytes()
		sig := acc.SignData(data)
		return sig, nil
	} else {
		return types.ZeroSignature, fmt.Errorf("invalid address, exp: %s, act: %s",
			param.Owner.String(), acc.Address().String())
	}
}

func (param *DestroyParam) String() string {
	return util.ToIndentString(param)
}

func (param *DestroyParam) toBytes() []byte {
	var data []byte

	data = append(data, param.Owner[:]...)
	data = append(data, param.Previous[:]...)
	data = append(data, param.Token[:]...)
	data = append(data, param.Amount.Bytes()...)
	return data
}

// Verify destroy params
func (param *DestroyParam) Verify() (bool, error) {
	if param.Owner.IsZero() {
		return false, errors.New("invalid account")
	}

	if param.Previous.IsZero() {
		return false, errors.New("invalid previous")
	}

	if param.Token != config.GasToken() {
		return false, errors.New("invalid token to be destroyed")
	}

	if param.Amount == nil || param.Amount.Sign() <= 0 {
		return false, errors.New("invalid amount")
	}

	data := param.toBytes()

	return param.Owner.Verify(data, param.Sign[:]), nil
}

type DestroyInfo struct {
	Owner     types.Address `json:"owner"`
	Previous  types.Hash    `json:"previous"`
	Token     types.Hash    `json:"token"`
	Amount    *big.Int      `json:"amount"`
	TimeStamp int64         `json:"timestamp"`
}

func PackSendBlock(ctx *vmstore.VMContext, param *DestroyParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, errors.New("invalid input param")
	}

	if isVerified, err := param.Verify(); err != nil {
		return nil, err
	} else if !isVerified {
		return nil, errors.New("invalid sign of param")
	}

	if tm, err := ctx.GetTokenMeta(param.Owner, param.Token); err != nil {
		return nil, err
	} else {
		if tm.Balance.Compare(types.Balance{Int: param.Amount}) == types.BalanceCompSmaller {
			return nil, fmt.Errorf("not enough balance, [%s] of [%s]", param.Amount.String(), tm.Balance.String())
		}

		if singedData, err := BlackHoleABI.PackMethod(MethodNameDestroy, param.Owner, param.Previous, param.Token,
			param.Amount, param.Sign); err == nil {
			return &types.StateBlock{
				Type:    types.ContractSend,
				Token:   tm.Type,
				Address: param.Owner,
				Balance: tm.Balance.Sub(types.Balance{Int: param.Amount}),
				//Vote:           types.ZeroBalance,
				//Network:        types.ZeroBalance,
				//Oracle:         types.ZeroBalance,
				//Storage:        types.ZeroBalance,
				Previous:       param.Previous,
				Link:           types.Hash(contractaddress.BlackHoleAddress),
				Representative: tm.Representative,
				Data:           singedData,
				Timestamp:      common.TimeNow().Unix(),
			}, nil
		} else {
			return nil, err
		}
	}
}

// ParseDestroyInfo decode data into `DestroyInfo`
func ParseDestroyInfo(data []byte) (*DestroyInfo, error) {
	if len(data) == 0 {
		return nil, errors.New("token info data is nil")
	}
	di := new(DestroyInfo)

	if err := BlackHoleABI.UnpackVariable(di, VariableDestroyInfo, data); err == nil {
		return di, nil
	} else {
		return nil, err
	}
}

// GetTotalDestroyInfo query all destroyed GQAS by account
func GetTotalDestroyInfo(store ledger.Store, addr *types.Address) (types.Balance, error) {
	logger := log.NewLogger("GetRewardsDetail")
	defer func() {
		_ = logger.Sync()
	}()

	result := types.ZeroBalance
	iterator := store.NewVMIterator(&contractaddress.BlackHoleAddress)
	if err := iterator.Next(addr[:], func(key []byte, value []byte) error {
		if len(key) == KeySize && len(value) > 0 {
			if di, err := ParseDestroyInfo(value); err == nil {
				result = result.Add(types.Balance{Int: di.Amount})
			} else {
				logger.Error(err)
			}
		}
		return nil
	}); err == nil {
		return result, nil
	} else {
		return types.ZeroBalance, err
	}
}

// GetDestroyInfoDetail query destroyed GQAS detail by account
func GetDestroyInfoDetail(store ledger.Store, addr *types.Address) ([]*DestroyInfo, error) {
	logger := log.NewLogger("GetRewardsDetail")
	defer func() {
		_ = logger.Sync()
	}()

	var result []*DestroyInfo
	iterator := store.NewVMIterator(&contractaddress.BlackHoleAddress)
	if err := iterator.Next(addr[:], func(key []byte, value []byte) error {
		if len(key) == KeySize && len(value) > 0 {
			if di, err := ParseDestroyInfo(value); err == nil {
				result = append(result, di)
			} else {
				logger.Error(err)
			}
		} else {
			logger.Infof("%d of %d ", len(key), KeySize)
		}
		return nil
	}); err == nil {
		return result, nil
	} else {
		return nil, err
	}
}
