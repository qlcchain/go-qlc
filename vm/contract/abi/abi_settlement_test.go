/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/util"

	"github.com/qlcchain/go-qlc/crypto/random"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/mock"
)

var (
	createContractParam = CreateContractParam{
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
		SignDate:  time.Now().AddDate(0, 0, -5).Unix(),
		StartDate: time.Now().AddDate(0, 0, -2).Unix(),
		EndDate:   time.Now().AddDate(1, 0, 2).Unix(),
	}

	cdrParam = CDRParam{
		Index:       1,
		SmsDt:       time.Now().Unix(),
		Sender:      "PCCWG",
		Destination: "85257***343",
		//DstCountry:    "Hong Kong",
		//DstOperator:   "HKTCSL",
		//DstMcc:        454,
		//DstMnc:        0,
		//SellPrice:     1,
		//SellCurrency:  "USD",
		//CustomerName:  "Tencent",
		//CustomerID:    "11667",
		SendingStatus: SendingStatusSent,
		DlrStatus:     DLRStatusDelivered,
	}
)

func buildContractParam() (param *ContractParam) {
	cp := createContractParam
	cp.PartyA.Address = mock.Address()
	cp.PartyB.Address = mock.Address()

	cd := time.Now().Unix()
	param = &ContractParam{
		CreateContractParam: cp,
		PreStops:            []string{"PCCWG", "test1"},
		NextStops:           []string{"HKTCSL", "test2"},
		ConfirmDate:         cd,
		Status:              ContractStatusActived,
	}
	return
}

