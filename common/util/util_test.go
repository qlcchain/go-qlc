/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestRandomFixedString(t *testing.T) {
	l := 16
	pw := RandomFixedString(l)
	if len(pw) != l {
		t.Fatal("invalid len", len(pw))
	}
	t.Log(pw)
}

func TestVerifyEmailFormat(t *testing.T) {
	e1 := "11@qq.com"
	e2 := "2222@com"
	e3 := " abc.d@qlink.online"

	if !VerifyEmailFormat(e1) {
		t.Fatal()
	}

	if VerifyEmailFormat(e2) {
		t.Fatal()
	}

	if !VerifyEmailFormat(e3) {
		t.Fatal()
	}
}

func TestRandomFixedStringWithSeed(t *testing.T) {
	s := RandomFixedStringWithSeed(10, 100)
	if len(s) == 0 {
		t.Fatal()
	}
}

func TestReverseBytes(t *testing.T) {
	ori := []byte{1, 2, 3}
	rev := []byte{3, 2, 1}
	b := ReverseBytes(ori)
	if !bytes.Equal(rev, b) {
		t.Fatalf("exp: %v, act: %v", rev, b)
	}
}

func TestHexToBytes(t *testing.T) {
	b := []byte{1, 3, 4}
	s := hex.EncodeToString(b)
	b2 := HexToBytes(s)
	if !bytes.Equal(b, b2) {
		t.Fatalf("exp: %v, act: %v", b, b2)
	}
}

func TestTrimQuotes(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ok",
			args: args{
				s: `"123"`,
			},
			want: "123",
		}, {
			name: "ok",
			args: args{
				s: `"`,
			},
			want: `"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimQuotes(tt.args.s); got != tt.want {
				t.Errorf("TrimQuotes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToString(t *testing.T) {
	type tmp struct {
		I int64
		J int64
	}

	a := &tmp{
		I: 11,
		J: 22,
	}
	s1 := ToString(a)
	s2 := ToIndentString(a)
	t.Log(s1, s2)
}
