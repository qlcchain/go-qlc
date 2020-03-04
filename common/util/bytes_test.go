// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package util

import (
	"bytes"
	"testing"
)

func TestCopyBytes(t *testing.T) {
	data1 := []byte{1, 2, 3, 4}
	exp1 := []byte{1, 2, 3, 4}
	res1 := CopyBytes(data1)
	if !bytes.Equal(res1, exp1) {
		t.Fatalf("exp: %v, act: %v", exp1, res1)
	}
}

func TestLeftPadBytes(t *testing.T) {
	val1 := []byte{1, 2, 3, 4}
	exp1 := []byte{0, 0, 0, 0, 1, 2, 3, 4}

	res1 := LeftPadBytes(val1, 8)
	res2 := LeftPadBytes(val1, 2)
	if !bytes.Equal(res1, exp1) {
		t.Fatalf("exp: %v, act: %v", exp1, res1)
	}
	if !bytes.Equal(res2, val1) {
		t.Fatalf("exp: %v, act: %v", val1, res2)
	}
}

func TestRightPadBytes(t *testing.T) {
	val := []byte{1, 2, 3, 4}
	exp := []byte{1, 2, 3, 4, 0, 0, 0, 0}

	resstd := RightPadBytes(val, 8)
	resshrt := RightPadBytes(val, 2)
	if !bytes.Equal(resstd, exp) {
		t.Fatalf("exp: %v, act: %v", exp, resstd)
	}
	if !bytes.Equal(resshrt, val) {
		t.Fatalf("exp: %v, act: %v", val, resshrt)
	}
}

func TestFromHex(t *testing.T) {
	input := "0x01"
	expected := []byte{1}
	result := FromHex(input)
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %x got %x", expected, result)
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"", true},
		{"0", false},
		{"00", true},
		{"a9e67e", true},
		{"A9E67E", true},
		{"0xa9e67e", false},
		{"a9e67e001", false},
		{"0xHELLO_MY_NAME_IS_STEVEN_@#$^&*", false},
	}
	for _, test := range tests {
		if ok := isHex(test.input); ok != test.ok {
			t.Errorf("isHex(%q) = %v, want %v", test.input, ok, test.ok)
		}
	}
}

func TestFromHexOddLength(t *testing.T) {
	input := "0x1"
	expected := []byte{1}
	result := FromHex(input)
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %x got %x", expected, result)
	}
}

func TestNoPrefixShortHexOddLength(t *testing.T) {
	input := "1"
	expected := []byte{1}
	result := FromHex(input)
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %x got %x", expected, result)
	}
}

func TestBE_Uint32ToBytes(t *testing.T) {
	i := uint32(100)
	data := BE_Uint32ToBytes(i)
	j := BE_BytesToUint32(data)
	if i != j {
		t.Fatalf("invalid i, exp: %d, act: %d", i, j)
	}
}

func TestLE_Uint32ToBytes(t *testing.T) {
	i := uint32(100)
	data := LE_Uint32ToBytes(i)
	j := LE_BytesToUint32(data)
	if i != j {
		t.Fatalf("invalid i, exp: %d, act: %d", i, j)
	}

}

func TestBE_Uint64ToBytes(t *testing.T) {
	i := uint64(100)
	data := BE_Uint64ToBytes(i)
	j := BE_BytesToUint64(data)
	if i != j {
		t.Fatalf("invalid i, exp: %d, act: %d", i, j)
	}
}

func TestLE_Uint64ToBytes(t *testing.T) {
	i := uint64(100)
	data := LE_Uint64ToBytes(i)
	j := LE_BytesToUint64(data)
	if i != j {
		t.Fatalf("invalid i, exp: %d, act: %d", i, j)
	}
}

func TestLE_Int64ToBytes(t *testing.T) {
	i := int64(100)
	data := LE_Int64ToBytes(i)
	j := LE_BytesToInt64(data)
	if i != j {
		t.Fatalf("invalid i, exp: %d, act: %d", i, j)
	}
}

func TestBE_Int2Bytes(t *testing.T) {
	i := int64(100)
	data := BE_Int2Bytes(i)
	j := BE_Bytes2Int(data)
	if i != j {
		t.Fatalf("invalid i, exp: %d, act: %d", i, j)
	}
}

func TestString2Bytes(t *testing.T) {
	b := String2Bytes("hahah")
	if len(b) == 0 {
		t.Fatal()
	}
}

func TestBool2Bytes(t *testing.T) {
	b1 := Bool2Bytes(true)
	t.Log(b1)
}