func TestGetContractsByAddress(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	var contracts []*ContractParam

	for i := 0; i < 4; i++ {
		param := buildContractParam()
		contracts = append(contracts, param)
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	//if err := ctx.Iterator(types.SettlementAddress[:], func(key []byte, value []byte) error {
	//	t.Log(hex.EncodeToString(key), " >>> ", hex.EncodeToString(value))
	//	return nil
	//}); err != nil {
	//	t.Fatal(err)
	//}

	if contracts == nil || len(contracts) != 4 {
		t.Fatalf("invalid mock contract data, exp: 4, act: %d", len(contracts))
	}

	a := contracts[0].PartyA.Address

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "1st",
			args: args{
				ctx:  ctx,
				addr: &a,
			},
			want:    []*ContractParam{contracts[0]},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsByAddress(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContractsByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("GetContractsByAddress() len(go) != len(tt.want), %d,%d", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				a1, _ := got[i].Address()
				a2, _ := tt.want[i].Address()
				if a1 != a2 {
					t.Fatalf("GetContractsByAddress() i[%d] %v,%v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func mockContractData(size int) []*ContractParam {
	var contracts []*ContractParam
	accounts := []types.Address{mock.Address(), mock.Address()}

	for i := 0; i < size; i++ {
		cp := createContractParam

		var a1 types.Address
		var a2 types.Address
		if i%2 == 0 {
			a1 = accounts[0]
			a2 = accounts[1]
		} else {
			a1 = accounts[1]
			a2 = accounts[0]
		}
		cp.PartyA.Address = a1
		cp.PartyB.Address = a2

		for _, s := range cp.Services {
			s.Mcc = s.Mcc + uint64(i)
			s.Mnc = s.Mnc + uint64(i)
		}

		cd := time.Now().Unix()
		param := &ContractParam{
			CreateContractParam: cp,
			ConfirmDate:         cd,
		}
		contracts = append(contracts, param)
	}
	return contracts
}

func TestParseContractParam(t *testing.T) {
	param := buildContractParam()
	abi, err := param.ToABI()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		v []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *ContractParam
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				v: abi,
			},
			want:    param,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseContractParam(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContractParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseContractParam() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCDRStatus(t *testing.T) {
	param1 := cdrParam

	if data, err := param1.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		cp := &CDRParam{}
		if err := cp.FromABI(data); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(cp, &param1) {
				t.Fatal("invalid param")
			} else {
				t.Log(cp.String())
			}
		}
	}

	param2 := param1
	param2.Sender = "PCCWG"

	a1 := mock.Address().String()
	a2 := mock.Address().String()
	s := CDRStatus{
		Params: map[string][]CDRParam{a1: {param1}, a2: {param2}},
		Status: 0,
	}

	if data, err := s.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		s1 := &CDRStatus{}
		if err := s1.FromABI(data); err != nil {
			t.Fatal(err)
		} else {
			if len(s1.Params) != 2 {
				t.Fatalf("invalid param size, exp: 2, act:%d", len(s1.Params))
			}
			if !reflect.DeepEqual(param1, s1.Params[a1][0]) {
				t.Fatalf("invalid csl data, exp: %s, act: %s", util.ToIndentString(param1), util.ToIndentString(s1.Params[a1][0]))
			}
			if !reflect.DeepEqual(param2, s1.Params[a2][0]) {
				t.Fatal("invalid pccwg data")
			}
			t.Log(s1.String())
		}
	}
}

func TestGetAllSettlementContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	for i := 0; i < 4; i++ {
		param := buildContractParam()
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  *vmstore.VMContext
		size int
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx:  ctx,
				size: 4,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllSettlementContract(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllSettlementContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.args.size {
				t.Fatalf("invalid size: exp: %d, act: %d", tt.args.size, len(got))
			}
			for _, c := range got {
				t.Log(c.String())
			}
		})
	}
}

func TestCDRParam_Verify(t *testing.T) {
	type fields struct {
		Index       uint64
		SmsDt       int64
		Sender      string
		Destination string
		//DstCountry    string
		//DstOperator   string
		//DstMcc        uint64
		//DstMnc        uint64
		//SellPrice     float64
		//SellCurrency  string
		//CustomerName  string
		//CustomerID    string
		SendingStatus SendingStatus
		DlrStatus     DLRStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				Index:       1,
				SmsDt:       time.Now().Unix(),
				Sender:      "PCCWG",
				Destination: "85257***343",
				//DstCountry:    "Hong Kong",
				//DstOperator:   "HKTCSL",
				//DstMcc:        454,
				//DstMnc:        0,
				//SellPrice:     1,
				//SellCurrency:  "USD",
				//CustomerName:  "Tencent",
				//CustomerID:    "11667",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
			},
			wantErr: false,
		},
		{
			name: "f1",
			fields: fields{
				Index:       0,
				SmsDt:       time.Now().Unix(),
				Sender:      "PCCWG",
				Destination: "85257***343",
				//DstCountry:    "Hong Kong",
				//DstOperator:   "HKTCSL",
				//DstMcc:        454,
				//DstMnc:        0,
				//SellPrice:     1,
				//SellCurrency:  "USD",
				//CustomerName:  "Tencent",
				//CustomerID:    "11667",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
			},
			wantErr: true,
		},
		{
			name: "f2",
			fields: fields{
				Index:       1,
				SmsDt:       0,
				Sender:      "PCCWG",
				Destination: "85257***343",
				//DstCountry:    "Hong Kong",
				//DstOperator:   "HKTCSL",
				//DstMcc:        454,
				//DstMnc:        0,
				//SellPrice:     1,
				//SellCurrency:  "USD",
				//CustomerName:  "Tencent",
				//CustomerID:    "11667",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
			},
			wantErr: true,
		},
		{
			name: "f3",
			fields: fields{
				Index:       1,
				SmsDt:       time.Now().Unix(),
				Sender:      "",
				Destination: "85257***343",
				//DstCountry:    "Hong Kong",
				//DstOperator:   "HKTCSL",
				//DstMcc:        454,
				//DstMnc:        0,
				//SellPrice:     1,
				//SellCurrency:  "USD",
				//CustomerName:  "Tencent",
				//CustomerID:    "11667",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
			},
			wantErr: true,
		},
		{
			name: "f4",
			fields: fields{
				Index:       1,
				SmsDt:       time.Now().Unix(),
				Sender:      "PCCWG",
				Destination: "",
				//DstCountry:    "Hong Kong",
				//DstOperator:   "HKTCSL",
				//DstMcc:        454,
				//DstMnc:        0,
				//SellPrice:     1,
				//SellCurrency:  "USD",
				//CustomerName:  "Tencent",
				//CustomerID:    "11667",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
			},
			wantErr: true,
		},
		//{
		//	name: "f5",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		//DstCountry:    "",
		//		//DstOperator:   "HKTCSL",
		//		//DstMcc:        454,
		//		//DstMnc:        0,
		//		//SellPrice:     1,
		//		//SellCurrency:  "USD",
		//		//CustomerName:  "Tencent",
		//		//CustomerID:    "11667",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//}, {
		//	name: "f6",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		//DstCountry:    "Hong Kong",
		//		//DstOperator:   "",
		//		//DstMcc:        454,
		//		//DstMnc:        0,
		//		//SellPrice:     1,
		//		//SellCurrency:  "USD",
		//		//CustomerName:  "Tencent",
		//		//CustomerID:    "11667",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//},
		////{
		////	name: "f7",
		////	fields: fields{
		////		Index:         1,
		////		SmsDt:         time.Now().Unix(),
		////		Sender:        "PCCWG",
		////		Destination:   "85257***343",
		////		DstCountry:    "Hong Kong",
		////		DstOperator:   "HKTCSL",
		////		DstMcc:        0,
		////		DstMnc:        0,
		////		SellPrice:     1,
		////		SellCurrency:  "USD",
		////		CustomerName:  "Tencent",
		////		CustomerID:    "11667",
		////		SendingStatus: "Send",
		////		DlrStatus:     "Delivered",
		////	},
		////	wantErr: true,
		////}, {
		////	name: "f8",
		////	fields: fields{
		////		Index:         1,
		////		SmsDt:         time.Now().Unix(),
		////		Sender:        "PCCWG",
		////		Destination:   "85257***343",
		////		DstCountry:    "Hong Kong",
		////		DstOperator:   "HKTCSL",
		////		DstMcc:        454,
		////		DstMnc:        0,
		////		SellPrice:     1,
		////		SellCurrency:  "USD",
		////		CustomerName:  "Tencent",
		////		CustomerID:    "11667",
		////		SendingStatus: "Send",
		////		DlrStatus:     "Delivered",
		////	},
		////	wantErr: true,
		////},
		//{
		//	name: "f9",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		DstCountry:    "Hong Kong",
		//		DstOperator:   "HKTCSL",
		//		DstMcc:        454,
		//		DstMnc:        0,
		//		SellPrice:     0,
		//		SellCurrency:  "USD",
		//		CustomerName:  "Tencent",
		//		CustomerID:    "11667",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//}, {
		//	name: "f10",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		DstCountry:    "Hong Kong",
		//		DstOperator:   "HKTCSL",
		//		DstMcc:        454,
		//		DstMnc:        0,
		//		SellPrice:     1,
		//		SellCurrency:  "",
		//		CustomerName:  "Tencent",
		//		CustomerID:    "11667",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//}, {
		//	name: "f11",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		DstCountry:    "Hong Kong",
		//		DstOperator:   "HKTCSL",
		//		DstMcc:        454,
		//		DstMnc:        0,
		//		SellPrice:     1,
		//		SellCurrency:  "USD",
		//		CustomerName:  "",
		//		CustomerID:    "11667",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//}, {
		//	name: "f12",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		DstCountry:    "Hong Kong",
		//		DstOperator:   "HKTCSL",
		//		DstMcc:        454,
		//		DstMnc:        0,
		//		SellPrice:     1,
		//		SellCurrency:  "USD",
		//		CustomerName:  "Tencent",
		//		CustomerID:    "",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRParam{
				Index:       tt.fields.Index,
				SmsDt:       tt.fields.SmsDt,
				Sender:      tt.fields.Sender,
				Destination: tt.fields.Destination,
				//DstCountry:    tt.fields.DstCountry,
				//DstOperator:   tt.fields.DstOperator,
				//DstMcc:        tt.fields.DstMcc,
				//DstMnc:        tt.fields.DstMnc,
				//SellPrice:     tt.fields.SellPrice,
				//SellCurrency:  tt.fields.SellCurrency,
				//CustomerName:  tt.fields.CustomerName,
				//CustomerID:    tt.fields.CustomerID,
				SendingStatus: tt.fields.SendingStatus,
				DlrStatus:     tt.fields.DlrStatus,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseCDRStatus(t *testing.T) {
	status := &CDRStatus{
		Params: map[string][]CDRParam{mock.Address().String(): {cdrParam}},
		Status: SettlementStatusStage1,
	}

	if abi, err := status.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		if s2, err := ParseCDRStatus(abi); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(status, s2) {
				t.Fatalf("invalid cdr status %v, %v", status, s2)
			}
		}
	}
}

