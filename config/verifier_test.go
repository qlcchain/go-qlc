/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"errors"
	"testing"

	"gopkg.in/validator.v2"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
)

type config struct {
	Padding string `validate:"nonzero"`
	PovCfg  *PoVConfig
}

func TestValidateStructure(t *testing.T) {
	tests := []struct {
		name    string
		args    PoVConfig
		wantErr bool
	}{
		{
			name: "empty PovCfg config",
			args: PoVConfig{
				PovEnabled:   false,
				MinerEnabled: false,
				Coinbase:     "",
			},
			wantErr: false,
		}, {
			name: "invalid PovCfg config",
			args: PoVConfig{
				PovEnabled:   false,
				MinerEnabled: false,
				Coinbase:     "123",
			},
			wantErr: true,
		}, {
			name: "valid PovCfg config",
			args: PoVConfig{
				PovEnabled:   false,
				MinerEnabled: false,
				Coinbase:     "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.Validate(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Address() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		args    config
		wantErr bool
	}{
		{
			name: "empty PovCfg config inside cfg",
			args: config{
				PovCfg: &PoVConfig{
					PovEnabled:   false,
					MinerEnabled: false,
					Coinbase:     "",
				},
				Padding: "111",
			},
			wantErr: false,
		}, {
			name: "invalid PovCfg config  inside cfg",
			args: config{
				PovCfg: &PoVConfig{
					PovEnabled:   false,
					MinerEnabled: false,
					Coinbase:     "123",
				},
				Padding: "112",
			},
			wantErr: true,
		},
		{
			name: "valid PovCfg config  inside cfg",
			args: config{
				PovCfg: &PoVConfig{
					PovEnabled:   false,
					MinerEnabled: false,
					Coinbase:     "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
				},
				Padding: "",
			},
			wantErr: true,
		}, {
			name: "valid PovCfg config  inside cfg",
			args: config{
				PovCfg: &PoVConfig{
					PovEnabled:   false,
					MinerEnabled: false,
					Coinbase:     "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
				},
				Padding: "",
			},
			wantErr: true,
		},
		{
			name: "valid PovCfg config  inside cfg",
			args: config{
				PovCfg: &PoVConfig{
					PovEnabled:   false,
					MinerEnabled: false,
					Coinbase:     "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
				},
				Padding: "111",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.Validate(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Address() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_hash(t *testing.T) {
	type h1 struct {
		Hash *types.Hash `validate:"hash"`
	}

	type h2 struct {
		Hash types.Hash `validate:"hash"`
	}

	type h3 struct {
		Hash types.Address `validate:"hash"`
	}

	tests := []struct {
		name    string
		args    interface{}
		wantErr bool
	}{
		{
			name: "f1",
			args: h1{
				Hash: &types.ZeroHash,
			},
			wantErr: true,
		}, {
			name: "f2",
			args: h2{
				Hash: types.ZeroHash,
			},
			wantErr: true,
		}, {
			name: "f3",
			args: &h1{
				Hash: &types.ZeroHash,
			},
			wantErr: true,
		}, {
			name: "f4",
			args: h3{
				Hash: types.ZeroAddress,
			},
			wantErr: true,
		}, {
			name: "ok",
			args: h2{
				Hash: types.FFFFHash,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.Validate(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Hash() error = %v, wantErr %v", ErrToString(err), tt.wantErr)
			}
		})
	}
}

func TestAddress(t *testing.T) {
	type a1 struct {
		Address *types.Address `validate:"qlcaddress"`
	}

	type a2 struct {
		Address types.Address `validate:"qlcaddress"`
	}

	type a3 struct {
		Address types.Hash `validate:"qlcaddress"`
	}

	tests := []struct {
		name    string
		args    interface{}
		wantErr bool
	}{
		{
			name: "f1",
			args: a1{
				Address: &types.ZeroAddress,
			},
			wantErr: true,
		}, {
			name: "f2",
			args: a2{
				Address: types.ZeroAddress,
			},
			wantErr: true,
		}, {
			name: "f3",
			args: &a1{
				Address: &types.ZeroAddress,
			},
			wantErr: true,
		}, {
			name: "f4",
			args: a3{
				Address: types.ZeroHash,
			},
			wantErr: true,
		}, {
			name: "ok",
			args: a2{
				Address: contractaddress.MintageAddress,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.Validate(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Hash() error = %v, wantErr %v", ErrToString(err), tt.wantErr)
			}
		})
	}
}

func TestErrToString(t *testing.T) {
	type ValidateExample struct {
		Name        string `validate:"nonzero"`
		Description string
		Age         int    `validate:"min=18"`
		Email       string `validate:"regexp=^[0-9a-z]+@[0-9a-z]+(\\.[0-9a-z]+)+$"`
		Address     struct {
			Street string `validate:"nonzero"`
			City   string `validate:"nonzero"`
		}
	}

	// Fill in some values
	ve := ValidateExample{
		Name:        "Joe Doe", // valid as it's nonzero
		Description: "",        // valid no validation tag exists
		Age:         17,        // invalid as age is less than required 18
	}
	// invalid as Email won't match the regular expression
	ve.Email = "@not.a.valid.email"
	ve.Address.City = "Some City" // valid
	ve.Address.Street = ""        // invalid

	err := validator.Validate(ve)
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "c1",
			args: args{
				err: nil,
			},
			want: "",
		}, {
			name: "c2",
			args: args{
				err: errors.New("invalid"),
			},
			want: "invalid",
		}, {
			name: "c3",
			args: args{
				err: err,
			},
			want: `Invalid due to fields:
    - Address.Street (zero value)
    - Age (less than min)
    - Email (regular expression mismatch)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrToString(tt.args.err)
			t.Log(got)
		})
	}
}
