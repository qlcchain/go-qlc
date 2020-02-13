/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"testing"
)

func Test_calculateRange(t *testing.T) {
	type args struct {
		size   int
		count  int
		offset *int
	}
	tests := []struct {
		name      string
		args      args
		wantStart int
		wantEnd   int
		wantErr   bool
	}{
		{
			name: "f1",
			args: args{
				size:   2,
				count:  2,
				offset: offset(0),
			},
			wantStart: 0,
			wantEnd:   2,
			wantErr:   false,
		}, {
			name: "f2",
			args: args{
				size:   2,
				count:  10,
				offset: nil,
			},
			wantStart: 0,
			wantEnd:   2,
			wantErr:   false,
		}, {
			name: "overflow",
			args: args{
				size:   2,
				count:  10,
				offset: offset(2),
			},
			wantStart: 0,
			wantEnd:   0,
			wantErr:   true,
		}, {
			name: "f3",
			args: args{
				size:   2,
				count:  10,
				offset: offset(1),
			},
			wantStart: 1,
			wantEnd:   2,
			wantErr:   false,
		}, {
			name: "f4",
			args: args{
				size:   2,
				count:  0,
				offset: offset(1),
			},
			wantStart: 0,
			wantEnd:   0,
			wantErr:   true,
		}, {
			name: "f5",
			args: args{
				size:   2,
				count:  0,
				offset: offset(-1),
			},
			wantStart: 0,
			wantEnd:   0,
			wantErr:   true,
		}, {
			name: "f6",
			args: args{
				size:   2,
				count:  -1,
				offset: offset(0),
			},
			wantStart: 0,
			wantEnd:   0,
			wantErr:   true,
		}, {
			name: "f7",
			args: args{
				size:   10,
				count:  3,
				offset: offset(3),
			},
			wantStart: 3,
			wantEnd:   6,
			wantErr:   false,
		}, {
			name: "f8",
			args: args{
				size:   0,
				count:  3,
				offset: offset(3),
			},
			wantStart: 0,
			wantEnd:   0,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd, err := calculateRange(tt.args.size, tt.args.count, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStart != tt.wantStart {
				t.Errorf("calculateRange() gotStart = %v, want %v", gotStart, tt.wantStart)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("calculateRange() gotEnd = %v, want %v", gotEnd, tt.wantEnd)
			}
		})
	}
}

func offset(o int) *int {
	return &o
}
