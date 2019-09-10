/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/exec"
	"github.com/qlcchain/go-qlc/vm/platform"
	"github.com/qlcchain/go-qlc/vm/resolver"
	"go.uber.org/zap"
)

const (
	ContractMethodName = "invoke"
	ContractInitMethod = "init"
)

type SignFunc func(addr types.Address, data []byte) (types.Signature, error)

type GenResult struct {
	Blocks  []*types.StateBlock
	IsRetry bool
	Error   error
}

type ExecuteEngine struct {
	vm     *exec.VirtualMachine
	block  *types.SmartContractBlock
	schema abi.ABIContract
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

	aotSvc := platform.FullAOTCompile(vm)
	if aotSvc != nil {
		vm.SetAOTService(aotSvc)
	}

	schema, _ := abi.JSONToABIContract(strings.NewReader(contract.AbiSchema))

	return &ExecuteEngine{
		vm,
		contract,
		schema,
		caller,
		ledger,
		log.NewLogger("engine"),
	}
}

// InitCall init method on deployment
func (e *ExecuteEngine) InitCall(args []byte) (result []byte, err error) {
	return e.call(ContractInitMethod, args)
}

// Call invoke to run the contract code by block.data
func (e *ExecuteEngine) Call(data []byte) (result []byte, err error) {
	method, err := e.schema.MethodById(data[:4])
	if err != nil {
		return nil, err
	}
	return e.call(method.Name, data[4:])
}

func (e *ExecuteEngine) call(action string, args []byte) (result []byte, err error) {
	defer func() {
		if err := recover(); err != nil {
			result = nil
			err = errors.New("execute engine call panic error")
		}
	}()

	entryID, ok := e.vm.GetFunctionExport(action)
	if !ok {
		return nil, fmt.Errorf("Entry function %s not found; starting from 0.\n", action)
	}

	argIdx, err := e.vm.Memory.SetPointerMemory(args)
	if err != nil {
		return nil, errors.New("failed to copy input args to vm memory")
	}

	// If any function prior to the entry function was declared to be
	// called by the module, run it first.
	if e.vm.Module.Base.Start != nil {
		startID := int(e.vm.Module.Base.Start.Index)
		_, err := e.vm.Run(startID)
		if err != nil {
			e.vm.PrintStackTrace()
			panic(err)
		}
	}

	var params []int64
	params = append(params, int64(argIdx))
	ret, err := e.vm.Run(entryID, params...)
	if err != nil {
		e.vm.PrintStackTrace()
		panic(err)
	}

	return util.LE_Int64ToBytes(ret), nil
}