func TestCDRParam_ToHash(t *testing.T) {
	if h, err := cdrParam.ToHash(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(h)
	}
}

func TestGetContracts(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	ctx := vmstore.NewVMContext(l)
	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		} else {
			//t.Log(hex.EncodeToString(abi))
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	for _, d := range data {
		if addr, err := d.Address(); err != nil {
			t.Fatal(err)
		} else {
			if contract, err := GetSettlementContract(ctx, &addr); err != nil {
				t.Fatal(err)
			} else {
				if !reflect.DeepEqual(contract, d) {
					t.Fatalf("invalid %v, %v", contract, d)
				}
			}
		}
	}
}

func TestCDRStatus_DoSettlement(t *testing.T) {
	addr1 := mock.Address()
	smsDt := time.Now().Unix()

	type fields struct {
		Params map[string][]CDRParam
		Status SettlementStatus
	}

	type args struct {
		cdr    SettlementCDR
		status SettlementStatus
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "1 records",
			fields: fields{
				Params: nil,
				Status: SettlementStatusStage1,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         time.Now().Unix(),
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusDelivered,
					},
					From: mock.Address(),
				},
				status: SettlementStatusStage1,
			},
			wantErr: false,
		},
		{
			name: "2 records",
			fields: fields{
				Params: map[string][]CDRParam{mock.Address().String(): {
					{
						Index:       1,
						SmsDt:       time.Now().Unix(),
						Sender:      "PCCWG",
						Destination: "85257***343",
						//DstCountry:    "Hong Kong",
						//DstOperator:   "HKTCSL",
						//DstMcc:        454,
						//DstMnc:        0,
						//SellPrice:     1,
						//SellCurrency:  "USD",
						//CustomerName:  "Tencent",
						//CustomerID:    "11668",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusDelivered,
					},
				}},
				Status: 0,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         time.Now().Unix(),
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusDelivered,
					},
					From: mock.Address(),
				},
				status: SettlementStatusSuccess,
			},
			wantErr: false,
		}, {
			name: "2 records failed",
			fields: fields{
				Params: map[string][]CDRParam{mock.Address().String(): {
					{
						Index:       1,
						SmsDt:       time.Now().Unix(),
						Sender:      "PCCWG",
						Destination: "85257***343",
						//DstCountry:    "Hong Kong",
						//DstOperator:   "HKTCSL",
						//DstMcc:        454,
						//DstMnc:        0,
						//SellPrice:     1,
						//SellCurrency:  "USD",
						//CustomerName:  "Tencent",
						//CustomerID:    "11668",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusRejected,
					},
				}},
				Status: 0,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         time.Now().Unix(),
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusRejected,
					},
					From: mock.Address(),
				},
				status: SettlementStatusFailure,
			},
			wantErr: false,
		}, {
			name: "2 duplicate records of one account",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {
					{
						Index:       1,
						SmsDt:       smsDt,
						Sender:      "PCCWG",
						Destination: "85257***343",
						//DstCountry:    "Hong Kong",
						//DstOperator:   "HKTCSL",
						//DstMcc:        454,
						//DstMnc:        0,
						//SellPrice:     1,
						//SellCurrency:  "USD",
						//CustomerName:  "Tencent",
						//CustomerID:    "11668",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusUndelivered,
					},
				}},
				Status: SettlementStatusFailure,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         smsDt,
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusDelivered,
					},
					From: addr1,
				},
				status: SettlementStatusStage1,
			},
			wantErr: false,
		},
		{
			name: "2 duplicate records",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {
					{
						Index:       1,
						SmsDt:       smsDt,
						Sender:      "PCCWG",
						Destination: "85257***343",
						//DstCountry:    "Hong Kong",
						//DstOperator:   "HKTCSL",
						//DstMcc:        454,
						//DstMnc:        0,
						//SellPrice:     1,
						//SellCurrency:  "USD",
						//CustomerName:  "Tencent",
						//CustomerID:    "11668",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusUndelivered,
					},
				}, mock.Address().String(): {
					{
						Index:       1,
						SmsDt:       smsDt,
						Sender:      "PCCWG",
						Destination: "85257***343",
						//DstCountry:    "Hong Kong",
						//DstOperator:   "HKTCSL",
						//DstMcc:        454,
						//DstMnc:        0,
						//SellPrice:     1,
						//SellCurrency:  "USD",
						//CustomerName:  "Tencent",
						//CustomerID:    "11668",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusUndelivered,
					},
				},
				},
				Status: SettlementStatusFailure,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         smsDt,
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusDelivered,
					},
					From: addr1,
				},
				status: SettlementStatusDuplicate,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRStatus{
				Params: tt.fields.Params,
				Status: tt.fields.Status,
			}
			if err := z.DoSettlement(tt.args.cdr); (err != nil) != tt.wantErr {
				t.Errorf("DoSettlement() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if z.Status != tt.args.status {
					t.Fatalf("invalid status, exp: %s, act: %s", tt.args.status.String(), z.Status.String())
				}
			}
		})
	}
}

