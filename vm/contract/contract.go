/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"errors"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var (
	ErrToken            = errors.New("token err")
	ErrUnpackMethod     = errors.New("unpack method err")
	ErrPackMethod       = errors.New("pack method err")
	ErrNotEnoughPledge  = errors.New("not enough pledge")
	ErrCheckParam       = errors.New("check param err")
	ErrSetStorage       = errors.New("set storage err")
	ErrCalcAmount       = errors.New("calc amount err")
	ErrNotEnoughFee     = errors.New("not enough fee")
	ErrGetVerifier      = errors.New("get verifier err")
	ErrAccountInvalid   = errors.New("invalid account")
	ErrAccountNotExist  = errors.New("account not exist")
	ErrGetNodeHeight    = errors.New("get node height err")
	ErrEndHeightInvalid = errors.New("invalid claim end height")
	ErrClaimRepeat      = errors.New("claim reward repeatedly")
	ErrGetRewardHistory = errors.New("get reward history err")
	ErrVerifierNum      = errors.New("verifier num err")
	ErrPledgeNotReady   = errors.New("pledge is not ready")
	ErrGetAdmin         = errors.New("get admin err")
	ErrInvalidAdmin     = errors.New("invalid admin")
	ErrInvalidLen       = errors.New("invalid len")
)

type BaseContract struct {
	Describe
}

func (c *BaseContract) GetDescribe() Describe {
	return c.Describe
}

func (c *BaseContract) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	return types.ZeroAddress, nil
}

func (c *BaseContract) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

// refund data at receive error
func (c *BaseContract) GetRefundData() []byte {
	return []byte{1}
}

// DoPending generate pending info from send block
func (c *BaseContract) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return nil, nil, nil
}

// ProcessSend verify or update StateBlock.Data
func (c *BaseContract) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	return errors.New("not implemented")
}

// check status, update state
func (c *BaseContract) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, errors.New("not implemented")
}

// ProcessSend verify or update StateBlock.Data
func (c *BaseContract) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return nil, nil, errors.New("not implemented")
}

func (c *BaseContract) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	return common.ContractNoGap, nil, nil
}

func (c *BaseContract) DoSendOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock) error {
	return errors.New("not implemented")
}

func (c *BaseContract) DoReceiveOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock, input *types.StateBlock) error {
	return errors.New("not implemented")
}

type Contract interface {
	// Contract meta describe
	GetDescribe() Describe
	// Target receiver address
	GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error)

	GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error)
	// check status, update state
	DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error)
	// refund data at receive error
	GetRefundData() []byte

	// DoPending generate pending info from send block
	DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error)
	// ProcessSend verify or update StateBlock.Data
	DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error

	// ProcessSend verify or update StateBlock.Data
	ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error)
	DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error)

	DoSendOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock) error
	DoReceiveOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock, input *types.StateBlock) error
}

const (
	SpecVerInvalid = iota
	SpecVer1       = 1
	SpecVer2       = 2
)

type Describe struct {
	specVer   int
	signature bool
	pending   bool
	povState  bool
	work      bool
}

func (d Describe) GetVersion() int {
	return d.specVer
}
func (d Describe) WithSignature() bool {
	return d.signature
}
func (d Describe) WithPending() bool {
	return d.pending
}
func (d Describe) WithPovState() bool {
	return d.povState
}
func (d Describe) WithWork() bool {
	return d.work
}

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

type qlcChainContract struct {
	m       map[string]Contract
	abi     abi.ABIContract
	abiJson string
}

func NewChainContract(m map[string]Contract, abi abi.ABIContract, abiJson string) *qlcChainContract {
	return &qlcChainContract{
		m:       m,
		abi:     abi,
		abiJson: abiJson,
	}
}

var qlcAllChainContracts = make(map[types.Address]*qlcChainContract)

func RegisterContracts(addr types.Address, contract *qlcChainContract) {
	if _, ok := qlcAllChainContracts[addr]; !ok {
		qlcAllChainContracts[addr] = contract
	}
}

func GetChainContract(addr types.Address, methodSelector []byte) (Contract, bool, error) {
	if p, ok := qlcAllChainContracts[addr]; ok {
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
	if p, ok := qlcAllChainContracts[addr]; ok {
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
	if _, ok := qlcAllChainContracts[addr]; ok {
		return true
	}
	return false
}

func GetAbiByContractAddress(addr types.Address) (string, error) {
	if contract, ok := qlcAllChainContracts[addr]; ok {
		return contract.abiJson, nil
	}
	return "", errors.New("contract not found")
}

func init() {
	RegisterContracts(contractaddress.MintageAddress, MintageContract)
	RegisterContracts(contractaddress.NEP5PledgeAddress, Nep5Contract)
	RegisterContracts(contractaddress.RewardsAddress, RewardsContract)
	RegisterContracts(contractaddress.BlackHoleAddress, BlackHoleContract)
	RegisterContracts(contractaddress.MinerAddress, MinerContract)
	RegisterContracts(contractaddress.RepAddress, RepContract)
	RegisterContracts(contractaddress.SettlementAddress, SettlementContract)
	RegisterContracts(contractaddress.PubKeyDistributionAddress, PKDContract)
	RegisterContracts(contractaddress.PermissionAddress, PermissionContract)
	RegisterContracts(contractaddress.PrivacyDemoKVAddress, PdkvContract)
	RegisterContracts(contractaddress.PtmKeyKVAddress, PtmkeyContract)
}
