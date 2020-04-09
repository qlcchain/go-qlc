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

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestTerminateParam_ToABI(t *testing.T) {
	param := &TerminateParam{ContractAddress: mock.Address(), Request: true}
	if abi, err := param.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		p2 := &TerminateParam{}
		if err := p2.FromABI(abi); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(param, p2) {
				t.Fatalf("invalid param, %v, %v", param, p2)
			} else {
				t.Log(param.String())
			}
		}
	}
}

func TestTerminateParam_Verify(t *testing.T) {
	type fields struct {
		ContractAddress types.Address
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				ContractAddress: mock.Address(),
			},
			wantErr: false,
		}, {
			name: "fail",
			fields: fields{
				ContractAddress: types.ZeroAddress,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &TerminateParam{
				ContractAddress: tt.fields.ContractAddress,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
