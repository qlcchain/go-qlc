package contractaddress

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestIsChainContractAddress(t *testing.T) {
	type args struct {
		address types.Address
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ok",
			args: args{
				address: RewardsAddress,
			},
			want: true,
		}, {
			name: "ok",
			args: args{
				address: types.Address{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsChainContractAddress(tt.args.address); got != tt.want {
				t.Errorf("IsChainContractAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsContractAddress(t *testing.T) {
	type args struct {
		address types.Address
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ok",
			args: args{
				address: RewardsAddress,
			},
			want: true,
		}, {
			name: "ok",
			args: args{
				address: types.Address{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContractAddress(tt.args.address); got != tt.want {
				t.Errorf("IsContractAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRewardContractAddress(t *testing.T) {
	type args struct {
		address types.Address
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "fail",
			args: args{
				address: types.Address{},
			},
			want: false,
		}, {
			name: "ok",
			args: args{
				address: MinerAddress,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRewardContractAddress(tt.args.address); got != tt.want {
				t.Errorf("IsRewardContractAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
