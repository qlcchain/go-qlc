/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

import (
	"math/big"
	"reflect"
	"testing"
)

func TestStringToBigInt(t *testing.T) {
	s := "100"
	type args struct {
		str *string
	}
	tests := []struct {
		name    string
		args    args
		want    *big.Int
		wantErr bool
	}{
		{
			name: "",
			args: args{
				str: &s,
			},
			want:    big.NewInt(100),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringToBigInt(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToBigInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToBigInt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaddedBigBytes(t *testing.T) {
	type args struct {
		bigint *big.Int
		n      int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ok",
			args: args{
				bigint: big.NewInt(1),
				n:      2,
			},
			want: []byte{00, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PaddedBigBytes(tt.args.bigint, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PaddedBigBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
