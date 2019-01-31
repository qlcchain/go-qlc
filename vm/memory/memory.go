// modified from https://github.com/ontio/ontology-wasm
package memory

import (
	"encoding/binary"
	"errors"
	"math"
	"reflect"
)

type PointerType int

const (
	PInt8 PointerType = iota
	PInt16
	PInt32
	PInt64
	PFloat32
	PFloat64
	PString
	PStruct
	PUnkown
)

var (
	vmNilPointer        = 0
	notInitializedError = errors.New("memory is not initialized")
	outOfMemoryError    = errors.New("memory out of bound")
	notSupportSlice     = errors.New("not supported slice type")
	notSupport          = errors.New("not supported type")
)

type TypeLength struct {
	Type   PointerType
	Length int
}

type VMemory struct {
	Memory            []byte // linear memory
	AllocatedMemIndex int    // index of allocated memory
	PointedMemIndex   int    //for the pointed objects,string,array,structs
	MemPoints         map[uint64]*TypeLength
}

func NewVMemory(size uint64) VMemory {
	return VMemory{
		Memory:            make([]byte, size),
		AllocatedMemIndex: 0,
		PointedMemIndex:   0,
		MemPoints:         make(map[uint64]*TypeLength),
	}
}

//Malloc memory for base types by size, return the address in memory
func (vm *VMemory) Malloc(size int) (int, error) {
	if vm.Memory == nil || len(vm.Memory) == 0 {
		return 0, notInitializedError
	}
	if vm.AllocatedMemIndex+size+1 > len(vm.Memory) {
		return 0, outOfMemoryError
	}

	if vm.AllocatedMemIndex+size+1 > vm.PointedMemIndex {
		return 0, outOfMemoryError
	}

	offset := vm.AllocatedMemIndex + 1
	vm.AllocatedMemIndex += size

	return offset, nil
}

// MemSet
func (vm *VMemory) MemSet(dest int, ch int, count int) (int, error) {
	if vm.Memory == nil || len(vm.Memory) == 0 {
		return 0, notInitializedError
	}

	if dest > vm.Length() {
		return 0, outOfMemoryError
	}

	end := dest + count
	if end > vm.Length() {
		end = vm.Length()
	}

	for i := dest; i <= end; i++ {
		vm.Memory[i] = byte(ch)
	}

	return count, nil
}

func (vm *VMemory) Length() int {
	return len(vm.Memory)
}

//MallocPointer malloc memory for pointer types, return the address in memory
func (vm *VMemory) MallocPointer(size int, pType PointerType) (int, error) {
	if vm.Memory == nil || len(vm.Memory) == 0 {
		return 0, notInitializedError
	}
	if vm.PointedMemIndex+size > len(vm.Memory) {
		return 0, outOfMemoryError
	}

	offset := vm.PointedMemIndex + 1
	vm.PointedMemIndex += size
	//save the point and length
	vm.MemPoints[uint64(offset)] = &TypeLength{Type: pType, Length: size}
	return offset, nil
}

func (vm *VMemory) copyMemAndGetIdx(b []byte, pType PointerType) (int, error) {
	idx, err := vm.MallocPointer(len(b), pType)
	if err != nil {
		return 0, err
	}
	copy(vm.Memory[idx:idx+len(b)], b)

	return idx, nil
}

// Grow up memory size
func (vm *VMemory) Grow(size uint64) {
	tmp := []byte(vm.Memory)
	vm.Memory = append(tmp, make([]byte, size)...)
}

//GetPointerMemSize return pointed memory size
func (vm *VMemory) GetPointerMemSize(addr uint64) int {
	//nil case
	if addr == uint64(vmNilPointer) {
		return 0
	}

	v, ok := vm.MemPoints[addr]
	if ok {
		return v.Length
	} else {
		return 0
	}
}

//GetPointerMemory return pointed memory data of the address
//when wasm returns a pointer, call this function to get the pointed memory
func (vm *VMemory) GetPointerMemory(addr uint64) ([]byte, error) {
	//nil case
	if addr == uint64(vmNilPointer) {
		return nil, nil
	}

	length := vm.GetPointerMemSize(addr)
	if length == 0 {
		return nil, nil
	}

	if int(addr)+length > len(vm.Memory) {
		return nil, outOfMemoryError
	} else {
		return vm.Memory[int(addr) : int(addr)+length], nil
	}
}

