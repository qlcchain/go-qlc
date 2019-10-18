/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	JsonRewards = `[
		{"type":"function","name":"AirdropRewards","inputs":[{"name":"id","type":"hash"},{"name":"beneficial","type":"address"},{"name":"txHeader","type":"hash"},{"name":"rxHeader","type":"hash"},{"name":"amount","type":"uint256"},{"name":"sign","type":"signature"}]  },
		{"type":"function","name":"UnsignedAirdropRewards","inputs":[{"name":"id","type":"hash"},{"name":"beneficial","type":"address"},{"name":"txHeader","type":"hash"},{"name":"rxHeader","type":"hash"},{"name":"amount","type":"uint256"}]  },
		{"type":"function","name":"ConfidantRewards","inputs":[{"name":"id","type":"bytes32"},{"name":"beneficial","type":"address"},{"name":"txHeader","type":"hash"},{"name":"rxHeader","type":"hash"},{"name":"amount","type":"uint256"},{"name":"sign","type":"signature"}]  },
		{"type":"function","name":"UnsignedConfidantRewards","inputs":[{"name":"id","type":"hash"},{"name":"beneficial","type":"address"},{"name":"txHeader","type":"hash"},{"name":"rxHeader","type":"hash"},{"name":"amount","type":"uint256"}]  },
		{"type":"variable","name":"rewardsInfo","inputs":[{"name":"type","type":"uint8"},{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"txHeader","type":"hash"},{"name":"rxHeader","type":"hash"},{"name":"amount","type":"uint256"}]  }
	]
`

	MethodNameUnsignedAirdropRewards   = "UnsignedAirdropRewards"
	MethodNameAirdropRewards           = "AirdropRewards"
	MethodNameUnsignedConfidantRewards = "UnsignedConfidantRewards"
	MethodNameConfidantRewards         = "ConfidantRewards"
	VariableNameRewards                = "rewardsInfo"
)

var (
	RewardsABI, _ = abi.JSONToABIContract(strings.NewReader(JsonRewards))
)

const (
	Confidant = iota
	Rewards
)

type RewardsParam struct {
	Id         types.Hash      `json:"id"`
	Beneficial types.Address   `json:"beneficial"`
	TxHeader   types.Hash      `json:"txHeader"`
	RxHeader   types.Hash      `json:"rxHeader"`
	Amount     *big.Int        `json:"amount"`
	Sign       types.Signature `json:"signature"`
}

func (ap *RewardsParam) Verify(address types.Address, methodName string) (bool, error) {
	if !ap.Id.IsZero() && !ap.TxHeader.IsZero() && !ap.Beneficial.IsZero() && ap.Amount.Sign() > 0 {
		if data, err := RewardsABI.PackMethod(methodName, ap.Id, ap.Beneficial, ap.TxHeader, ap.RxHeader, ap.Amount); err == nil {
			h := types.HashData(data)

			if address.Verify(h[:], ap.Sign[:]) {
				return true, nil
			} else {
				return false, fmt.Errorf("invalid sign[%s] of hash[%s]", ap.Sign.String(), h.String())
			}
		} else {
			return false, err
		}
	} else {
		return false, fmt.Errorf("invalid Param")
	}
}

type RewardsInfo struct {
	Type     uint8         `json:"type"`
	From     types.Address `json:"from"`
	To       types.Address `json:"to"`
	TxHeader types.Hash    `json:"txHeader"`
	RxHeader types.Hash    `json:"rxHeader"`
	Amount   *big.Int      `json:"amount"`
}

func ParseRewardsInfo(data []byte) (*RewardsInfo, error) {
	if len(data) == 0 {
		return nil, errors.New("pledge info data is nil")
	}

	info := new(RewardsInfo)
	if err := RewardsABI.UnpackVariable(info, VariableNameRewards, data); err == nil {
		return info, nil
	} else {
		return nil, err
	}
}

func GetRewardsKey(txId, txHeader, rxHeader []byte) []byte {
	result := []byte(txId)
	result = append(result, txHeader...)
	result = append(result, rxHeader...)
	return result
}

