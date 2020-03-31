/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"bytes"
	"testing"
)

func TestContractKey_Serialize(t *testing.T) {
	a1, _, _ := GenerateAddress()
	a2, _, _ := GenerateAddress()
	var hash Hash
	s := "2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"
	_ = hash.Of(s)

	type fields struct {
		ContractAddress Address
		AccountAddress  Address
		Hash            Hash
		Suffix          []byte
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"1", fields{
			ContractAddress: a1,
			AccountAddress:  a2,
			Hash:            ZeroHash,
			Suffix:          nil,
		}, false},
		{"2", fields{
			ContractAddress: a1,
			AccountAddress:  a2,
			Hash:            hash,
			Suffix:          nil,
		}, false},
		{"3", fields{
			ContractAddress: a1,
			AccountAddress:  a2,
			Hash:            ZeroHash,
			Suffix:          []byte{0, 1, 2},
		}, false},
		{"4", fields{
			ContractAddress: a1,
			AccountAddress:  a2,
			Hash:            hash,
			Suffix:          []byte{0, 1, 3},
		}, false},
	}
	for idx, test := range tests {
		tt := test
		i := idx

		t.Run(tt.name, func(t *testing.T) {
			z := &ContractKey{
				ContractAddress: tt.fields.ContractAddress,
				AccountAddress:  tt.fields.AccountAddress,
				Hash:            tt.fields.Hash,
				Suffix:          tt.fields.Suffix,
			}
			got, err := z.Serialize()
			if (err != nil) != tt.wantErr {
				t.Errorf("Serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := &ContractKey{}
			if err = want.Deserialize(got); err != nil {
				t.Fatalf("%s[%d]: Deserialize failed, %s", tt.name, i, err)
			} else {
				if want.ContractAddress != tt.fields.ContractAddress {
					t.Fatalf("invalid contract address, exp: %s, act: %s",
						tt.fields.ContractAddress, want.ContractAddress)
				}
				if want.AccountAddress != tt.fields.AccountAddress {
					t.Fatalf("invalid account address, exp: %s, act: %s",
						tt.fields.AccountAddress, want.AccountAddress)
				}

				if !want.Hash.IsZero() {
					if !bytes.EqualFold(want.Hash[:], tt.fields.Hash[:]) {
						t.Fatalf("exp: %s, act: %s", tt.fields.Hash[:], want.Hash[:])
					}
				} else {
					if !tt.fields.Hash.IsZero() {
						t.Fatalf("hash should be nil,%s", tt.fields.Hash)
					}
				}

				if len(tt.fields.Suffix) > 0 {
					if !bytes.EqualFold(want.Suffix[:], tt.fields.Suffix) {
						t.Fatalf("exp: %s, act: %s", tt.fields.Suffix, want.Suffix[:])
					}
				} else {
					if want.Suffix != nil {
						t.Fatalf("suffix should be nil, %v", want.Suffix)
					}
				}
			}
		})
	}
}

func TestContractKey_String(t *testing.T) {
	ck := &ContractKey{
		ContractAddress: Address{},
		AccountAddress:  Address{},
		Hash:            ZeroHash,
		Suffix:          nil,
	}
	t.Log(ck.String())
}