func TestContractParam_FromABI(t *testing.T) {
	cp := buildContractParam()

	abi, err := cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              cp.Status,
			},
			args: args{
				data: abi,
			},
			wantErr: false,
		}, {
			name: "failed",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              cp.Status,
			},
			args: args{
				data: abi[:10],
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				PreStops:            tt.fields.PreStops,
				NextStops:           tt.fields.NextStops,
				ConfirmDate:         tt.fields.ConfirmDate,
				Status:              tt.fields.Status,
			}
			if err := z.FromABI(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("FromABI() error = %v, wantErr %v", err, tt.wantErr)
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
			want:    types.Balance{Int: new(big.Int).Mul(new(big.Int).SetUint64(100), new(big.Int).SetUint64(0.04*1e8))},
			wantErr: false,
		}, {
			name: "overflow",
			fields: fields{
				ServiceId:   mock.Hash().String(),
				Mcc:         22,
				Mnc:         1,
				TotalAmount: math.MaxUint64,
				UnitPrice:   0.04,
				Currency:    "USD",
			},
			want:    types.ZeroBalance,
			wantErr: true,
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
			want:    types.Balance{Int: new(big.Int).Mul(big.NewInt(100), big.NewInt(2*1e8))}.Add(types.Balance{Int: new(big.Int).Mul(big.NewInt(300), big.NewInt(4*1e8))}),
			wantErr: false,
		},
		{
			name: "overflow",
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
					TotalAmount: math.MaxUint64,
					UnitPrice:   4,
					Currency:    "USD",
				}},
				SignDate:  time.Now().Unix(),
				StartDate: time.Now().AddDate(0, 0, 2).Unix(),
				EndData:   time.Now().AddDate(1, 0, 2).Unix(),
			},
			want:    types.ZeroBalance,
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

func TestCDRParam_FromABI(t *testing.T) {
	cdr := cdrParam
	abi, err := cdr.ToABI()
	if err != nil {
		t.Fatal(err)
	}
	type fields struct {
		Index         uint64
		SmsDt         int64
		Sender        string
		Destination   string
		DstCountry    string
		DstOperator   string
		DstMcc        uint64
		DstMnc        uint64
		SellPrice     float64
		SellCurrency  string
		CustomerName  string
		CustomerID    string
		SendingStatus SendingStatus
		DlrStatus     DLRStatus
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				Index:       cdr.Index,
				SmsDt:       cdr.SmsDt,
				Sender:      cdr.Sender,
				Destination: cdr.Destination,
				//DstCountry:    cdr.DstCountry,
				//DstOperator:   cdr.DstOperator,
				//DstMcc:        cdr.DstMcc,
				//DstMnc:        cdr.DstMnc,
				//SellPrice:     cdr.SellPrice,
				//SellCurrency:  cdr.SellCurrency,
				//CustomerName:  cdr.CustomerName,
				//CustomerID:    cdr.CustomerID,
				SendingStatus: cdr.SendingStatus,
				DlrStatus:     cdr.DlrStatus,
			},
			args: args{
				data: abi,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRParam{
				Index:       tt.fields.Index,
				SmsDt:       tt.fields.SmsDt,
				Sender:      tt.fields.Sender,
				Destination: tt.fields.Destination,
				//DstCountry:    tt.fields.DstCountry,
				//DstOperator:   tt.fields.DstOperator,
				//DstMcc:        tt.fields.DstMcc,
				//DstMnc:        tt.fields.DstMnc,
				//SellPrice:     tt.fields.SellPrice,
				//SellCurrency:  tt.fields.SellCurrency,
				//CustomerName:  tt.fields.CustomerName,
				//CustomerID:    tt.fields.CustomerID,
				SendingStatus: tt.fields.SendingStatus,
				DlrStatus:     tt.fields.DlrStatus,
			}
			if err := z.FromABI(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("FromABI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCDRParam_String(t *testing.T) {
	s := cdrParam.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}
}

func TestCDRStatus_FromABI(t *testing.T) {
	status := CDRStatus{
		Params: map[string][]CDRParam{
			mock.Address().String(): {cdrParam},
		},
		Status: SettlementStatusSuccess,
	}

	abi, err := status.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	s2 := &CDRStatus{}
	if err = s2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&status, s2) {
			t.Fatalf("invalid cdr status data, %v, %v", &status, s2)
		}
	}
}

