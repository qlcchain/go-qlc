/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

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
			name: "",
			x:    ContractStatusActiveStage1,
			want: "ActiveStage1",
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

func TestDLRStatusNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "ok",
			want: []string{"Delivered", "Rejected", "Unknown", "Undelivered", "Empty"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DLRStatusNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DLRStatusNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDLRStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		x       DLRStatus
		want    []byte
		wantErr bool
	}{
		{
			name:    "",
			x:       DLRStatusDelivered,
			want:    []byte("Delivered"),
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

func TestDLRStatus_String(t *testing.T) {
	tests := []struct {
		name string
		x    DLRStatus
		want string
	}{
		{
			name: "ok",
			x:    DLRStatusDelivered,
			want: "Delivered",
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

func TestDLRStatus_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		x       DLRStatus
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			x:    DLRStatusDelivered,
			args: args{
				text: []byte("Delivered"),
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

func TestParseDLRStatus(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    DLRStatus
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				name: "Delivered",
			},
			want:    DLRStatusDelivered,
			wantErr: false,
		}, {
			name: "fail",
			args: args{
				name: "Delivereddd",
			},
			want:    DLRStatusDelivered,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDLRStatus(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDLRStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDLRStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSendingStatus(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    SendingStatus
		wantErr bool
	}{
		{
			name: "",
			args: args{
				name: "Sent",
			},
			want:    SendingStatusSent,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSendingStatus(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSendingStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSendingStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			name: "",
			args: args{
				name: "stage1",
			},
			want:    SettlementStatusStage1,
			wantErr: false,
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

func TestSendingStatusNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "OK",
			want: []string{"Sent", "Error", "Empty"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SendingStatusNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SendingStatusNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendingStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		x       SendingStatus
		want    []byte
		wantErr bool
	}{
		{
			name:    "",
			x:       SendingStatusSent,
			want:    []byte("Sent"),
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

func TestSendingStatus_String(t *testing.T) {
	tests := []struct {
		name string
		x    SendingStatus
		want string
	}{
		{
			name: "",
			x:    SendingStatusSent,
			want: "Sent",
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

func TestSendingStatus_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		x       SendingStatus
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			x:    SendingStatusSent,
			args: args{
				text: []byte("Sent"),
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
