package resolver

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/exec"
	"github.com/qlcchain/go-qlc/vm/memory"
	"go.uber.org/zap"
	"io"
	"strconv"
	"time"
)

// Resolver defines imports for WebAssembly modules ran in Life.
type Resolver struct {
	Stderr       bytes.Buffer
	BlockedCalls []FCall
	PendingCalls []FCall
	NewCalls     int
	logger       *zap.SugaredLogger
}

func NewResolver() *Resolver {
	return &Resolver{logger: log.NewLogger("Resolver")}
}

// ResolveFunc defines a set of import functions that may be called within a WebAssembly module.
func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	fmt.Printf("Resolve func: %s %s\n", module, field)
	switch module {
	case "env":
		switch field {
		case "__life_ping":
			return func(vm *exec.VirtualMachine) int64 {
				return vm.GetCurrentFrame().Locals[0] + 1
			}
		case "__life_log":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(uint32(frame.Locals[0]))
				msgLen := int(uint32(frame.Locals[1]))
				//TODO: verify
				msg := vm.Memory.Memory[ptr : ptr+msgLen]
				fmt.Printf("[app] %s\n", string(msg))
				return 0
			}
		case "print_i64":
			return func(vm *exec.VirtualMachine) int64 {
				fmt.Printf("[app] print_i64: %d\n", vm.GetCurrentFrame().Locals[0])
				return 0
			}
		case "memcpy":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				dest := int(frame.Locals[0])
				src := int(frame.Locals[1])
				length := int(frame.Locals[2])

				//memcpy overlapped
				if dest < src && dest+length > src {
					return 0
				}
				copy(vm.Memory.Memory[dest:dest+length], vm.Memory.Memory[src:src+length])
				return int64(dest)
			}
		case "calloc":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				count := int(frame.Locals[0])
				length := int(frame.Locals[1])

				//we don't know whats the alloc type here
				index, err := vm.Memory.MallocPointer(count*length, memory.PUnknown)
				if err != nil {
					return 0
				}
				return int64(index)
			}
		case "malloc":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				count := int(frame.Locals[0])
				index, err := vm.Memory.MallocPointer(count, memory.PUnknown)

				if err != nil {
					return 0
				}
				return int64(index)
			}
		case "memset":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				dest := int(frame.Locals[0])
				char := int(frame.Locals[1])
				cnt := int(frame.Locals[2])

				tmp := make([]byte, cnt)
				for i := 0; i < cnt; i++ {
					tmp[i] = byte(char)
				}
				copy(vm.Memory.Memory[dest:dest+cnt], tmp)

				return int64(dest)
			}
		case "strcmp":
			return func(vm *exec.VirtualMachine) int64 {
				var ret int64
				frame := vm.GetCurrentFrame()
				addr1 := uint64(frame.Locals[0])
				addr2 := uint64(frame.Locals[1])

				if addr1 == addr2 {
					ret = 0
				} else {
					bytes1, err := vm.Memory.GetPointerMemory(addr1)
					if err != nil {
						return -1
					}

					bytes2, err := vm.Memory.GetPointerMemory(addr2)
					if err != nil {
						return -1
					}

					if util.TrimBuffToString(bytes1) == util.TrimBuffToString(bytes2) {
						ret = 0
					} else {
						ret = 1
					}
				}
				return ret
			}
		case "strcat":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()

				str1, err := vm.Memory.GetPointerMemory(uint64(frame.Locals[0]))
				if err != nil {
					return 0
				}

				str2, err := vm.Memory.GetPointerMemory(uint64(frame.Locals[1]))
				if err != nil {
					return 0
				}

				newString := util.TrimBuffToString(str1) + util.TrimBuffToString(str2)

				idx, err := vm.Memory.SetPointerMemory(newString)
				if err != nil {
					return 0
				}

				return int64(idx)
			}
		case "atoi":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				if len(frame.Locals) != 1 {
					return 0
				}

				addr := frame.Locals[0]

				pBytes, err := vm.Memory.GetPointerMemory(uint64(addr))
				if err != nil {
					return 0
				}
				if pBytes == nil || len(pBytes) == 0 {
					return 0
				}

				str := util.TrimBuffToString(pBytes)
				i, err := strconv.Atoi(str)
				if err != nil {
					return 0
				}

				return int64(i)
			}
		case "atoi64":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				if len(frame.Locals) != 1 {
					return 0
				}

				addr := frame.Locals[0]

				pBytes, err := vm.Memory.GetPointerMemory(uint64(addr))
				if err != nil {
					return 0
				}
				if pBytes == nil || len(pBytes) == 0 {
					return 0
				}

				str := util.TrimBuffToString(pBytes)
				i, err := strconv.ParseInt(str, 10, 64)
				if err != nil {
					return 0
				}

				return int64(i)
			}
		case "itoa":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				i := int(frame.Locals[0])

				str := strconv.Itoa(i)

				idx, err := vm.Memory.SetPointerMemory(str)
				if err != nil {
					return 0
				}
				return int64(idx)
			}
		case "i64toa":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				i := frame.Locals[0]
				radix := int(frame.Locals[1])

				str := strconv.FormatInt(i, radix)
				idx, err := vm.Memory.SetPointerMemory(str)
				if err != nil {
					return 0
				}

				return int64(idx)
			}
		case "QLC_ArrayLen":
			return r.qlcArrayLen
		case "QLC_ReadInt32Param":
			return r.qlcReadInt32Param
		case "QLC_ReadInt64Param":
			return r.qlcReadInt64Param
		case "QLC_ReadStringParam":
			return r.qlcReadStringParam
		case "QLC_UnmarshalInputs":
			return r.qlcUnMarshalInputs
		case "QLC_MarshalResult":
			return r.qlcMarshalResult
		case "QLC_GetCallerAddress":
			return r.qlcGetCaller
		case "QLC_GetSelfAddress":
			return r.qlcGetSelfAddress
		case "QLC_Test":
			return r.qlcTest
		case "QLC_Hash":
			return r.qlcHash
		default:
			panic(fmt.Errorf("unknown field: %s", field))
		}
	case "go":
		switch field {
		case "runtime.wasmWrite":
			return r.Write
		case "runtime.nanotime":
			return r.NanoTime
		case "runtime.walltime":
			return r.WallTime
		case "syscall/js.valueGet":
			return r.ValueGet
		case "syscall/js.valueCall":
			return r.ValueCall
		case "syscall/js.valueInvoke":
			return r.ValueInvoke
		case "syscall/js.valueNew":
			return r.ValueNew
		case "syscall/js.valueLength":
			return r.ValueLength
		case "syscall/js.valueIndex":
			return r.ValueIndex
		case "syscall/js.valuePrepareString":
			return r.ValuePrepString
		case "syscall/js.valueLoadString":
			return r.ValueLoadString
		case "runtime.getRandomData":
			return r.RandomData
		case "runtime.wasmExit":
			return r.WasmExit
		default:
			return r.Stub(field)
		}

	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