func TestCDRStatus_String(t *testing.T) {
	status := CDRStatus{
		Params: map[string][]CDRParam{
			mock.Address().String(): {cdrParam},
		},
		Status: SettlementStatusSuccess,
	}
	s := status.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}

}

func TestContractParam_Equal(t *testing.T) {
	param := buildContractParam()
	cp := param.CreateContractParam

	type fields struct {
		CreateContractParam CreateContractParam
		ConfirmDate         int64
		SignatureB          *types.Signature
	}
	type args struct {
		cp *CreateContractParam
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "equal",
			fields: fields{
				CreateContractParam: cp,
				ConfirmDate:         param.ConfirmDate,
			},
			args: args{
				cp: &cp,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "nil",
			fields: fields{
				CreateContractParam: cp,
				ConfirmDate:         param.ConfirmDate,
			},
			args: args{
				cp: nil,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				ConfirmDate:         tt.fields.ConfirmDate,
			}
			got, err := z.Equal(tt.args.cp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Equal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Equal() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContractParam_String(t *testing.T) {
	param := buildContractParam()
	s := param.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}
}

func TestContractParam_ToABI(t *testing.T) {
	param := buildContractParam()
	abi, err := param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	p2 := &ContractParam{}
	if err = p2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		if param.String() != p2.String() {
			t.Fatalf("invalid contract param data, %s, %s", param.String(), p2.String())
		}
	}
}

func TestContractService_ToABI(t *testing.T) {
	param := ContractService{
		ServiceId:   mock.Hash().String(),
		Mcc:         1,
		Mnc:         2,
		TotalAmount: 100,
		UnitPrice:   2,
		Currency:    "USD",
	}
	abi, err := param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	p2 := &ContractService{}
	if err = p2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&param, p2) {
			t.Fatalf("invalid contract service data, %v, %v", &param, p2)
		}
	}
}

func TestContractor_FromABI(t *testing.T) {
	c := createContractParam.PartyA
	abi, err := c.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	c2 := &Contractor{}
	if err = c2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&c, c2) {
			t.Fatalf("invalid contractor, %v, %v", &c, c2)
		}
	}
}

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