//SetPointerMemory set pointer types into memory, return address of memory
func (vm *VMemory) SetPointerMemory(val interface{}) (int, error) {
	//nil case
	if val == nil {
		return vmNilPointer, nil
	}

	switch reflect.TypeOf(val).Kind() {
	case reflect.String:
		b := []byte(val.(string))
		return vm.copyMemAndGetIdx(b, PString)
	case reflect.Array, reflect.Struct, reflect.Ptr:
		//TODO  implement
		return 0, nil
	case reflect.Slice:
		switch val.(type) {
		case []byte:
			return vm.copyMemAndGetIdx(val.([]byte), PString)

		case []int:
			intBytes := make([]byte, len(val.([]int))*4)
			for i, v := range val.([]int) {
				tmp := make([]byte, 4)
				binary.LittleEndian.PutUint32(tmp, uint32(v))
				copy(intBytes[i*4:(i+1)*4], tmp)
			}
			return vm.copyMemAndGetIdx(intBytes, PInt32)
		case []int64:
			intBytes := make([]byte, len(val.([]int64))*8)
			for i, v := range val.([]int64) {
				tmp := make([]byte, 8)
				binary.LittleEndian.PutUint64(tmp, uint64(v))
				copy(intBytes[i*8:(i+1)*8], tmp)
			}
			return vm.copyMemAndGetIdx(intBytes, PInt64)

		case []float32:
			floatBytes := make([]byte, len(val.([]float32))*4)
			for i, v := range val.([]float32) {
				tmp := float32ToByte(v)
				copy(floatBytes[i*4:(i+1)*4], tmp)
			}
			return vm.copyMemAndGetIdx(floatBytes, PFloat32)

		case []float64:
			floatBytes := make([]byte, len(val.([]float64))*4)
			for i, v := range val.([]float64) {
				tmp := float64ToByte(v)
				copy(floatBytes[i*8:(i+1)*8], tmp)
			}
			return vm.copyMemAndGetIdx(floatBytes, PFloat64)

		default:
			return 0, notSupportSlice
		}

	default:
		return 0, notSupport
	}

}

//SetStructMemory set struct into memory , return address of memory
func (vm *VMemory) SetStructMemory(val interface{}) (int, error) {
	if reflect.TypeOf(val).Kind() != reflect.Struct {
		return 0, errors.New("SetStructMemory :input is not a struct")
	}
	valref := reflect.ValueOf(val)
	//var totalsize = 0
	var index = 0
	for i := 0; i < valref.NumField(); i++ {
		field := valref.Field(i)

		//nested struct case
		if reflect.TypeOf(field.Type()).Kind() == reflect.Struct {
			idx, err := vm.SetStructMemory(field)
			if err != nil {
				return 0, err
			} else {
				if i == 0 && index == 0 {
					index = idx
				}
			}
		} else {
			var fieldVal interface{}
			//TODO: how to determine the value is int or int64
			var idx int
			var err error
			switch field.Kind() {
			case reflect.Int, reflect.Int32, reflect.Uint, reflect.Uint32:
				fieldVal = int(field.Int())
				idx, err = vm.SetMemory(fieldVal)
			case reflect.Int64, reflect.Uint64:
				fieldVal = field.Int()
				idx, err = vm.SetMemory(fieldVal)
			case reflect.Float32, reflect.Float64:
				fieldVal = field.Float()
				idx, err = vm.SetMemory(fieldVal)
			case reflect.String:
				fieldVal = field.String()
				tmp, err := vm.SetPointerMemory(fieldVal)
				if err != nil {
					return 0, err
				}
				//add the point address to memory
				idx, err = vm.SetMemory(tmp)

			case reflect.Slice:
				//fieldVal = field.Interface()
				//TODO note the struct field MUST be public
				tmp, err := vm.SetPointerMemory(fieldVal)
				if err != nil {
					return 0, err
				}
				//add the point address to memory
				idx, err = vm.SetMemory(tmp)
			}

			if err != nil {
				return 0, err
			} else {
				if i == 0 && index == 0 {
					index = idx
				}
			}
		}
	}
	return index, nil

}

//SetMemory set base types into memory, return address of memory
func (vm *VMemory) SetMemory(val interface{}) (int, error) {
	switch val.(type) {
	case string: //use SetPointerMemory for string
		return vm.SetPointerMemory(val.(string))
	case int:
		tmp := make([]byte, 4)
		binary.LittleEndian.PutUint32(tmp, uint32(val.(int)))
		idx, err := vm.Malloc(len(tmp))
		if err != nil {
			return 0, err
		}
		copy(vm.Memory[idx:idx+len(tmp)], tmp)
		return idx, nil
	case int64:
		tmp := make([]byte, 8)
		binary.LittleEndian.PutUint64(tmp, uint64(val.(int64)))
		idx, err := vm.Malloc(len(tmp))
		if err != nil {
			return 0, err
		}
		copy(vm.Memory[idx:idx+len(tmp)], tmp)
		return idx, nil
	case float32:
		tmp := float32ToByte(val.(float32))

		idx, err := vm.Malloc(len(tmp))
		if err != nil {
			return 0, err
		}
		copy(vm.Memory[idx:idx+len(tmp)], tmp)
		return idx, nil
	case float64:
		tmp := float64ToByte(val.(float64))
		idx, err := vm.Malloc(len(tmp))
		if err != nil {
			return 0, err
		}
		copy(vm.Memory[idx:idx+len(tmp)], tmp)
		return idx, nil

	default:
		return 0, notSupport
	}
}

func float32ToByte(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)

	return bytes
}

func float64ToByte(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)

	return bytes
}
