/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"reflect"
	"testing"
)

func TestParseSettlementStatus(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    SettlementStatus
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				name: "stage1",
			},
			want:    SettlementStatusStage1,
			wantErr: false,
		}, {
			name: "invalid",
			args: args{
				name: "invalid",
			},
			want:    SettlementStatusUnknown,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSettlementStatus(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSettlementStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSettlementStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSettlementStatusNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "ok",
			want: []string{"unknown", "stage1", "success", "failure", "missing", "duplicate"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SettlementStatusNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SettlementStatusNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSettlementStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		x       SettlementStatus
		want    []byte
		wantErr bool
	}{
		{
			name:    "ok",
			x:       SettlementStatusStage1,
			want:    []byte("stage1"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.x.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalText() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSettlementStatus_String(t *testing.T) {
	tests := []struct {
		name string
		x    SettlementStatus
		want string
	}{
		{
			name: "ok",
			x:    SettlementStatusStage1,
			want: "stage1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSettlementStatus_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		x       SettlementStatus
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			x:    SettlementStatusStage1,
			args: args{
				text: []byte("stage1"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.x.UnmarshalText(tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