func TestGetContractsIDByAddressAsPartyA(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)
	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	a1 := data[0].PartyA.Address

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  ctx,
				addr: &a1,
			},
			want:    []*ContractParam{data[0]},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsIDByAddressAsPartyA(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContractsIDByAddressAsPartyA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("GetContractsIDByAddressAsPartyA() len(go) != len(tt.want), %d,%d", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				a1, _ := got[i].Address()
				a2, _ := tt.want[i].Address()
				if a1 != a2 {
					t.Fatalf("GetContractsIDByAddressAsPartyA() i[%d] %v,%v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetContractsIDByAddressAsPartyB(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)
	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	a2 := data[0].PartyB.Address

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  ctx,
				addr: &a2,
			},
			want:    []*ContractParam{data[0]},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsIDByAddressAsPartyB(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetContractsIDByAddressAsPartyB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("TestGetContractsIDByAddressAsPartyB() len(go) != len(tt.want), %d,%d", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				a1, _ := got[i].Address()
				a2, _ := tt.want[i].Address()
				if a1 != a2 {
					t.Fatalf("TestGetContractsIDByAddressAsPartyB() i[%d] %v,%v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIsContractAvailable(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)

	var contracts []*ContractParam

	for i := 0; i < 2; i++ {
		param := buildContractParam()

		if i%2 == 1 {
			param.Status = ContractStatusActiveStage1
		}

		contracts = append(contracts, param)
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}

		if storage, err := ctx.GetStorage(types.SettlementAddress[:], a[:]); err == nil {
			if !bytes.Equal(storage, abi) {
				t.Fatalf("invalid saved contract, exp: %v, act: %v", abi, storage)
			} else {
				if p, err := ParseContractParam(storage); err == nil {
					t.Log(a.String(), ": ", p.String())
				} else {
					t.Fatal(err)
				}
			}
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	if contracts == nil || len(contracts) != 2 {
		t.Fatal("invalid mock contract data")
	}

	a1, _ := contracts[0].Address()
	a2, _ := contracts[1].Address()

	t.Log(a1.String(), " >>> ", a2.String())

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  ctx,
				addr: &a1,
			},
			want: true,
		},
		{
			name: "fail",
			args: args{
				ctx:  ctx,
				addr: &a2,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContractAvailable(tt.args.ctx, tt.args.addr); got != tt.want {
				t.Errorf("IsContractAvailable() of %s = %v, want %v", tt.args.addr.String(), got, tt.want)
			}
		})
	}
}

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

func TestGetSettlementContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)

	var contracts []*ContractParam

	for i := 0; i < 2; i++ {
		param := buildContractParam()

		if i%2 == 1 {
			param.Status = ContractStatusActiveStage1
		}

		contracts = append(contracts, param)
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}

		//if storage, err := ctx.GetStorage(types.SettlementAddress[:], a[:]); err == nil {
		//	if !bytes.Equal(storage, abi) {
		//		t.Fatalf("invalid saved contract, exp: %v, act: %v", abi, storage)
		//	} else {
		//		if p, err := ParseContractParam(storage); err == nil {
		//			t.Log(a.String(), ": ", p.String())
		//		} else {
		//			t.Fatal(err)
		//		}
		//	}
		//}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	if contracts == nil || len(contracts) != 2 {
		t.Fatal("invalid mock contract data")
	}
	a1 := contracts[0].PartyA.Address
	cdr := cdrParam
	cdr.NextStop = "HKTCSL"

	if c, err := FindSettlementContract(ctx, &a1, &cdr); err != nil {
		t.Fatal(err)
	} else {
		t.Log(c)
	}
}

func TestStopParam_Verify(t *testing.T) {
	type fields struct {
		StopName string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				StopName: "test1",
			},
			wantErr: false,
		}, {
			name: "false",
			fields: fields{
				StopName: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &StopParam{
				StopName: tt.fields.StopName,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateStopParam_Verify(t *testing.T) {
	type fields struct {
		StopName string
		NewName  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				StopName: "test1",
				NewName:  "test2",
			},
			wantErr: false,
		}, {
			name: "false",
			fields: fields{
				StopName: "",
				NewName:  "222",
			},
			wantErr: true,
		}, {
			name: "false2",
			fields: fields{
				StopName: "111",
				NewName:  "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &UpdateStopParam{
				StopName: tt.fields.StopName,
				New:      tt.fields.NewName,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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

func TestContractParam_IsPreStop(t *testing.T) {
	param := buildContractParam()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
	}
	type args struct {
		n string
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
				CreateContractParam: param.CreateContractParam,
				PreStops:            []string{"PCCWG", "111"},
				NextStops:           nil,
				ConfirmDate:         param.ConfirmDate,
				Status:              param.Status,
			},
			args: args{
				n: "PCCWG",
			},
			want: true,
		}, {
			name: "ok",
			fields: fields{
				CreateContractParam: param.CreateContractParam,
				PreStops:            []string{"PCCWG", "111"},
				NextStops:           nil,
				ConfirmDate:         param.ConfirmDate,
				Status:              param.Status,
			},
			args: args{
				n: "PCCWG1",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				PreStops:            tt.fields.PreStops,
				NextStops:           tt.fields.NextStops,
				ConfirmDate:         tt.fields.ConfirmDate,
				Status:              tt.fields.Status,
			}
			if got := z.IsPreStop(tt.args.n); got != tt.want {
				t.Errorf("IsPreStop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCDRParam_Status(t *testing.T) {
	type fields struct {
		Index         uint64
		SmsDt         int64
		Sender        string
		Destination   string
		SendingStatus SendingStatus
		DlrStatus     DLRStatus
		PreStop       string
		NextStop      string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "ok",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "",
				Destination:   "",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
				PreStop:       "",
				NextStop:      "",
			},
			want: true,
		}, {
			name: "f1",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "",
				Destination:   "",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusUnknown,
				PreStop:       "",
				NextStop:      "",
			},
			want: true,
		}, {
			name: "f2",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "",
				Destination:   "",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusUndelivered,
				PreStop:       "",
				NextStop:      "",
			},
			want: false,
		}, {
			name: "f4",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "",
				Destination:   "",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusEmpty,
				PreStop:       "",
				NextStop:      "",
			},
			want: true,
		}, {
			name: "f5",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "",
				Destination:   "",
				SendingStatus: SendingStatusError,
				DlrStatus:     DLRStatusEmpty,
				PreStop:       "",
				NextStop:      "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRParam{
				Index:         tt.fields.Index,
				SmsDt:         tt.fields.SmsDt,
				Sender:        tt.fields.Sender,
				Destination:   tt.fields.Destination,
				SendingStatus: tt.fields.SendingStatus,
				DlrStatus:     tt.fields.DlrStatus,
				PreStop:       tt.fields.PreStop,
				NextStop:      tt.fields.NextStop,
			}
			if got := z.Status(); got != tt.want {
				t.Errorf("Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStopParam_ToABI(t *testing.T) {
	param := StopParam{
		StopName: "test",
	}
	abi, err := param.ToABI(MethodNameAddNextStop)
	if err != nil {
		t.Fatal(err)
	}

	p2 := &StopParam{}
	if err = p2.FromABI(MethodNameAddNextStop, abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&param, p2) {
			t.Fatalf("invalid param, %v, %v", &param, p2)
		} else {
			if err := p2.Verify(); err != nil {
				t.Fatalf("verify failed, %s", err)
			}
		}
	}
}

func TestUpdateStopParam_ToABI(t *testing.T) {
	param := UpdateStopParam{
		StopName: "test",
		New:      "hahah",
	}
	abi, err := param.ToABI(MethodNameUpdateNextStop)
	if err != nil {
		t.Fatal(err)
	}

	p2 := &UpdateStopParam{}
	if err = p2.FromABI(MethodNameUpdateNextStop, abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&param, p2) {
			t.Fatalf("invalid param, %v, %v", &param, p2)
		} else {
			if err := p2.Verify(); err != nil {
				t.Fatalf("verify failed, %s", err)
			}
		}
	}
}

func TestTerminateParam_ToABI(t *testing.T) {
	param := &TerminateParam{ContractAddress: mock.Address()}
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

func TestContractParam_IsContractor(t *testing.T) {
	cp := buildContractParam()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
		Terminator          *types.Address
	}
	type args struct {
		addr types.Address
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "partyA",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              cp.Status,
			},
			args: args{
				addr: cp.PartyA.Address,
			},
			want: true,
		}, {
			name: "partyB",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              cp.Status,
			},
			args: args{
				addr: cp.PartyB.Address,
			},
			want: true,
		}, {
			name: "failed",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              cp.Status,
			},
			args: args{
				addr: mock.Address(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				PreStops:            tt.fields.PreStops,
				NextStops:           tt.fields.NextStops,
				ConfirmDate:         tt.fields.ConfirmDate,
				Status:              tt.fields.Status,
				Terminator:          tt.fields.Terminator,
			}
			if got := z.IsContractor(tt.args.addr); got != tt.want {
				t.Errorf("IsContractor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContractParam_DoActive(t *testing.T) {
	cp := buildContractParam()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
		Terminator          *types.Address
	}
	type args struct {
		operator types.Address
		status   ContractStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActiveStage1,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusActived,
			},
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActiveStage1,
			},
			args: args{
				operator: cp.PartyA.Address,
				status:   ContractStatusActiveStage1,
			},
			wantErr: true,
		},
		{
			name: "f2",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActived,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusActiveStage1,
			},
			wantErr: true,
		}, {
			name: "f3",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyed,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusActiveStage1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				PreStops:            tt.fields.PreStops,
				NextStops:           tt.fields.NextStops,
				ConfirmDate:         tt.fields.ConfirmDate,
				Status:              tt.fields.Status,
				Terminator:          tt.fields.Terminator,
			}
			if err := z.DoActive(tt.args.operator); (err != nil) != tt.wantErr {
				t.Errorf("DoActive() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if err == nil && tt.args.status != z.Status {
					t.Errorf("DoActive() status = %s, want = %s", z.Status.String(), tt.args.status.String())
				}
			}
		})
	}
}

