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

func TestContractAddressList_Append(t *testing.T) {
	a1 := mock.Address()
	a2 := mock.Address()
	cl := newContractAddressList(&a1)

	type fields struct {
		AddressList []*types.Address
	}
	type args struct {
		address *types.Address
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "ok",
			fields: fields{
				AddressList: cl.AddressList,
			},
			args: args{
				address: &a2,
			},
			want: true,
		}, {
			name: "exist",
			fields: fields{
				AddressList: cl.AddressList,
			},
			args: args{
				address: &a1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractAddressList{
				AddressList: tt.fields.AddressList,
			}
			if got := z.Append(tt.args.address); got != tt.want {
				t.Errorf("Append() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContractAddressList_ToABI(t *testing.T) {
	a1 := mock.Address()
	cl := newContractAddressList(&a1)

	if abi, err := cl.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		cl2 := &ContractAddressList{}
		if err := cl2.FromABI(abi); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(cl, cl2) {
				t.Fatalf("invalid %v,%v", cl, cl2)
			} else {
				t.Log(cl.String())
			}
		}
	}
}