// ResolveGlobal defines a set of global variables for use within a WebAssembly module.
func (r *Resolver) ResolveGlobal(module, field string) int64 {
	fmt.Printf("Resolve global: %s %s\n", module, field)
	switch module {
	case "env":
		switch field {
		case "__life_magic":
			return 424
		default:
			panic(fmt.Errorf("unknown field: %s", field))
		}
	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

func (r *Resolver) Stub(name string) func(vm *exec.VirtualMachine) int64 {
	return func(vm *exec.VirtualMachine) int64 {
		panic(name)
		return 0
	}
}

func (r *Resolver) RandomData(vm *exec.VirtualMachine) int64 {
	return 0
}

func (r *Resolver) WasmExit(vm *exec.VirtualMachine) int64 {
	return 0
}

func (r *Resolver) ValueLength(vm *exec.VirtualMachine) int64 {
	setInt64(vm, 16, 1)
	return 0
}
func (r *Resolver) ValueIndex(vm *exec.VirtualMachine) int64 {
	setInt64(vm, 16, 1)
	return 0
}
func (r *Resolver) ValuePrepString(vm *exec.VirtualMachine) int64 {
	setInt64(vm, 16, jsCurCbStr)
	setInt64(vm, 24, int64(len(curCB.Output)))
	return 0
}
func (r *Resolver) ValueLoadString(vm *exec.VirtualMachine) int64 {
	arr := getInt64(vm, 16)
	dataLen := getInt64(vm, 24)
	//TODO: verify
	copy(vm.Memory.Memory[arr:arr+dataLen], curCB.Output)
	return 0
}

func (r *Resolver) WallTime(vm *exec.VirtualMachine) int64 {
	setInt64(vm, 8, time.Now().Unix())
	setInt64(vm, 16, time.Now().UnixNano()/1000000000)
	return 0
}

func (r *Resolver) NanoTime(vm *exec.VirtualMachine) int64 {
	setInt64(vm, 8, time.Now().UnixNano())
	return 0
}

func (r *Resolver) Write(vm *exec.VirtualMachine) int64 {
	sp := int(uint32(vm.GetCurrentFrame().Locals[0]))
	//TODO: verify
	fd := binary.LittleEndian.Uint64(vm.Memory.Memory[sp+8 : sp+16])
	ptr := binary.LittleEndian.Uint64(vm.Memory.Memory[sp+16 : sp+24])
	msgLen := binary.LittleEndian.Uint64(vm.Memory.Memory[sp+24 : sp+32])
	msg := vm.Memory.Memory[ptr : ptr+msgLen]
	var out io.Writer
	switch fd {
	case 2:
		out = &r.Stderr
	default:
		panic("only stderr file descriptor is supported")
	}
	n, err := out.Write(msg)
	if err != nil {
		panic(err)
	}
	return int64(n)
}

// function parameters as JSON in method name to
// avoid unnecessary conversion logic
type FCall struct {
	Method string
	CB     int
	Input  json.RawMessage
	Output json.RawMessage
}

// this is invoked only for _makeCallbackHelper, which we don't need
// so we just ignore the stuff
func (r *Resolver) ValueInvoke(vm *exec.VirtualMachine) int64 {
	setInt8(vm, 48, 1)
	return 0
}

func (r *Resolver) ValueCall(vm *exec.VirtualMachine) int64 {
	ptr := getInt64(vm, 8)
	b := loadBytes(vm, 16)
	if ptr == jsPendCbs && string(b) == "shift" {
		if len(r.PendingCalls) == 0 {
			setInt64(vm, 56, jsZero)
			setInt8(vm, 64, 1)
			return 0
		}
		curCB, r.PendingCalls = r.PendingCalls[0], r.PendingCalls[1:]
		setInt64(vm, 56, jsCurCb)
		setInt8(vm, 64, 1)
		return 0
	}
	if ptr != jsKwasm {
		panic(fmt.Sprintf("Call implemented only for kwasm object: %v %v", ptr, string(b)))
	}
	var c FCall
	err := json.Unmarshal(b, &c)
	if err != nil {
		panic(err)
	}
	switch c.Method {
	case "kwasm.log":
		var inp string
		err := json.Unmarshal(c.Input, &inp)
		if err != nil {
			panic(err)
		}
		_, err = r.Stderr.WriteString(inp)
		if err != nil {
			panic(err)
		}
	default:
		r.BlockedCalls = append(r.BlockedCalls, c)
	}
	setInt8(vm, 64, 1)
	return 0
}

func (r *Resolver) ValueNew(vm *exec.VirtualMachine) int64 {
	if r.NewCalls > 0 {
		panic("only single call to New() is allowed! (for pending Callbacks)")
	}
	r.NewCalls++ // prevent any use of syscall/js API, i.e. fail fast
	setInt64(vm, 40, jsPendCbs)
	setInt64(vm, 48, 1)
	return 0
}

// because VM should not have business with host's memory
// we have a set of hardcoded pointers & values that always have
// consistent value across all VM hosts.
// We don't wanna manage garbage collection and other shit
// the whole point is to make minimal functionality to run GO inside life VM
func (r *Resolver) ValueGet(vm *exec.VirtualMachine) int64 {
	ptr := getInt64(vm, 8)
	str := loadString(vm, 16)
	prefix := "?."
	switch ptr {
	case jsGlobal:
		prefix = "jsGlobal."
	case linMem:
		prefix = "linMem."
	case jsMemBuf:
		prefix = "jsMemBuf."
	case jsStub:
		prefix = "jsStub."
	case jsZero:
		prefix = "jsZero."
	case jsGo:
		prefix = "jsGo."
	case jsKwasm:
		prefix = "jsKwasm."
	case jsFs:
		prefix = "jsFs."
	case jsProcess:
		prefix = "jsProcess."
	case jsCurCb:
		prefix = "jsCurCb."
	}
	name := prefix + str
	switch name {
	case "jsGlobal.Go":
		setInt64(vm, 32, jsGo)
		return 0
	case "jsGlobal.process":
		setInt64(vm, 32, jsProcess)
		return 0
	case "linMem.buffer":
		setInt64(vm, 32, jsMemBuf)
		return 0
	case "jsGlobal.fs":
		setInt64(vm, 32, jsFs)
		return 0
	case "jsGo._callbackShutdown":
		setInt64(vm, 32, jsFalse)
		return 0
	case "jsGlobal.kwasm":
		setInt64(vm, 32, jsKwasm)
		return 0
	case "jsGo._makeCallbackHelper":
		setInt64(vm, 32, cbHelper)
		return 0
	case "jsFs.O_WRONLY":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.O_RDWR":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.O_CREAT":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.O_TRUNC":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.O_APPEND":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.O_EXCL":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.constants":
		return 0 // stub
	case "jsCurCb.id":
		setJsInt(vm, 32, int64(curCB.CB))
		return 0
	case "jsCurCb.args":
		setJsInt(vm, 32, jsCurCbArgs)
		return 0
	}
	setInt64(vm, 32, jsStub)
	return 0
}
