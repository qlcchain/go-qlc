/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"fmt"
	"testing"

	"gopkg.in/validator.v2"
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
		{name: "empty PovCfg config", args: PoVConfig{
			PovEnabled:   false,
			MinerEnabled: false,
			Coinbase:     "",
		}, wantErr: false}, {name: "invalid PovCfg config", args: PoVConfig{
			PovEnabled:   false,
			MinerEnabled: false,
			Coinbase:     "123",
		}, wantErr: true}, {name: "valid PovCfg config", args: PoVConfig{
			PovEnabled:   false,
			MinerEnabled: false,
			Coinbase:     "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
		}, wantErr: false},
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
		{name: "empty PovCfg config inside cfg", args: config{
			PovCfg: &PoVConfig{
				PovEnabled:   false,
				MinerEnabled: false,
				Coinbase:     "",
			},
			Padding: "111",
		}, wantErr: false},
		{name: "invalid PovCfg config  inside cfg", args: config{
			PovCfg: &PoVConfig{
				PovEnabled:   false,
				MinerEnabled: false,
				Coinbase:     "123",
			}, Padding: "112",
		}, wantErr: true},
		{name: "valid PovCfg config  inside cfg", args: config{
			PovCfg: &PoVConfig{
				PovEnabled:   false,
				MinerEnabled: false,
				Coinbase:     "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
			}, Padding: "",
		}, wantErr: true},
		{name: "valid PovCfg config  inside cfg", args: config{
			PovCfg: &PoVConfig{
				PovEnabled:   false,
				MinerEnabled: false,
				Coinbase:     "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
			}, Padding: "111",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.Validate(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Address() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestErrToString(t *testing.T) {
	// First create a struct to be validated
	// according to the validator tags.
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
	if err == nil {
		fmt.Println("Values are valid.")
	} else {
		fmt.Println(ErrToString(err))
	}

	// Output:
	// Invalid due to fields:
	//	 - Address.Street (zero value)
	// 	 - Age (less than min)
	// 	 - Email (regular expression mismatch)
}
