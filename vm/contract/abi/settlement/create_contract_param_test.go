/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestCreateContractParam_Address(t *testing.T) {
	cp := createContractParam
	if address, err := cp.Address(); err != nil {
		t.Fatal(err)
	} else {
		cp2 := createContractParam
		if address2, err := cp2.Address(); err != nil {
			t.Fatal(err)
		} else {
			if address != address2 {
				t.Fatalf("invalid address, %v, %v", address, address2)
			}
		}
	}
}

func TestCreateContractParam_String(t *testing.T) {
	param := buildContractParam()
	s := param.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}
}

func TestCreateContractParam_ToContractParam(t *testing.T) {
	cp := createContractParam
	type fields struct {
		PartyA    Contractor
		PartyB    Contractor
		Previous  types.Hash
		Services  []ContractService
		SignDate  int64
		StartDate int64
		EndDate   int64
	}
	tests := []struct {
		name   string
		fields fields
		want   *ContractParam
	}{
		{
			name: "OK",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want: &ContractParam{
				CreateContractParam: cp,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:    tt.fields.PartyA,
				PartyB:    tt.fields.PartyB,
				Previous:  tt.fields.Previous,
				Services:  tt.fields.Services,
				SignDate:  tt.fields.SignDate,
				StartDate: tt.fields.StartDate,
				EndDate:   cp.EndDate,
			}
			if got := z.ToContractParam(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToContractParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateContractParam_Verify(t *testing.T) {
	cp := createContractParam

	type fields struct {
		PartyA    Contractor
		PartyB    Contractor
		Previous  types.Hash
		Services  []ContractService
		SignDate  int64
		StartDate int64
		EndDate   int64
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
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "f1",
			fields: fields{
				PartyA: Contractor{
					Address: types.ZeroAddress,
					Name:    "CC",
				},
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				PartyA: Contractor{
					Address: mock.Address(),
					Name:    "",
				},
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f3",
			fields: fields{
				PartyA: cp.PartyA,
				PartyB: Contractor{
					Address: mock.Address(),
					Name:    "",
				},
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f4",
			fields: fields{
				PartyA: cp.PartyA,
				PartyB: Contractor{
					Address: types.ZeroAddress,
					Name:    "CC",
				},
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f5",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  types.ZeroHash,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f6",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  nil,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f7",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  0,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f8",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: 0,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f9",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   0,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f10",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.EndDate,
				EndDate:   cp.StartDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f11",
			fields: fields{
				PartyA:   cp.PartyA,
				PartyB:   cp.PartyB,
				Previous: cp.Previous,
				Services: []ContractService{
					{
						ServiceId:   "",
						Mcc:         1,
						Mnc:         2,
						TotalAmount: 100,
						UnitPrice:   2,
						Currency:    "USD",
					},
				},
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:    tt.fields.PartyA,
				PartyB:    tt.fields.PartyB,
				Previous:  tt.fields.Previous,
				Services:  tt.fields.Services,
				SignDate:  tt.fields.SignDate,
				StartDate: tt.fields.StartDate,
				EndDate:   tt.fields.EndDate,
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

func TestContractService_Balance(t *testing.T) {
	type fields struct {
		ServiceId   string
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		UnitPrice   float64
		Currency    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    types.Balance
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				ServiceId:   mock.Hash().String(),
				Mcc:         22,
				Mnc:         1,
				TotalAmount: 100,
				UnitPrice:   0.04,
				Currency:    "USD",
			},
			want:    types.Balance{Int: big.NewInt(1e8)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractService{
				ServiceId:   tt.fields.ServiceId,
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
			}
			got, err := z.Balance()
			if (err != nil) != tt.wantErr {
				t.Errorf("Balance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Balance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateContractParam_Balance(t *testing.T) {
	type fields struct {
		PartyA    Contractor
		PartyB    Contractor
		Previous  types.Hash
		Services  []ContractService
		SignDate  int64
		StartDate int64
		EndData   int64
	}
	tests := []struct {
		name    string
		fields  fields
		want    types.Balance
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				PartyA: Contractor{
					Address: mock.Address(),
					Name:    "PCCWG",
				},
				PartyB: Contractor{
					Address: mock.Address(),
					Name:    "HKTCSL",
				},
				Previous: mock.Hash(),
				Services: []ContractService{{
					ServiceId:   mock.Hash().String(),
					Mcc:         1,
					Mnc:         2,
					TotalAmount: 100,
					UnitPrice:   2,
					Currency:    "USD",
				}, {
					ServiceId:   mock.Hash().String(),
					Mcc:         22,
					Mnc:         1,
					TotalAmount: 300,
					UnitPrice:   4,
					Currency:    "USD",
				}},
				SignDate:  time.Now().Unix(),
				StartDate: time.Now().AddDate(0, 0, 2).Unix(),
				EndData:   time.Now().AddDate(1, 0, 2).Unix(),
			},
			want:    types.Balance{Int: big.NewInt(2 * 1e8)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:    tt.fields.PartyA,
				PartyB:    tt.fields.PartyB,
				Previous:  tt.fields.Previous,
				Services:  tt.fields.Services,
				SignDate:  tt.fields.SignDate,
				StartDate: tt.fields.StartDate,
				EndDate:   tt.fields.EndData,
			}
			got, err := z.Balance()
			if (err != nil) != tt.wantErr {
				t.Errorf("Balance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Balance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateContractParam_ToABI(t *testing.T) {
	cp := createContractParam

	abi, err := cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	cp2 := &CreateContractParam{}
	if err = cp2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		a1, err := cp.Address()
		if err != nil {
			t.Fatal(err)
		}
		a2, err := cp2.Address()
		if err != nil {
			t.Fatal(err)
		}
		if a1 != a2 {
			t.Fatalf("invalid create contract params, %v, %v", cp, cp2)
		} else {
			t.Log(cp2.String())
		}
	}
}
