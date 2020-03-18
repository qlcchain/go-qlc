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
		}, {
			name: "ok",
			x:    5,
			want: "DLRStatus(5)",
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
			name: "ok",
			args: args{
				name: "Sent",
			},
			want:    SendingStatusSent,
			wantErr: false,
		}, {
			name: "invalid",
			args: args{
				name: "invalid",
			},
			want:    SendingStatusSent,
			wantErr: true,
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
			name: "ok",
			x:    SendingStatusSent,
			want: "Sent",
		}, {
			name: "invalid",
			x:    4,
			want: "SendingStatus(4)",
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

func TestParseAssertStatus(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    AssetStatus
		wantErr bool
	}{
		{
			name: "Deactivated",
			args: args{
				name: "Deactivated",
			},
			want:    0,
			wantErr: false,
		}, {
			name: "Activated",
			args: args{
				name: "Activated",
			},
			want:    AssetStatusActivated,
			wantErr: false,
		}, {
			name: "invalid",
			args: args{
				name: "invalid",
			},
			want:    AssetStatusDeactivated,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAssetStatus(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAssertStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseAssertStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertStatus_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		x       AssetStatus
		args    args
		wantErr bool
	}{
		{
			name: "Activated",
			x:    AssetStatusActivated,
			args: args{
				text: []byte("Activated"),
			},
			wantErr: false,
		}, {
			name: "Deactivated",
			x:    AssetStatusDeactivated,
			args: args{
				text: []byte("Deactivated"),
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

func TestAssertStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		x       AssetStatus
		want    []byte
		wantErr bool
	}{
		{
			name:    "Activated",
			x:       AssetStatusActivated,
			want:    []byte("Activated"),
			wantErr: false,
		}, {
			name:    "Deactivated",
			x:       AssetStatusDeactivated,
			want:    []byte("Deactivated"),
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

func TestAssertStatusNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "ok",
			want: []string{"Deactivated", "Activated"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AssetStatusNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AssertStatusNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertStatus_String(t *testing.T) {
	tests := []struct {
		name string
		x    AssetStatus
		want string
	}{
		{
			name: "Deactivated",
			x:    0,
			want: "Deactivated",
		}, {
			name: "Deactivated",
			x:    2,
			want: "AssetStatus(2)",
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

func TestSLATypeNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "ok",
			want: []string{"DeliveredRate", "Latency"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SLATypeNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SLATypeNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSLAType_String(t *testing.T) {
	tests := []struct {
		name string
		x    SLAType
		want string
	}{
		{
			name: "DeliveredRate",
			x:    0,
			want: "DeliveredRate",
		}, {
			name: "invalid",
			x:    2,
			want: "SLAType(2)",
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

func TestParseSLAType(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    SLAType
		wantErr bool
	}{
		{
			name: "DeliveredRate",
			args: args{
				name: "DeliveredRate",
			},
			want:    0,
			wantErr: false,
		}, {
			name: "Latency",
			args: args{
				name: "Latency",
			},
			want:    SLATypeLatency,
			wantErr: false,
		}, {
			name: "invalid",
			args: args{
				name: "invalid",
			},
			want:    SLATypeDeliveredRate,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSLAType(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSLAType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSLAType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSLAType_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		x       SLAType
		want    []byte
		wantErr bool
	}{
		{
			name:    "DeliveredRate",
			x:       0,
			want:    []byte("DeliveredRate"),
			wantErr: false,
		}, {
			name:    "Latency",
			x:       SLATypeLatency,
			want:    []byte("Latency"),
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

func TestSLAType_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		x       SLAType
		args    args
		wantErr bool
	}{
		{
			name: "DeliveredRate",
			x:    0,
			args: args{
				text: []byte("DeliveredRate"),
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
