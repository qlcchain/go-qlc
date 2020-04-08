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

func TestContractStatusNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "ok",
			want: []string{"ActiveStage1", "Activated", "DestroyStage1", "Destroyed", "Rejected"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContractStatusNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContractStatusNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContractStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		x       ContractStatus
		want    []byte
		wantErr bool
	}{
		{
			name:    "ok",
			x:       ContractStatusActiveStage1,
			want:    []byte("ActiveStage1"),
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

func TestContractStatus_String(t *testing.T) {
	tests := []struct {
		name string
		x    ContractStatus
		want string
	}{
		{
			name: "ok",
			x:    ContractStatusActiveStage1,
			want: "ActiveStage1",
		}, {
			name: "",
			x:    6,
			want: "ContractStatus(6)",
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

func TestContractStatus_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		x       ContractStatus
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			x:    ContractStatusActiveStage1,
			args: args{
				text: []byte("ActiveStage1"),
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

func TestParseContractStatus(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    ContractStatus
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				name: "ActiveStage1",
			},
			want:    ContractStatusActiveStage1,
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				name: "ActiveStage21",
			},
			want:    ContractStatusActiveStage1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseContractStatus(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContractStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseContractStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}
