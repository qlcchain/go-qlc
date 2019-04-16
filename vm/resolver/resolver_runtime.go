/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package resolver

import (
	"encoding/binary"
	"math"

	"github.com/qlcchain/go-qlc/vm/exec"
)

// see https://github.com/golang/go/blob/master/src/syscall/js/js.go
// for more info
const (
	nanHead     = 0x7FF80000
	jsUndefined = 0
	jsNan       = nanHead<<32 | 0
	jsZero      = nanHead<<32 | 1
	jsNull      = nanHead<<32 | 2
	jsTrue      = nanHead<<32 | 3
	jsFalse     = nanHead<<32 | 4
	jsGlobal    = nanHead<<32 | 5
	linMem      = nanHead<<32 | 6
	jsGo        = nanHead<<32 | 7

	// not part of syscall/js
	cbHelper    = nanHead<<32 | 8
	jsKwasm     = nanHead<<32 | 9
	jsFs        = nanHead<<32 | 10
	jsProcess   = nanHead<<32 | 11
	jsStub      = nanHead<<32 | 12
	jsMemBuf    = nanHead<<32 | 13
	jsPendCbs   = nanHead<<32 | 15
	jsCurCb     = nanHead<<32 | 16
	jsCurCbArgs = nanHead<<32 | 17
	jsCurCbStr  = nanHead<<32 | 18
)

var curCB FCall

func setJsInt(vm *exec.VirtualMachine, offset int, val int64) {
	sp := int(uint32(vm.GetCurrentFrame().Locals[0]))
	v := math.Float64bits(float64(val))
	//TODO: verify
	binary.LittleEndian.PutUint64(vm.Memory.Memory[sp+offset:sp+offset+8], uint64(v))
}

func setInt64(vm *exec.VirtualMachine, offset int, val int64) {
	sp := int(uint32(vm.GetCurrentFrame().Locals[0]))
	binary.LittleEndian.PutUint64(vm.Memory.Memory[sp+offset:sp+offset+8], uint64(val))
}

func setInt8(vm *exec.VirtualMachine, offset int, val byte) {
	sp := int(uint32(vm.GetCurrentFrame().Locals[0]))
	vm.Memory.Memory[sp+offset] = byte(val)
}

func getInt64(vm *exec.VirtualMachine, offset int) int64 {
	sp := int(uint32(vm.GetCurrentFrame().Locals[0]))
	//TODO: verify
	return int64(binary.LittleEndian.Uint64(vm.Memory.Memory[sp+offset : sp+offset+8]))
}

func loadString(vm *exec.VirtualMachine, offset int) string {
	addr := getInt64(vm, offset)
	dataLen := getInt64(vm, offset+8)
	//TODO: verify
	return string(vm.Memory.Memory[addr : addr+dataLen])
}

func loadBytes(vm *exec.VirtualMachine, offset int) []byte {
	addr := getInt64(vm, offset)
	dataLen := getInt64(vm, offset+8)
	//TODO: verify
	return vm.Memory.Memory[addr : addr+dataLen]
}