func TestContractParam_DoTerminate(t *testing.T) {
	cp := buildContractParam()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
		Terminator          *types.Address
	}
	type args struct {
		operator types.Address
		status   ContractStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "partyA_destroy_active_stage1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActiveStage1,
			},
			args: args{
				operator: cp.PartyA.Address,
				status:   ContractStatusDestroyed,
			},
			wantErr: false,
		}, {
			name: "partyB_destroy_active_stage1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActiveStage1,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusRejected,
			},
			wantErr: false,
		}, {
			name: "partyA_destroy_stage1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActived,
			},
			args: args{
				operator: cp.PartyA.Address,
				status:   ContractStatusDestroyStage1,
			},
			wantErr: false,
		}, {
			name: "partyB_destroy",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyStage1,
				Terminator:          &cp.PartyA.Address,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusDestroyed,
			},
			wantErr: false,
		},
		{
			name: "f1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyStage1,
				Terminator:          &cp.PartyA.Address,
			},
			args: args{
				operator: cp.PartyA.Address,
				status:   ContractStatusDestroyed,
			},
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyStage1,
				Terminator:          &cp.PartyA.Address,
			},
			args: args{
				operator: mock.Address(),
				status:   ContractStatusDestroyed,
			},
			wantErr: true,
		}, {
			name: "f3",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyed,
				Terminator:          &cp.PartyA.Address,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusDestroyed,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				PreStops:            tt.fields.PreStops,
				NextStops:           tt.fields.NextStops,
				ConfirmDate:         tt.fields.ConfirmDate,
				Status:              tt.fields.Status,
				Terminator:          tt.fields.Terminator,
			}
			if err := z.DoTerminate(tt.args.operator); (err != nil) != tt.wantErr {
				t.Errorf("DoTerminate() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if err == nil && tt.args.status != z.Status {
					t.Errorf("DoTerminate() status = %s, want = %s", z.Status.String(), tt.args.status.String())
				}
			}
		})
	}
}

func buildCDRStatus() *CDRStatus {
	cdr1 := cdrParam
	i, _ := random.Intn(10000)
	cdr1.Index = uint64(i)

	status := &CDRStatus{
		Params: map[string][]CDRParam{
			mock.Address().String(): {cdr1},
		},
		Status: SettlementStatusSuccess,
	}

	return status
}

