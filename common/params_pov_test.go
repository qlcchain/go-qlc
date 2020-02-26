/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestPovIsAlgoSupported(t *testing.T) {
	type args struct {
		algoType types.PovAlgoType
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ok",
			args: args{
				algoType: types.ALGO_SHA256D,
			},
			want: true,
		}, {
			name: "f",
			args: args{
				algoType: 100,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PovIsAlgoSupported(tt.args.algoType); got != tt.want {
				t.Errorf("PovIsAlgoSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}
