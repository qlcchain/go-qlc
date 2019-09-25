package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Reverse reverses a byte array.
func ReverseByte(src []byte) []byte {
	dst := make([]byte, len(src))
	for i := len(src); i > 0; i-- {
		dst[len(src)-i] = src[i-1]
	}
	return dst
}

// ReverseString reverses a hex string.
func ReverseHexString(s string) (string, error) {
	a := strings.Split(s, "")
	sRev := ""
	if len(a)%2 != 0 {
		return "", fmt.Errorf("invalid hex length")
	}
	for i := 0; i < len(a); i += 2 {
		tmp := []string{a[i], a[i+1], sRev}
		sRev = strings.Join(tmp, "")
	}
	return sRev, nil
}

// ReverseHexToUInt32 reverse a string and converts to uint32.
func ReverseHexToUInt32(s string) (uint32, error) {
	sRev, err := ReverseHexString(s)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseUint(sRev, 16, 32)
	return uint32(i), err
}

// ReverseHexToUInt64 reverse a string and converts to uint64.
func ReverseHexToUInt64(s string) (uint64, error) {
	sRev, err := ReverseHexString(s)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseUint(sRev, 16, 64)
	return uint64(i), err
}

// ReverseHexHash reverses a hex hash in string format.
func ReverseHashHex(hash string) string {
	revHash := ""
	for i := 0; i < 7; i++ {
		j := i * 8
		part := fmt.Sprintf("%c%c%c%c%c%c%c%c",
			hash[6+j], hash[7+j], hash[4+j], hash[5+j],
			hash[2+j], hash[3+j], hash[0+j], hash[1+j])
		revHash += part
	}
	return revHash
}

func UInt32ToHexLE(num uint32) string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, num)
	return hex.EncodeToString(buf)
}

func HexLEToUInt32(numHex string) uint32 {
	numBytes, err := hex.DecodeString(numHex)
	if err != nil {
		return 0
	}
	return binary.LittleEndian.Uint32(numBytes)
}

func UInt64ToHexLE(num uint64) string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint64(buf, num)
	return hex.EncodeToString(buf)
}

func HexLEToUInt64(numHex string) uint64 {
	numBytes, err := hex.DecodeString(numHex)
	if err != nil {
		return 0
	}
	return binary.LittleEndian.Uint64(numBytes)
}

func UInt32ToHexBE(num uint32) string {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, num)
	return hex.EncodeToString(buf)
}

func HexBEToUInt32(numHex string) uint32 {
	numBytes, err := hex.DecodeString(numHex)
	if err != nil {
		return 0
	}
	return binary.BigEndian.Uint32(numBytes)
}

func UInt64ToHexBE(num uint64) string {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint64(buf, num)
	return hex.EncodeToString(buf)
}

func HexBEToUInt64(numHex string) uint64 {
	numBytes, err := hex.DecodeString(numHex)
	if err != nil {
		return 0
	}
	return binary.BigEndian.Uint64(numBytes)
}

// DiffToTarget converts a whole number difficulty into a target.
func DiffToTarget(diff float64, powLimit *big.Int) (*big.Int, error) {
	if diff <= 0 {
		return nil, fmt.Errorf("invalid pool difficulty %v (0 or less than "+
			"zero passed)", diff)
	}

	// Round down in the case of a non-integer diff since we only support
	// ints (unless diff < 1 since we don't allow 0)..
	if diff < 1 {
		diff = 1
	} else {
		diff = math.Floor(diff)
	}
	divisor := new(big.Int).SetInt64(int64(diff))
	max := powLimit
	target := new(big.Int)
	target.Div(max, divisor)

	return target, nil
}

var BTCPowLimitInt, _ = new(big.Int).SetString("00000000ffff0000000000000000000000000000000000000000000000000000", 16)

// TargetToDiff converts a target into a whole number difficulty.
func TargetToDiff(target *big.Int, powLimit *big.Int) float64 {
	if powLimit == nil {
		powLimit = BTCPowLimitInt
	}

	powLimitFlt := new(big.Float).SetInt(powLimit)
	targetFlt := new(big.Float).SetInt(target)

	diffFlt := new(big.Float).Quo(powLimitFlt, targetFlt)
	diff, _ := diffFlt.Float64()
	/*
		if diff < 1 {
			diff = 1
		}
	*/

	return diff
}
