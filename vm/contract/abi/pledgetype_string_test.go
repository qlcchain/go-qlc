/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import "testing"

func TestPledgeType_String(t *testing.T) {
	tests := []struct {
		name string
		i    PledgeType
		want string
	}{
		{
			name: "Network",
			i:    Network,
			want: "Network",
		}, {
			name: "overflow",
			i:    5,
			want: "PledgeType(5)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
