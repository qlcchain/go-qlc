/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

import (
	"bytes"
	"math/big"
	"testing"
)

func TestMin(t *testing.T) {
	tests := []struct {
		x, y, result uint64
	}{
		{1, 2, 1},
		{1, 1, 1},
		{2, 1, 1},
	}
	for _, test := range tests {
		result := Min(test.x, test.y)
		if result != test.result {
			t.Fatalf("get min fail, input: %v, %v, expected %v, got %v", test.x, test.y, test.result, result)
		}
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		x, y, result uint64
	}{
		{1, 2, 2},
		{1, 1, 1},
		{2, 1, 2},
	}
	for _, test := range tests {
		result := Max(test.x, test.y)
		if result != test.result {
			t.Fatalf("get max fail, input: %v, %v, expected %v, got %v", test.x, test.y, test.result, result)
		}
	}
}

func TestSafeAdd(t *testing.T) {
	tests := []struct {
		x, y, result uint64
		overflow     bool
	}{
		{1, 2, 3, false},
		{MaxUint64 - 1, 1, MaxUint64, false},
		{MaxUint64, 1, 0, true},
		{MaxUint64, 2, 1, true},
	}
	for _, test := range tests {
		result, overflow := SafeAdd(test.x, test.y)
		if result != test.result || overflow != test.overflow {
			t.Fatalf("safe add fail, input: %v, %v, expected [%v, %v], got [%v, %v]", test.x, test.y, test.result, test.overflow, result, overflow)
		}
	}
}

func TestSafeMul(t *testing.T) {
	tests := []struct {
		x, y, result uint64
		overflow     bool
	}{
		{1, 0, 0, false},
		{0, 1, 0, false},
		{1, 2, 2, false},
		{MaxUint64, 2, MaxUint64 - 1, true},
		{MaxUint64 / 2, 2, MaxUint64 - 1, false},
	}
	for _, test := range tests {
		result, overflow := SafeMul(test.x, test.y)
		if result != test.result || overflow != test.overflow {
			t.Fatalf("safe add fail, input: %v, %v, expected [%v, %v], got [%v, %v]", test.x, test.y, test.result, test.overflow, result, overflow)
		}
	}
}

func TestBigPow(t *testing.T) {
	tests := []struct {
		x, y   int64
		result *big.Int
	}{
		{0, 0, big.NewInt(1)},
		{0, 2, big.NewInt(0)},
		{2, 0, big.NewInt(1)},
		{2, 4, big.NewInt(16)},
		{2, 33, big.NewInt(8589934592)},
		{2, 66, new(big.Int).Mul(big.NewInt(8589934592), big.NewInt(8589934592))},
	}
	for _, test := range tests {
		result := BigPow(test.x, test.y)
		if result.Cmp(test.result) != 0 {
			t.Fatalf("big pow fail, %v ** %v = expected %v, got %v", test.x, test.y, test.result, result)
		}
	}
}

func TestReadBits(t *testing.T) {
	tests := []struct {
		data        *big.Int
		buf, result []byte
	}{
		{Tt256m1, make([]byte, 32), []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}},
		{Tt256m1, make([]byte, 33), []byte{0, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}},
		{Tt256m1, make([]byte, 2), []byte{255, 255}},
		{Big0, make([]byte, 2), []byte{0, 0}},
		{Big0, make([]byte, 32), []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
	}
	for _, test := range tests {
		ReadBits(test.data, test.buf)
		if !bytes.Equal(test.buf, test.result) {
			t.Fatalf("read bits fail, data: [%v], expected [%v], got [%v]", test.data, test.result, test.buf)
		}
	}
}

func TestU256(t *testing.T) {
	tests := []struct {
		input, result *big.Int
	}{
		{big.NewInt(0), big.NewInt(0)},
		{big.NewInt(1), big.NewInt(1)},
		{new(big.Int).Set(Tt256m1), new(big.Int).Set(Tt256m1)},
		{new(big.Int).Set(Tt256), big.NewInt(0)},
		{new(big.Int).Add(Tt256, Big1), big.NewInt(1)},
	}
	for _, test := range tests {
		result := U256(test.input)
		if result.Cmp(test.result) != 0 {
			t.Fatalf("get u256 fail, input: [%v], expected [%v], got [%v]", test.input, test.result, result)
		}
	}
}

func TestUInt32Max(t *testing.T) {
	t.Parallel()
	type args struct {
		x uint32
		y uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{"max1", args{x: uint32(1), y: uint32(2)}, uint32(2)},
		{"max2", args{x: uint32(3), y: uint32(1)}, uint32(3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UInt32Max(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUInt32Min(t *testing.T) {
	t.Parallel()
	type args struct {
		a uint32
		b uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{"min1", args{a: uint32(1), b: uint32(2)}, uint32(1)},
		{"min2", args{a: uint32(3), b: uint32(1)}, uint32(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UInt32Min(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("min() = %v, want %v", got, tt.want)
			}
		})
	}
}
