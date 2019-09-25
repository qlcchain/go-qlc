package sph

/*
void hash_x11(unsigned char *output, unsigned char *input, size_t len);
*/
import "C"

import (
	"reflect"
	"unsafe"
)

func CgoX11Hash(input []byte) []byte {
	outBytes := make([]byte, 32)
	outPtr := (*reflect.SliceHeader)(unsafe.Pointer(&outBytes))

	inPtr := (*reflect.SliceHeader)(unsafe.Pointer(&input))

	C.hash_x11((*C.uchar)(unsafe.Pointer(outPtr.Data)), (*C.uchar)(unsafe.Pointer(inPtr.Data)), C.ulong(inPtr.Len))

	return outBytes
}
