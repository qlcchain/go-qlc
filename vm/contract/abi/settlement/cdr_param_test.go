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
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/mock"
)

func TestCDRParam_Verify(t *testing.T) {
	type fields struct {
		Index         uint64
		SmsDt         int64
		Sender        string
		Customer      string
		Destination   string
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
		}, {
			name: "ok2",
			fields: fields{
				Index:       1,
				SmsDt:       time.Now().Unix(),
				Sender:      "PCCWG",
				Customer:    "hahah",
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

func TestCDRParam_ToHash(t *testing.T) {
	if h, err := cdrParam.ToHash(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(h)
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

func TestCDRParam_String(t *testing.T) {
	s := cdrParam.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}
}

func TestCDRParam_GetCustomer(t *testing.T) {
	type fields struct {
		Index         uint64
		SmsDt         int64
		Sender        string
		Customer      string
		Destination   string
		SendingStatus SendingStatus
		DlrStatus     DLRStatus
		PreStop       string
		NextStop      string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ok",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "WeChat",
				Customer:      "Tencent",
				Destination:   "",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
				PreStop:       "",
				NextStop:      "",
			},
			want: "Tencent",
		}, {
			name: "ok",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "WeChat",
				Customer:      "",
				Destination:   "",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
				PreStop:       "",
				NextStop:      "",
			},
			want: "WeChat",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRParam{
				Index:         tt.fields.Index,
				SmsDt:         tt.fields.SmsDt,
				Sender:        tt.fields.Sender,
				Customer:      tt.fields.Customer,
				Destination:   tt.fields.Destination,
				SendingStatus: tt.fields.SendingStatus,
				DlrStatus:     tt.fields.DlrStatus,
				PreStop:       tt.fields.PreStop,
				NextStop:      tt.fields.NextStop,
			}
			if got := z.GetCustomer(); got != tt.want {
				t.Errorf("GetCustomer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCDRStatus_ToHash(t *testing.T) {
	type fields struct {
		Params map[string][]CDRParam
		Status SettlementStatus
	}
	tests := []struct {
		name    string
		fields  fields
		want    types.Hash
		wantErr bool
	}{
		{
			name: "f",
			fields: fields{
				Params: nil,
				Status: SettlementStatusSuccess,
			},
			want:    types.ZeroHash,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRStatus{
				Params: tt.fields.Params,
				Status: tt.fields.Status,
			}
			got, err := z.ToHash()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCDRStatus_ExtractAccount(t *testing.T) {
	addr1 := mock.Address()
	addr2 := mock.Address()

	cdr1 := cdrParam
	cdr1.Sender = "WeChat"
	cdr2 := cdrParam
	cdr2.Sender = "WeChat"
	cdr2.Customer = "Tencent"
	cdr2.Account = "Tencent_DIRECTS"

	type fields struct {
		Params map[string][]CDRParam
		Status SettlementStatus
	}
	tests := []struct {
		name            string
		fields          fields
		wantDt          int64
		wantAccount     string
		wantDestination string
		wantErr         bool
	}{
		{
			name: "ok",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdr2}, addr2.String(): {cdr1}},
				Status: SettlementStatusSuccess,
			},
			wantDt:          cdr1.SmsDt,
			wantAccount:     "Tencent_DIRECTS",
			wantDestination: cdr1.Destination,
			wantErr:         false,
		}, {
			name: "f1",
			fields: fields{
				Params: map[string][]CDRParam{},
				Status: SettlementStatusSuccess,
			},
			wantDt:          0,
			wantAccount:     "",
			wantDestination: "",
			wantErr:         true,
		}, {
			name: "f2",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdr1}},
				Status: SettlementStatusSuccess,
			},
			wantDt:          cdr1.SmsDt,
			wantAccount:     "",
			wantDestination: cdr1.Destination,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRStatus{
				Params: tt.fields.Params,
				Status: tt.fields.Status,
			}
			gotDt, gotAccount, gotDestination, err := z.ExtractAccount()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDt != tt.wantDt {
				t.Errorf("ExtractAccount() gotDt = %v, want %v", gotDt, tt.wantDt)
			}
			if gotAccount != tt.wantAccount {
				t.Errorf("ExtractAccount() gotAccount = %v, want %v", gotAccount, tt.wantAccount)
			}
			if gotDestination != tt.wantDestination {
				t.Errorf("ExtractAccount() gotDestination = %v, want %v", gotDestination, tt.wantDestination)
			}
		})
	}
}

func TestCDRParamList_ToABI(t *testing.T) {
	p1 := cdrParam
	p2 := cdrParam
	params := &CDRParamList{
		ContractAddress: mock.Address(),
		Params:          []*CDRParam{&p1, &p2},
	}

	if abi, err := params.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		params2 := &CDRParamList{}
		if err := params2.FromABI(abi); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(params, params2) {
			t.Fatalf("invalid cdr param list, exp: %v,act: %v", params, params2)
		} else {
			t.Log(util.ToString(params))
			t.Log(util.ToString(params2))
		}
	}
}

func TestCDRParamList_Verify(t *testing.T) {
	type fields struct {
		ContractAddress types.Address
		Params          []*CDRParam
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
				Params:          []*CDRParam{&cdrParam},
			},
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				ContractAddress: types.ZeroAddress,
				Params:          []*CDRParam{&cdrParam},
			},
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				ContractAddress: mock.Address(),
				Params:          nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRParamList{
				ContractAddress: tt.fields.ContractAddress,
				Params:          tt.fields.Params,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
