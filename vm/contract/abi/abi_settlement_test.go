/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"bytes"
	"math"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
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
		EndData:   time.Now().AddDate(1, 0, 2).Unix(),
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
		SendingStatus: SendingStatusSend,
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

//
//func TestSettlement_IsContractAvailable(t *testing.T) {
//	teardownTestCase, l := setupTestCase(t)
//	defer teardownTestCase(t)
//
//	ctx := vmstore.NewVMContext(l)
//
//	var contracts []*contractContainer
//
//	for i := 0; i < 2; i++ {
//		param, a1, a2 := buildContractParam()
//		if i%2 == 1 {
//			param.SignatureB = &types.ZeroSignature
//		}
//		contracts = append(contracts, &contractContainer{
//			contract: param,
//			a1:       a1,
//			a2:       a2,
//		})
//		a, _ := param.Address()
//		abi, _ := param.ToABI()
//		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
//			t.Fatal(err)
//		}
//
//		if storage, err := ctx.GetStorage(types.SettlementAddress[:], a[:]); err == nil {
//			if !bytes.Equal(storage, abi) {
//				t.Fatalf("invalid saved contract, exp: %v, act: %v", abi, storage)
//			} else {
//				if p, err := ParseContractParam(storage); err == nil {
//					t.Log(a.String(), ": ", p.String())
//				} else {
//					t.Fatal(err)
//				}
//			}
//		}
//	}
//
//	if err := ctx.SaveStorage(); err != nil {
//		t.Fatal(err)
//	}
//
//	if contracts == nil || len(contracts) != 2 {
//		t.Fatal("invalid mock contract data")
//	}
//
//	a1, _ := contracts[0].contract.Address()
//	a2, _ := contracts[1].contract.Address()
//
//	t.Log(a1.String(), " >>> ", a2.String())
//
//	type args struct {
//		ctx  *vmstore.VMContext
//		addr *types.Address
//	}
//	tests := []struct {
//		name string
//		args args
//		want bool
//	}{
//		{
//			name: "ok",
//			args: args{
//				ctx:  ctx,
//				addr: &a1,
//			},
//			want: true,
//		},
//		{
//			name: "fail",
//			args: args{
//				ctx:  ctx,
//				addr: &a2,
//			},
//			want: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := IsContractAvailable(tt.args.ctx, tt.args.addr); got != tt.want {
//				t.Errorf("IsContractAvailable() of %s = %v, want %v", tt.args.addr.String(), got, tt.want)
//			}
//		})
//	}
//}

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

	a1 := mock.Address()
	a2 := mock.Address()
	s := CDRStatus{
		Params: []SettlementCDR{{
			CDRParam: param1,
			From:     a1,
		}, {
			CDRParam: param2,
			From:     a2,
		}},
		Status: SettlementStatusSuccess,
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
			if !reflect.DeepEqual(param1, s1.Params[0].CDRParam) {
				t.Fatalf("invalid csl data, exp: %s, act: %s", util.ToIndentString(param1), util.ToIndentString(s1.Params[0]))
			}
			if !reflect.DeepEqual(param2, s1.Params[1].CDRParam) {
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
				SendingStatus: SendingStatusSend,
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
				SendingStatus: SendingStatusSend,
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
				SendingStatus: SendingStatusSend,
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
				SendingStatus: SendingStatusSend,
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
				SendingStatus: SendingStatusSend,
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
		Params: []SettlementCDR{{
			CDRParam: cdrParam,
			From:     mock.Address(),
		}},
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
			if contract, err := GetContracts(ctx, &addr); err != nil {
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
	type fields struct {
		Params []SettlementCDR
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
				Params: []SettlementCDR{},
				Status: SettlementStatusStage1,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         time.Now().Unix(),
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSend,
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
				Params: []SettlementCDR{{
					CDRParam: CDRParam{
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
						SendingStatus: SendingStatusSend,
						DlrStatus:     DLRStatusDelivered,
					},
					From: mock.Address(),
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
						SendingStatus: SendingStatusSend,
						DlrStatus:     DLRStatusDelivered,
					},
					From: mock.Address(),
				},
				status: SettlementStatusSuccess,
			},
			wantErr: false,
		}, {
			name: "2 records",
			fields: fields{
				Params: []SettlementCDR{{
					CDRParam: CDRParam{
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
						SendingStatus: SendingStatusSend,
						DlrStatus:     DLRStatusUndelivered,
					},
					From: mock.Address(),
				}},
				Status: SettlementStatusFailure,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         time.Now().Unix(),
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSend,
						DlrStatus:     DLRStatusDelivered,
					},
					From: mock.Address(),
				},
				status: SettlementStatusFailure,
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
				EndData:   tt.fields.EndData,
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
		Params: []SettlementCDR{{
			CDRParam: cdrParam,
			From:     mock.Address(),
		}},
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
		Params: []SettlementCDR{{
			CDRParam: cdrParam,
			From:     mock.Address(),
		}},
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
		EndData   int64
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
				EndData:   cp.EndData,
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
				EndData:   tt.fields.EndData,
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
		EndData   int64
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
				EndData:   cp.EndData,
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
				EndData:   cp.EndData,
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
				EndData:   cp.EndData,
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
				EndData:   cp.EndData,
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
				EndData:   cp.EndData,
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
				EndData:   cp.EndData,
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
				EndData:   cp.EndData,
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
				EndData:   cp.EndData,
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
				EndData:   cp.EndData,
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
				EndData:   0,
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
				EndData:   tt.fields.EndData,
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

	if c, err := GetSettlementContract(ctx, &a1, &cdr); err != nil {
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
				SendingStatus: SendingStatusSend,
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
				SendingStatus: SendingStatusSend,
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
				SendingStatus: SendingStatusSend,
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
				SendingStatus: SendingStatusSend,
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
