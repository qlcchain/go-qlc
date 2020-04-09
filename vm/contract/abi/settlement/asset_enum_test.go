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
