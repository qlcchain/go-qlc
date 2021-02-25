/*
 * Copyright (c) 2021 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"math/big"
	"reflect"
	"testing"
)

func Test_stringToInt64(t *testing.T) {
	type args struct {
		amount string
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "1e12",
			args: args{
				amount: "1e12",
			},
			want: big.NewInt(1e12),
		},
		{
			name: "zero",
			args: args{
				amount: "0",
			},
			want: big.NewInt(0),
		}, {
			name: "1000",
			args: args{
				amount: "1000",
			},
			want: big.NewInt(1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringToInt64(tt.args.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stringToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isZero(t *testing.T) {
	type args struct {
		amount *big.Int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil",
			args: args{
				amount: nil,
			},
			want: true,
		},
		{
			name: "zero",
			args: args{
				amount: big.NewInt(0),
			},
			want: true,
		}, {
			name: "10",
			args: args{
				amount: big.NewInt(10),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isZero(tt.args.amount); got != tt.want {
				t.Errorf("isZero() = %v, want %v", got, tt.want)
			}
		})
	}
}
