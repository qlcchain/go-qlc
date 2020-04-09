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
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestSignContractParam_ToABI(t *testing.T) {
	sc := &SignContractParam{
		ContractAddress: mock.Address(),
		ConfirmDate:     time.Now().Unix(),
	}

	if abi, err := sc.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		sc2 := &SignContractParam{}
		if err := sc2.FromABI(abi); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(sc, sc2) {
				t.Errorf("ToABI() got = %v, want %v", sc2, sc)
			}
		}
	}
}

func TestSignContractParam_Verify(t *testing.T) {
	type fields struct {
		ContractAddress types.Address
		ConfirmDate     int64
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				ContractAddress: mock.Address(),
				ConfirmDate:     time.Now().Unix(),
			},
			want:    true,
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				ContractAddress: types.ZeroAddress,
				ConfirmDate:     time.Now().Unix(),
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				ContractAddress: mock.Address(),
				ConfirmDate:     0,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &SignContractParam{
				ContractAddress: tt.fields.ContractAddress,
				ConfirmDate:     tt.fields.ConfirmDate,
			}
			got, err := z.Verify()
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Verify() got = %v, want %v", got, tt.want)
			}
		})
	}
}
