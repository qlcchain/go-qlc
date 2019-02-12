/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package resolver

import (
	"encoding/binary"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/exec"
	"github.com/qlcchain/go-qlc/vm/memory"
	"strconv"
)

func (r *Resolver) qlcTest(vm *exec.VirtualMachine) int64 {
	frame := vm.GetCurrentFrame()
	i := frame.Locals[0]
	r.logger.Debug(i)
	pointer := frame.Locals[1]

	bytes, err := vm.Memory.GetPointerMemory(uint64(pointer))
	var msg string
	if err != nil {
		r.logger.Error(err)
		ptrSize := frame.Locals[2]
		msg = util.TrimBuffToString(vm.Memory.Memory[pointer : pointer+ptrSize])
	} else {
		msg = util.TrimBuffToString(bytes)
	}

	pIdx, err := vm.Memory.SetMemory(strconv.FormatInt(i, 10) + " >>> " + msg)
	if err != nil {
		r.logger.Error(err)
		return 0
	}
	return int64(pIdx)
}

func (r *Resolver) qlcArrayLen(vm *exec.VirtualMachine) int64 {
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
		case memory.PUnknown:
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

func (r *Resolver) qlcReadInt32Param(vm *exec.VirtualMachine) int64 {
	frame := vm.GetCurrentFrame()
	addr := frame.Locals[0]
	paramBytes, err := vm.Memory.GetPointerMemory(uint64(addr))
	if err != nil {
		return 0
	}

	pIdx := vm.Memory.ParamIndex

	if pIdx+4 > len(paramBytes) {
		return 0
	}

	retInt := binary.LittleEndian.Uint32(paramBytes[pIdx : pIdx+4])
	vm.Memory.ParamIndex += 4

	return int64(retInt)
}

func (r *Resolver) qlcReadInt64Param(vm *exec.VirtualMachine) int64 {
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

func (r *Resolver) qlcReadStringParam(vm *exec.VirtualMachine) int64 {
	frame := vm.GetCurrentFrame()
	addr := frame.Locals[0]
	paramBytes, err := vm.Memory.GetPointerMemory(uint64(addr))
	if err != nil {
		r.logger.Error(err)
		return 0
	}
	var length int
	pIdx := vm.Memory.ParamIndex
	switch paramBytes[pIdx] {
	case 0xfd: //uint16
		if pIdx+3 > len(paramBytes) {
			return 0
		}
		length = int(binary.LittleEndian.Uint16(paramBytes[pIdx+1 : pIdx+3]))
		pIdx += 3
	case 0xfe: //uint32
		if pIdx+5 > len(paramBytes) {
			return 0
		}
		length = int(binary.LittleEndian.Uint16(paramBytes[pIdx+1 : pIdx+5]))
		pIdx += 5
	case 0xff:
		if pIdx+9 > len(paramBytes) {
			return 0
		}
		length = int(binary.LittleEndian.Uint16(paramBytes[pIdx+1 : pIdx+9]))
		pIdx += 9
	default:
		length = int(paramBytes[pIdx])
	}

	if pIdx+length > len(paramBytes) {
		return 0
	}
	pIdx += length + 1

	bytes := paramBytes[vm.Memory.ParamIndex+1 : pIdx]

	retIdx, err := vm.Memory.SetPointerMemory(bytes)
	if err != nil {
		return 0
	}

	vm.Memory.ParamIndex = pIdx

	return int64(retIdx)
}

func (r *Resolver) qlcHash(vm *exec.VirtualMachine) int64 {
	return 0
}
