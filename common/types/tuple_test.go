/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"reflect"
	"testing"
)

func TestNewTuple(t *testing.T) {
	type args struct {
		first  interface{}
		second interface{}
	}
	tests := []struct {
		name string
		args args
		want *Tuple
	}{
		{"", args{first: 1, second: "111"}, &Tuple{First: 1, Second: "111"}},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTuple(tt.args.first, tt.args.second); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTuple() = %v, want %v", got, tt.want)
			}
		})
	}
}
