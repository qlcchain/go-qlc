/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

// Package common contains various helper functions.
package util

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"math"
)

// FromHex returns the bytes represented by the hexadecimal string s.
// s may be prefixed with "0x".
func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// CopyBytes returns an exact copy of the provided bytes.
func CopyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

// RightPadBytes zero-pads slice to the right up to length l.
func RightPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded, slice)

	return padded
}

// LeftPadBytes zero-pads slice to the left up to length l.
func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

func BE_Uint32ToBytes(i uint32) []byte {
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, i)
	return tmp
}

func BE_BytesToUint32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}

func LE_Uint32ToBytes(i uint32) []byte {
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, i)
	return tmp
}

func LE_BytesToUint32(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf)
}

func BE_Uint64ToBytes(i uint64) []byte {
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint64(tmp, i)
	return tmp
}

func BE_BytesToUint64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}

func LE_Uint64ToBytes(i uint64) []byte {
	tmp := make([]byte, 8)
	binary.LittleEndian.PutUint64(tmp, i)
	return tmp
}

func LE_BytesToUint64(buf []byte) uint64 {
	return binary.LittleEndian.Uint64(buf)
}

func LE_Int64ToBytes(i int64) []byte {
	return LE_Uint64ToBytes(uint64(i))
}

func LE_BytesToInt64(buf []byte) int64 {
	return int64(LE_BytesToUint64(buf))
}

func String2Bytes(s string) []byte {
	return []byte(s)
}

func BE_Int2Bytes(i int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}

func BE_Bytes2Int(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func Bool2Bytes(b bool) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(b)
	return buf.Bytes()
}

func LE_EncodeVarInt(val uint64) []byte {
	buf := make([]byte, 9)

	if val < 0xfd {
		buf[0] = uint8(val)
		return buf[0:1]
	}

	if val <= math.MaxUint16 {
		buf[0] = uint8(0xfd)
		binary.LittleEndian.PutUint16(buf[1:3], uint16(val))
		return buf[0:3]
	}

	if val <= math.MaxUint32 {
		buf[0] = uint8(0xfe)
		binary.LittleEndian.PutUint32(buf[1:5], uint32(val))
		return buf[0:5]
	}

	buf[0] = uint8(0xff)
	binary.LittleEndian.PutUint64(buf[1:9], val)
	return buf[0:9]
}

func BE_Uint16ToBytes(i uint16) []byte {
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint16(tmp, i)
	return tmp
}

func BE_BytesToUint16(buf []byte) uint16 {
	return binary.BigEndian.Uint16(buf)
}

func LE_Uint16ToBytes(i uint16) []byte {
	tmp := make([]byte, 8)
	binary.LittleEndian.PutUint16(tmp, i)
	return tmp
}

func LE_BytesToUint16(buf []byte) uint16 {
	return binary.LittleEndian.Uint16(buf)
}
