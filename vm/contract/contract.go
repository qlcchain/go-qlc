/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"github.com/pkg/errors"
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

type ChainContract interface {
	GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error)
	// DoSend verify or update StateBlock.Data
	DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error
	// check status, update state
	DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error)
	// refund data at receive error
	GetRefundData() []byte
}

type qlcchainContract struct {
	m   map[string]ChainContract
	abi abi.ABIContract
}

var contractCache = map[types.Address]*qlcchainContract{
	types.MintageAddress: {
		map[string]ChainContract{
			cabi.MethodNameMintage:         &Mintage{},
			cabi.MethodNameMintageWithdraw: &WithdrawMintage{},
		},
		cabi.MintageABI,
	},
	types.NEP5PledgeAddress: {
		map[string]ChainContract{
			cabi.MethodNEP5Pledge:         &Nep5Pledge{},
			cabi.MethodWithdrawNEP5Pledge: &WithdrawNep5Pledge{},
		},
		cabi.NEP5PledgeABI,
	},
	types.MinerAddress: {
		map[string]ChainContract{
			cabi.MethodNameMinerReward: &MinerReward{},
		},
		cabi.MinerABI,
	},
}

func GetChainContract(addr types.Address, methodSelector []byte) (ChainContract, bool, error) {
	p, ok := contractCache[addr]
	if ok {
		if method, err := p.abi.MethodById(methodSelector); err == nil {
			c, ok := p.m[method.Name]
			return c, ok, nil
		} else {
			return nil, ok, errors.New("abi: method not found")
		}
	}
	return nil, ok, nil
}

func IsChainContract(addr types.Address) bool {
	if _, ok := contractCache[addr]; ok {
		return true
	}
	return false
}