func TestGetAllCDRStatus(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	contractAddr := mock.Address()

	var data []*CDRStatus
	for i := 0; i < 4; i++ {
		s := buildCDRStatus()
		if h, err := s.ToHash(); err != nil {
			t.Fatal(err)
		} else {
			if h.IsZero() {
				t.Fatal("invalid hash")
			}
			if abi, err := s.ToABI(); err != nil {
				t.Fatal(err)
			} else {
				if err := ctx.SetStorage(contractAddr[:], h[:], abi); err != nil {
					t.Fatal(err)
				} else {
					data = append(data, s)
				}
			}
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
		size int
	}
	tests := []struct {
		name    string
		args    args
		want    []*CDRStatus
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  ctx,
				addr: &contractAddr,
				size: len(data),
			},
			want:    data,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllCDRStatus(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllCDRStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(data) {
				t.Errorf("GetAllCDRStatus() got = %d, want %d", len(got), len(data))
			}
			//for i, s := range tt.want {
			//	for k, v := range s.Params {
			//		g := got[i].Params[k]
			//		if !reflect.DeepEqual(g, v) {
			//			t.Errorf("GetAllCDRStatus() got = %v, want %v", g, s)
			//		}
			//	}
			//}
		})
	}
}

func TestGetCDRStatus(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	contractAddr := mock.Address()
	contractAddr2 := mock.Address()

	s := buildCDRStatus()
	h, err := s.ToHash()
	if err != nil {
		t.Fatal(err)
	} else {
		if abi, err := s.ToABI(); err != nil {
			t.Fatal(err)
		} else {
			if err := ctx.SetStorage(contractAddr[:], h[:], abi); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
		hash types.Hash
	}
	tests := []struct {
		name    string
		args    args
		want    *CDRStatus
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  ctx,
				addr: &contractAddr,
				hash: h,
			},
			want:    s,
			wantErr: false,
		}, {
			name: "fail",
			args: args{
				ctx:  ctx,
				addr: &contractAddr,
				hash: mock.Hash(),
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "f2",
			args: args{
				ctx:  ctx,
				addr: &contractAddr2,
				hash: h,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCDRStatus(tt.args.ctx, tt.args.addr, tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCDRStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCDRStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCDRStatus_IsInCycle(t *testing.T) {
	param := cdrParam
	param.SmsDt = 1582001974

	type fields struct {
		Params map[string][]CDRParam
		Status SettlementStatus
	}
	type args struct {
		start int64
		end   int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "ok1",
			fields: fields{
				Params: map[string][]CDRParam{
					mock.Address().String(): {cdrParam},
				},
				Status: 0,
			},
			args: args{
				start: 0,
				end:   100,
			},
			want: true,
		}, {
			name: "ok2",
			fields: fields{
				Params: map[string][]CDRParam{
					mock.Address().String(): {param},
				},
				Status: 0,
			},
			args: args{
				start: 1581829174,
				end:   1582433974,
			},
			want: true,
		}, {
			name: "ok3",
			fields: fields{
				Params: map[string][]CDRParam{
					mock.Address().String(): {param, param},
				},
				Status: 0,
			},
			args: args{
				start: 1581829174,
				end:   1582433974,
			},
			want: true,
		}, {
			name: "f1",
			fields: fields{
				Params: map[string][]CDRParam{
					mock.Address().String(): {param},
				},
				Status: 0,
			},
			args: args{
				start: 1582433974,
				end:   1582434974,
			},
			want: false,
		}, {
			name: "f2",
			fields: fields{
				Params: map[string][]CDRParam{
					mock.Address().String(): {cdrParam, cdrParam},
				},
				Status: 0,
			},
			args: args{
				start: 1582433974,
				end:   1582434974,
			},
			want: false,
		}, {
			name: "f3",
			fields: fields{
				Params: nil,
				Status: 0,
			},
			args: args{
				start: 1582433974,
				end:   1582434974,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRStatus{
				Params: tt.fields.Params,
				Status: tt.fields.Status,
			}
			if got := z.IsInCycle(tt.args.start, tt.args.end); got != tt.want {
				t.Errorf("IsInCycle() = %v, want %v", got, tt.want)
			}
		})
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

func TestSummaryRecord_DoCalculate(t *testing.T) {
	r := &SummaryRecord{
		Total:   0,
		Success: 10,
		Fail:    2,
		Result:  0,
	}
	r.DoCalculate()
	if r.Total != 12 {
		t.Fail()
	}
	t.Log(r.String())
}

func TestNewSummaryResult(t *testing.T) {
	r := NewSummaryResult()

	for i := 0; i < 10; i++ {
		r.IncreasePartyASuccess()
		r.IncreaseSuccess("WeChat", true)
		r.IncreaseSuccess("Slack", false)
	}
	for i := 0; i < 2; i++ {
		r.IncreasePartyAFail()
		r.IncreaseFail("WeChat", true)
		r.IncreaseFail("Slack", true)
	}

	for i := 0; i < 20; i++ {
		r.IncreasePartyBSuccess()
		r.IncreaseSuccess("WeChat", false)
		r.IncreaseSuccess("Slack", true)
	}

	for i := 0; i < 3; i++ {
		r.IncreasePartyBFail()
		r.IncreaseFail("WeChat", false)
		r.IncreaseFail("Slack", true)
	}

	r.DoCalculate()
	t.Log(r.String())
}

func Test1(t *testing.T) {
	s := []int{3, 6, 1, 9}
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})

	fmt.Println(s)
}