func GetConfidantKey(address types.Address, txId, txHeader, rxHeader []byte) []byte {
	result := []byte(address[:])
	result = append(result, txId[:]...)
	result = append(result, txHeader...)
	result = append(result, rxHeader...)

	return result
}

func GetRewardsDetail(ctx *vmstore.VMContext, txId string) ([]*RewardsInfo, error) {
	logger := log.NewLogger("GetRewardsDetail")
	defer func() {
		_ = logger.Sync()
	}()

	id, err := hex.DecodeString(txId)
	if err != nil {
		return nil, err
	}
	var result []*RewardsInfo
	if err := ctx.Iterator(types.RewardsAddress[:], func(key []byte, value []byte) error {
		if bytes.HasPrefix(key[types.AddressSize+1:], id) && len(value) > 0 {
			info := new(RewardsInfo)
			if err := RewardsABI.UnpackVariable(info, VariableNameRewards, value); err == nil {
				if isValidContract(ctx, info) {
					if info.Type == uint8(Rewards) {
						result = append(result, info)
					} else {
						logger.Warnf("invalid reward type, %s==>%s", txId, util.ToString(info))
					}
				}
			} else {
				logger.Error(err)
			}
		}
		return nil
	}); err == nil {
		return result, nil
	} else {
		return nil, err
	}
}

func isValidContract(ctx *vmstore.VMContext, reward *RewardsInfo) bool {
	//txHeader := key[0:types.HashSize]
	//rxHeader := key[types.HashSize : types.HashSize+types.HashSize]
	//txHash, err := types.BytesToHash(txHeader)
	//if err != nil {
	//	return false
	//}
	_, err := ctx.GetStateBlock(reward.RxHeader)
	if err != nil {
		return false
	}
	_, err = ctx.GetStateBlock(reward.TxHeader)
	if err != nil {
		return false
	}
	return true
}

func GetTotalRewards(ctx *vmstore.VMContext, txId string) (*big.Int, error) {
	var result uint64
	if infos, err := GetRewardsDetail(ctx, txId); err == nil {
		for _, info := range infos {
			result, _ = util.SafeAdd(result, info.Amount.Uint64())
		}
	} else {
		return nil, err
	}

	return new(big.Int).SetUint64(result), nil
}

func GetConfidantRewordsDetail(ctx *vmstore.VMContext, confidant types.Address) (map[string][]*RewardsInfo, error) {
	logger := log.NewLogger("GetConfidantRewordsDetail")
	defer func() {
		_ = logger.Sync()
	}()

	result := make(map[string][]*RewardsInfo)

	if err := ctx.Iterator(types.RewardsAddress[:], func(key []byte, value []byte) error {
		k := key[types.AddressSize+1:]
		if bytes.HasPrefix(k, confidant[:]) && len(value) > 0 {
			info := new(RewardsInfo)
			if err := RewardsABI.UnpackVariable(info, VariableNameRewards, value); err == nil {
				if isValidContract(ctx, info) {
					if info.Type == uint8(Confidant) {
						s := hex.EncodeToString(k[types.AddressSize : types.AddressSize+types.HashSize])
						if infos, ok := result[s]; ok {
							result[s] = append(infos, info)
						} else {
							result[s] = []*RewardsInfo{info}
						}
					} else {
						logger.Warnf("invalid confidant type, %s==>%s", confidant.String(), util.ToString(info))
					}
				}
			} else {
				logger.Error(err)
			}
		}
		return nil
	}); err == nil {
		return result, nil
	} else {
		return nil, err
	}
}

func GetConfidantRewords(ctx *vmstore.VMContext, confidant types.Address) (map[string]*big.Int, error) {
	if infos, err := GetConfidantRewordsDetail(ctx, confidant); err != nil {
		return nil, err
	} else {
		result := make(map[string]*big.Int)
		var total uint64
		for id, values := range infos {
			total = uint64(0)
			for _, v := range values {
				total, _ = util.SafeAdd(total, v.Amount.Uint64())
			}
			result[id] = new(big.Int).SetUint64(total)
		}
		return result, nil
	}
}
