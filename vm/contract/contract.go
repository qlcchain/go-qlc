/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"errors"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/abi"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

//ContractBlock generated by contract
type ContractBlock struct {
	VMContext *vmstore.VMContext
	Block     *types.StateBlock
	ToAddress types.Address
	BlockType types.BlockType
	Amount    types.Balance
	Token     types.Hash
	Data      []byte
}

type Describe struct {
	withSignature bool
	withPending   bool
}

func (d Describe) WithSignature() bool {
	return d.withSignature
}
func (d Describe) WithPending() bool {
	return d.withPending
}

type WithSignAndPending struct {
}

func (WithSignAndPending) GetDescribe() Describe {
	return Describe{
		withSignature: true,
		withPending:   true,
	}
}

type NoSignWithPending struct {
}

func (NoSignWithPending) GetDescribe() Describe {
	return Describe{
		withSignature: false,
		withPending:   true,
	}
}

type WithSignNoPending struct {
}

func (WithSignNoPending) GetDescribe() Describe {
	return Describe{
		withSignature: true,
		withPending:   false,
	}
}

type InternalContract interface {
	GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error)
	// check status, update state
	DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error)
	// refund data at receive error
	GetRefundData() []byte
	GetDescribe() Describe
}

type ChainContractV1 interface {
	InternalContract
	// DoPending generate pending info from send block
	DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error)
	// ProcessSend verify or update StateBlock.Data
	DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error
}

type qlcchainContractV1 struct {
	m       map[string]ChainContractV1
	abi     abi.ABIContract
	abiJson string
}

var contractCacheV1 = map[types.Address]*qlcchainContractV1{
	types.MintageAddress: {
		map[string]ChainContractV1{
			cabi.MethodNameMintage:         &Mintage{},
			cabi.MethodNameMintageWithdraw: &WithdrawMintage{},
		},
		cabi.MintageABI,
		cabi.JsonMintage,
	},
	types.NEP5PledgeAddress: {
		map[string]ChainContractV1{
			cabi.MethodNEP5Pledge:         &Nep5Pledge{},
			cabi.MethodWithdrawNEP5Pledge: &WithdrawNep5Pledge{},
		},
		cabi.NEP5PledgeABI,
		cabi.JsonNEP5Pledge,
	},
	types.RewardsAddress: {
		map[string]ChainContractV1{
			cabi.MethodNameAirdropRewards:   &AirdropRewords{},
			cabi.MethodNameConfidantRewards: &ConfidantRewards{},
		},
		cabi.RewardsABI,
		cabi.JsonRewards,
	},
}

type ChainContractV2 interface {
	InternalContract
	// ProcessSend verify or update StateBlock.Data
	ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error)
	DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error)
}

type qlcchainContractV2 struct {
	m       map[string]ChainContractV2
	abi     abi.ABIContract
	abiJson string
}

var contractCacheV2 = map[types.Address]*qlcchainContractV2{
	types.BlackHoleAddress: {
		m: map[string]ChainContractV2{
			cabi.MethodNameDestroy: &BlackHole{},
		},
		abi:     cabi.BlackHoleABI,
		abiJson: cabi.JsonDestroy,
	},
	types.MinerAddress: {
		map[string]ChainContractV2{
			cabi.MethodNameMinerReward: &MinerReward{},
		},
		cabi.MinerABI,
		cabi.JsonMiner,
	},
	types.RepAddress: {
		map[string]ChainContractV2{
			cabi.MethodNameRepReward: &RepReward{},
		},
		cabi.RepABI,
		cabi.JsonRep,
	},
}

func GetChainContract(addr types.Address, methodSelector []byte) (interface{}, bool, error) {
	if p, ok := contractCacheV1[addr]; ok {
		if method, err := p.abi.MethodById(methodSelector); err == nil {
			c, ok := p.m[method.Name]
			return c, ok, nil
		} else {
			return nil, ok, errors.New("abi: method not found")
		}
	} else if p, ok := contractCacheV2[addr]; ok {
		if method, err := p.abi.MethodById(methodSelector); err == nil {
			c, ok := p.m[method.Name]
			return c, ok, nil
		} else {
			return nil, ok, errors.New("abi: method not found")
		}
	}
	return nil, false, nil
}

func GetChainContractName(addr types.Address, methodSelector []byte) (string, bool, error) {
	if p, ok := contractCacheV1[addr]; ok {
		if method, err := p.abi.MethodById(methodSelector); err == nil {
			_, ok := p.m[method.Name]
			return method.Name, ok, nil
		} else {
			return "", ok, errors.New("abi: method not found")
		}
	} else if p, ok := contractCacheV2[addr]; ok {
		if method, err := p.abi.MethodById(methodSelector); err == nil {
			_, ok := p.m[method.Name]
			return method.Name, ok, nil
		} else {
			return "", ok, errors.New("abi: method not found")
		}
	}

	return "", false, nil
}

func IsChainContract(addr types.Address) bool {
	if _, ok := contractCacheV1[addr]; ok {
		return true
	} else if _, ok := contractCacheV2[addr]; ok {
		return true
	}
	return false
}

func GetAbiByContractAddress(addr types.Address) (string, error) {
	if contract, ok := contractCacheV1[addr]; ok {
		return contract.abiJson, nil
	} else if contract, ok := contractCacheV2[addr]; ok {
		return contract.abiJson, nil
	}
	return "", errors.New("contract not found")
}
