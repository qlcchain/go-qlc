/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import "github.com/qlcchain/go-qlc/vm/exec"

type ContractApi interface {
	calloc(vm *exec.VirtualMachine) (bool, error)
	malloc(vm *exec.VirtualMachine) (bool, error)
	arrayLen(vm *exec.VirtualMachine) (bool, error)
	memcpy(vm *exec.VirtualMachine) (bool, error)
	memset(vm *exec.VirtualMachine) (bool, error)
	//utility apis
	stringcmp(vm *exec.VirtualMachine) (bool, error)
	stringconcat(vm *exec.VirtualMachine) (bool, error)
	Atoi(vm *exec.VirtualMachine) (bool, error)
	Atoi64(vm *exec.VirtualMachine) (bool, error)
	Itoa(vm *exec.VirtualMachine) (bool, error)
	i64add(vm *exec.VirtualMachine) (bool, error)
	QLC_ReadInt32Param(vm *exec.VirtualMachine) (bool, error)
	QLC_ReadInt64Param(vm *exec.VirtualMachine) (bool, error)
	QLC_ReadStringParam(vm *exec.VirtualMachine) (bool, error)
	QLC_GetCallerAddress(vm *exec.VirtualMachine) (bool, error)
	QLC_GetSelfAddress(vm *exec.VirtualMachine) (bool, error)
}
