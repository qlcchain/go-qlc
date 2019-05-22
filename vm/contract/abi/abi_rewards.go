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

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/log"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/vm/abi"
)

const (
	jsonRewards = `[
		{"type":"function","name":"AirdropRewards","inputs":[{"name":"NEP5TxId","type":"bytes32"},{"name":"sHeader","type":"bytes32"},{"name":"rHeader","type":"bytes32"},{"name":"amount","type":"uint256"},{"name":"sign","type":"bytes64"}]},
		{"type":"function","name":"ConfidantRewards","inputs":[{"name":"id","type":"byte32"},{"name":"sHeader","type":"bytes32"},{"name":"rHeader","type":"bytes32"},{"name":"amount","type":"uint256"},{"name":"sign","type":"bytes64"}]},
		{"type":"variable","name":"rewardsInfo","inputs":[{"name":"type","type":"uint8"},{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"sHeader","type":"bytes32"},{"name":"rHeader","type":"bytes32"},{"name":"amount","type":"uint256"}]}
	]`

	MethodNameAirdropRewards   = "AirdropRewards"
	MethodNameConfidantRewards = "ConfidantRewards"
	VariableNameRewards        = "rewardsInfo"
)

var (
	RewardsABI, _ = abi.JSONToABIContract(strings.NewReader(jsonRewards))
)

const (
	Confidant = iota
	Rewards
)

type RewardsParam struct {
	Id      []byte
	SHeader []byte
	RHeader []byte
	Amount  *big.Int
	Sign    []byte
}

func (ap *RewardsParam) Verify(address types.Address) (bool, error) {
	h1, err := types.BytesToHash(ap.SHeader)
	if err != nil {
		return false, err
	}

	_, err = types.BytesToHash(ap.RHeader)
	if err != nil {
		return false, err
	}

	if len(ap.Id) > 0 && !h1.IsZero() && ap.Amount != nil {
		var data []byte
		data = append(data, ap.Id...)
		data = append(data, ap.SHeader...)
		data = append(data, ap.RHeader...)
		data = append(data, ap.Amount.Bytes()...)

		h := types.HashData(data)

		//verify sign
		if !address.Verify(h[:], ap.Sign) {
			return false, fmt.Errorf("invalid sign")
		} else {
			return true, nil
		}
	} else {
		return false, fmt.Errorf("invalid param")
	}
}

type RewardsInfo struct {
	Type    uint8
	From    *types.Address
	To      *types.Address
	SHeader []byte
	RHeader []byte
	Amount  *big.Int
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

func GetRewardsKey(txId, sHeader, rHeader []byte) []byte {
	result := []byte(txId)
	result = append(result, sHeader...)
	result = append(result, rHeader...)
	return result
}

func GetConfidantKey(address types.Address, txId, sHeader, rHeader []byte) []byte {
	result := []byte(address[:])
	result = append(result, txId[:]...)
	result = append(result, sHeader...)
	result = append(result, rHeader...)

	return result
}

func GetRewardsDetail(ctx *vmstore.VMContext, txId string) ([]*RewardsInfo, error) {
	logger := log.NewLogger("GetPledgeBeneficialAmount")
	defer func() {
		_ = logger.Sync()
	}()

	id, err := hex.DecodeString(txId)
	if err != nil {
		return nil, err
	}
	var result []*RewardsInfo
	if err := ctx.Iterator(types.RewardsAddress[:], func(key []byte, value []byte) error {
		if bytes.HasPrefix(key[1:], id) {
			info := new(RewardsInfo)
			if err := NEP5PledgeABI.UnpackVariable(info, VariableNameRewards, value); err == nil && info.Type == uint8(Rewards) {
				result = append(result, info)
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

func GetRewards(ctx *vmstore.VMContext, txId string) (*big.Int, error) {
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
	logger := log.NewLogger("GetPledgeBeneficialAmount")
	defer func() {
		_ = logger.Sync()
	}()

	result := make(map[string][]*RewardsInfo)

	if err := ctx.Iterator(types.RewardsAddress[:], func(key []byte, value []byte) error {
		if bytes.HasPrefix(key[1:], confidant[:]) {
			info := new(RewardsInfo)
			if err := NEP5PledgeABI.UnpackVariable(info, VariableNameRewards, value); err == nil && info.Type == uint8(Confidant) {
				s := hex.EncodeToString(key[types.AddressSize+1 : types.AddressSize+types.HashSize])
				if infos, ok := result[s]; ok {
					result[s] = append(infos, info)
				} else {
					result[s] = []*RewardsInfo{info}
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
	if infos, err := GetConfidantRewordsDetail(ctx, confidant); err == nil {
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
