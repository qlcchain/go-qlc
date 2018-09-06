/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"encoding/hex"
	"golang.org/x/crypto/blake2b"
)

func ReverseBytes(bytes []byte) []byte {
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return bytes
}

func HexToBytes(s string) []byte {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bytes
}

func Hex32ToBytes(s string) [32]byte {
	var res [32]byte
	bytes := HexToBytes(s)
	copy(res[:], bytes)
	return res
}

func Hex64ToBytes(s string) [64]byte {
	var res [64]byte
	bytes := HexToBytes(s)
	copy(res[:], bytes)
	return res
}


func Hash256(data ...[]byte) []byte {
	d, _ := blake2b.New256(nil)
	for _, item := range data {
		d.Write(item)
	}
	return d.Sum(nil)
}

func Hash(size int, data ...[]byte) []byte {
	d, _ := blake2b.New(size, nil)
	for _, item := range data {
		d.Write(item)
	}
	return d.Sum(nil)
}
