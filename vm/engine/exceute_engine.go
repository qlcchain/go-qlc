/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package engine

import (
	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/exec"
	"github.com/qlcchain/go-qlc/vm/resolver"
	"go.uber.org/zap"
)

const (
	ContractMethodName = "invoke"
	ContractInitMethod = "init"
)

type GenResult struct {
	Blocks  []*types.StateBlock
	IsRetry bool
	Error   error
}

type ExecuteEngine struct {
	vm     *exec.VirtualMachine
	block  *types.SmartContractBlock
	caller types.Address
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

// NewExecuteEngine new vm execute engine
func NewExecuteEngine(caller types.Address, contract *types.SmartContractBlock, ledger *ledger.Ledger) *ExecuteEngine {
	vm, _ := exec.NewVirtualMachine(contract.Abi.Abi, exec.VMConfig{
		DefaultMemoryPages:   128,
		DefaultTableSize:     65536,
		DisableFloatingPoint: false,
	}, resolver.NewResolver(), nil)

	return &ExecuteEngine{
		vm,
		contract,
		caller,
		ledger,
		log.NewLogger("engine"),
	}
}

// Call invoke to run the contract code
func (e *ExecuteEngine) Call(args []interface{}) ([]byte, error) {
	// call vm.Run()
	return nil, nil
}

func (e *ExecuteEngine) GenerateBlock(block *types.StateBlock) (*GenResult, error) {
	var sendBlock *types.StateBlock = nil
	var err error
	if block.IsReceiveBlock() {
		if sendBlock, err = e.ledger.GetStateBlock(block.Link); err != nil {
			return nil, errors.New("can not fetch block link")
		}
	}
	result, err := e.generate(block, sendBlock)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (e *ExecuteEngine) generate(block *types.StateBlock, input *types.StateBlock) (*GenResult, error) {
	return &GenResult{}, nil
}
