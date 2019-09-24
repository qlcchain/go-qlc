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
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/abi"
)

const (
	jsonDestroy = `[
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
	KeySize             = types.AddressSize + types.HashSize + 1
)

var (
	BlackHoleABI, _ = abi.JSONToABIContract(strings.NewReader(jsonDestroy))
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
		var data []byte

		data = append(data, param.Owner[:]...)
		data = append(data, param.Previous[:]...)
		data = append(data, param.Token[:]...)
		data = append(data, param.Amount.Bytes()...)
		var sig types.Signature
		copy(sig[:], ed25519.Sign(acc.PrivateKey(), data))
		return sig, nil
	} else {
		return types.ZeroSignature, fmt.Errorf("invalid address, exp: %s, act: %s",
			param.Owner.String(), acc.Address().String())
	}
}

// Verify destroy params
func (param *DestroyParam) Verify() (bool, error) {
	if param.Owner.IsZero() {
		return false, errors.New("invalid account")
	}

	if param.Previous.IsZero() {
		return false, errors.New("invalid previous")
	}

	if param.Token != common.GasToken() {
		return false, errors.New("invalid token to be destroyed")
	}

	if param.Amount == nil || param.Amount.Sign() <= 0 {
		return false, errors.New("invalid amount")
	}

	var data []byte

	data = append(data, param.Owner[:]...)
	data = append(data, param.Previous[:]...)
	data = append(data, param.Token[:]...)
	data = append(data, param.Amount.Bytes()...)

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
				Type:           types.ContractSend,
				Token:          tm.Type,
				Address:        param.Owner,
				Balance:        tm.Balance.Sub(types.Balance{Int: param.Amount}),
				Vote:           types.ZeroBalance,
				Network:        types.ZeroBalance,
				Oracle:         types.ZeroBalance,
				Storage:        types.ZeroBalance,
				Previous:       param.Previous,
				Link:           types.Hash(types.BlackHoleAddress),
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
func GetTotalDestroyInfo(ctx *vmstore.VMContext, addr *types.Address) (types.Balance, error) {
	logger := log.NewLogger("GetRewardsDetail")
	defer func() {
		_ = logger.Sync()
	}()

	result := types.ZeroBalance
	if err := ctx.Iterator(addr[:], func(key []byte, value []byte) error {
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
func GetDestroyInfoDetail(ctx *vmstore.VMContext, addr *types.Address) ([]*DestroyInfo, error) {
	logger := log.NewLogger("GetRewardsDetail")
	defer func() {
		_ = logger.Sync()
	}()

	var result []*DestroyInfo
	if err := ctx.Iterator(addr[:], func(key []byte, value []byte) error {
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
