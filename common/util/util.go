/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"math/rand"
	"os"
	"time"

	"golang.org/x/crypto/blake2b"
)

func ReverseBytes(str []byte) (result []byte) {
	for i := len(str) - 1; i >= 0; i-- {
		result = append(result, str[i])
	}
	return result
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

// TrimQuotes trim quotes of string if quotes exist
func TrimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

//CreateDirIfNotExist create given folder
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0700)
		return err
	}
	return nil
}

func ToString(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func ToIndentString(v interface{}) string {
	b, err := json.MarshalIndent(&v, "", "\t")
	if err != nil {
		return ""
	}
	return string(b)
}

//trim the '\00' byte
func TrimBuffToString(bytes []byte) string {
	for i, b := range bytes {
		if b == 0 {
			return string(bytes[:i])
		}
	}
	return string(bytes)

}

const (
	MaxUint64 = uint64(1<<64 - 1)

	// number of bits in a big.Word
	WordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// number of bytes in a big.Word
	WordBytes = WordBits / 8
	// number of bytes in a vm word
	WordSize = 32
)

var (
	Big0   = big.NewInt(0)
	Big1   = big.NewInt(1)
	Big2   = big.NewInt(2)
	Big10  = big.NewInt(10)
	Big31  = big.NewInt(31)
	Big32  = big.NewInt(32)
	Big256 = big.NewInt(256)
	Big257 = big.NewInt(257)

	Tt255   = BigPow(2, 255)
	Tt256   = BigPow(2, 256)
	Tt256m1 = new(big.Int).Sub(Tt256, big.NewInt(1))

	chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

// ToWordSize returns the ceiled word size required for memory expansion.
func ToWordSize(size uint64) uint64 {
	if size > MaxUint64-WordSize+1 {
		return MaxUint64/WordSize + 1
	}

	return (size + WordSize - 1) / WordSize
}

// BigUint64 returns the integer casted to a uint64 and returns whether it
// overflowed in the process.
func BigUint64(v *big.Int) (uint64, bool) {
	return v.Uint64(), v.BitLen() > 64
}

func BytesToString(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}

func AllZero(b []byte) bool {
	for _, byte := range b {
		if byte != 0 {
			return false
		}
	}
	return true
}

func JoinBytes(data ...[]byte) []byte {
	newData := []byte{}
	for _, d := range data {
		newData = append(newData, d...)
	}
	return newData
}

func RandomFixedString(length int) string {
	if length == 0 {
		return ""
	}
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return string(b)
}
