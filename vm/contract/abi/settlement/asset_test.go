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

func TestAssert_ToABI(t *testing.T) {
	asset := assetParam.Assets[0]

	if msg, err := asset.MarshalMsg(nil); err != nil {
		t.Fatal(err)
	} else {
		a := &Asset{}
		if _, err := a.UnmarshalMsg(msg); err != nil {
			t.Fatal(err)
		} else {
			if id1, err := asset.ToAssertID(); err != nil {
				t.Fatal(err)
			} else {
				if id2, err := a.ToAssertID(); err != nil {
					t.Fatal(err)
				} else if id1 != id2 {
					t.Fatalf("invalid unmarshal, exp: %v, act: %v", asset, a)
				}
			}
		}
	}
}
func TestAssert_ToAssertID(t *testing.T) {
	asset := assetParam.Assets[0]
	t.Log(asset.String())

	type fields struct {
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		SLAs        []*SLA
	}
	tests := []struct {
		name    string
		fields  fields
		want    types.Hash
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				Mcc:         asset.Mcc,
				Mnc:         asset.Mnc,
				TotalAmount: asset.TotalAmount,
				SLAs:        asset.SLAs,
			},
			want:    types.Hash{},
			wantErr: false,
		}, {
			name: "ok",
			fields: fields{
				Mcc:         asset.Mcc,
				Mnc:         asset.Mnc,
				TotalAmount: asset.TotalAmount,
				SLAs:        nil,
			},
			want:    types.Hash{},
			wantErr: false,
		}, {
			name: "ok",
			fields: fields{
				Mcc:         asset.Mcc,
				Mnc:         asset.Mnc,
				TotalAmount: asset.TotalAmount,
				SLAs: []*SLA{
					NewLatency(60*time.Second, nil),
				},
			},
			want:    types.Hash{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &Asset{
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				SLAs:        tt.fields.SLAs,
			}
			got, err := z.ToAssertID()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToAssertID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.IsZero() {
				t.Errorf("ToAssertID() got = %v", got)
			}
		})
	}
}

func TestAssertParam_FromABI(t *testing.T) {
	a := &assetParam
	t.Log(a.String())
	if abi, err := a.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		if param, err := ParseAssertParam(abi); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(a, param) {
			t.Fatalf("exp: %v, act: %v", a, param)
		}
	}
}

func TestSLA_Deserialize(t *testing.T) {
	sla := NewLatency(60*time.Second, []*Compensation{
		{
			Low:  50,
			High: 60,
			Rate: 10,
		},
		{
			Low:  60,
			High: 80,
			Rate: 20.5,
		},
	})

	if data, err := sla.Serialize(); err != nil {
		t.Fatal(err)
	} else {
		s := &SLA{}
		if err := s.Deserialize(data); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(sla, s) {
			t.Fatalf("invalid sla, exp: %v, act: %v", sla, s)
		} else {
			t.Log(s.String())
		}
	}
}

func TestAssetParam_Verify(t *testing.T) {
	type fields struct {
		Owner     Contractor
		Previous  types.Hash
		Asserts   []*Asset
		SignDate  int64
		StartDate int64
		EndDate   int64
		Status    AssetStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				Owner:     assetParam.Owner,
				Previous:  assetParam.Previous,
				Asserts:   assetParam.Assets,
				SignDate:  assetParam.SignDate,
				StartDate: assetParam.StartDate,
				EndDate:   assetParam.EndDate,
				Status:    assetParam.Status,
			},
			wantErr: false,
		}, {
			name: "contractor#address",
			fields: fields{
				Owner:     Contractor{Address: types.ZeroAddress, Name: "HKT-CSL"},
				Previous:  assetParam.Previous,
				Asserts:   assetParam.Assets,
				SignDate:  assetParam.SignDate,
				StartDate: assetParam.StartDate,
				EndDate:   assetParam.EndDate,
				Status:    assetParam.Status,
			},
			wantErr: true,
		}, {
			name: "contractor#name",
			fields: fields{
				Owner:     Contractor{Address: mock.Address(), Name: ""},
				Previous:  assetParam.Previous,
				Asserts:   assetParam.Assets,
				SignDate:  assetParam.SignDate,
				StartDate: assetParam.StartDate,
				EndDate:   assetParam.EndDate,
				Status:    assetParam.Status,
			},
			wantErr: true,
		}, {
			name: "previous_hash",
			fields: fields{
				Owner:     assetParam.Owner,
				Previous:  types.ZeroHash,
				Asserts:   assetParam.Assets,
				SignDate:  assetParam.SignDate,
				StartDate: assetParam.StartDate,
				EndDate:   assetParam.EndDate,
				Status:    assetParam.Status,
			},
			wantErr: true,
		}, {
			name: "asset_nil",
			fields: fields{
				Owner:     assetParam.Owner,
				Previous:  assetParam.Previous,
				Asserts:   nil,
				SignDate:  assetParam.SignDate,
				StartDate: assetParam.StartDate,
				EndDate:   assetParam.EndDate,
				Status:    assetParam.Status,
			},
			wantErr: true,
		}, {
			name: "sign_date",
			fields: fields{
				Owner:     assetParam.Owner,
				Previous:  assetParam.Previous,
				Asserts:   assetParam.Assets,
				SignDate:  0,
				StartDate: assetParam.StartDate,
				EndDate:   assetParam.EndDate,
				Status:    assetParam.Status,
			},
			wantErr: true,
		}, {
			name: "start_date",
			fields: fields{
				Owner:     assetParam.Owner,
				Previous:  assetParam.Previous,
				Asserts:   assetParam.Assets,
				SignDate:  assetParam.SignDate,
				StartDate: 0,
				EndDate:   assetParam.EndDate,
				Status:    assetParam.Status,
			},
			wantErr: true,
		}, {
			name: "end_date",
			fields: fields{
				Owner:     assetParam.Owner,
				Previous:  assetParam.Previous,
				Asserts:   assetParam.Assets,
				SignDate:  assetParam.SignDate,
				StartDate: assetParam.StartDate,
				EndDate:   0,
				Status:    assetParam.Status,
			},
			wantErr: true,
		}, {
			name: "asset",
			fields: fields{
				Owner:    assetParam.Owner,
				Previous: assetParam.Previous,
				Asserts: []*Asset{
					{
						Mcc:         0,
						Mnc:         0,
						TotalAmount: 0,
						SLAs:        nil,
					},
				},
				SignDate:  assetParam.SignDate,
				StartDate: assetParam.StartDate,
				EndDate:   assetParam.EndDate,
				Status:    assetParam.Status,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &AssetParam{
				Owner:     tt.fields.Owner,
				Previous:  tt.fields.Previous,
				Assets:    tt.fields.Asserts,
				SignDate:  tt.fields.SignDate,
				StartDate: tt.fields.StartDate,
				EndDate:   tt.fields.EndDate,
				Status:    tt.fields.Status,
			}
			err := z.Verify()
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
