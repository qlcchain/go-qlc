/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package resolver

import (
	"encoding/binary"
	"github.com/qlcchain/go-qlc/vm/exec"
	"github.com/qlcchain/go-qlc/vm/memory"
)

func (r *Resolver) qlc_test(vm *exec.VirtualMachine) int64 {
	return 0
}

func (r *Resolver) qlc_ArrayLen(vm *exec.VirtualMachine) int64 {
	frame := vm.GetCurrentFrame()
	pointer := uint64(frame.Locals[0])
	var result int64
	if tl, ok := vm.Memory.MemPoints[pointer]; ok {
		switch tl.Type {
		case memory.PInt8, memory.PString:
			result = int64(tl.Length / 1)
		case memory.PInt16:
			result = int64(tl.Length / 2)
		case memory.PInt32, memory.PFloat32:
			result = int64(tl.Length / 4)
		case memory.PInt64, memory.PFloat64:
			result = int64(tl.Length / 8)
		case memory.PUnkown:
			//TODO: assume it's byte
			result = int64(tl.Length / 1)
		default:
			result = int64(0)
		}
	} else {
		result = int64(0)
	}

	return result
}

func (r *Resolver) qlc_ReadInt32Param(vm *exec.VirtualMachine) int64 {
	frame := vm.GetCurrentFrame()
	addr := frame.Locals[0]
	paramBytes, err := vm.Memory.GetPointerMemory(uint64(addr))
	if err != nil {
		return 0
	}

	pidx := vm.Memory.ParamIndex

	if pidx+4 > len(paramBytes) {
		return 0
	}

	retInt := binary.LittleEndian.Uint32(paramBytes[pidx : pidx+4])
	vm.Memory.ParamIndex += 4

	return int64(retInt)
}

func (r *Resolver) qlc_ReadInt64Param(vm *exec.VirtualMachine) int64 {
	frame := vm.GetCurrentFrame()
	addr := frame.Locals[0]
	paramBytes, err := vm.Memory.GetPointerMemory(uint64(addr))
	if err != nil {
		return 0
	}

	pidx := vm.Memory.ParamIndex

	if pidx+8 > len(paramBytes) {
		return 0
	}

	retInt := binary.LittleEndian.Uint64(paramBytes[pidx : pidx+8])
	vm.Memory.ParamIndex += 8

	return int64(retInt)
}

func (r *Resolver) qlc_ReadStringParam(vm *exec.VirtualMachine) int64 {
	frame := vm.GetCurrentFrame()
	addr := frame.Locals[0]
	paramBytes, err := vm.Memory.GetPointerMemory(uint64(addr))
	if err != nil {
		r.logger.Error(err)
		return 0
	}
	var length int
	pidx := vm.Memory.ParamIndex
	switch paramBytes[pidx] {
	case 0xfd: //uint16
		if pidx+3 > len(paramBytes) {
			return 0
		}
		length = int(binary.LittleEndian.Uint16(paramBytes[pidx+1 : pidx+3]))
		pidx += 3
	case 0xfe: //uint32
		if pidx+5 > len(paramBytes) {
			return 0
		}
		length = int(binary.LittleEndian.Uint16(paramBytes[pidx+1 : pidx+5]))
		pidx += 5
	case 0xff:
		if pidx+9 > len(paramBytes) {
			return 0
		}
		length = int(binary.LittleEndian.Uint16(paramBytes[pidx+1 : pidx+9]))
		pidx += 9
	default:
		length = int(paramBytes[pidx])
	}

	if pidx+length > len(paramBytes) {
		return 0
	}
	pidx += length + 1

	bytes := paramBytes[vm.Memory.ParamIndex+1 : pidx]

	retIdx, err := vm.Memory.SetPointerMemory(bytes)
	if err != nil {
		return 0
	}

	vm.Memory.ParamIndex = pidx

	return int64(retIdx)
}

func (r *Resolver) qlc_hash(vm *exec.VirtualMachine) int64 {

}
