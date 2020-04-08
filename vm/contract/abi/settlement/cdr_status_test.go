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

func TestCDRStatus(t *testing.T) {
	param1 := cdrParam
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

func TestCDRStatus_State(t *testing.T) {
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
	type args struct {
		addr *types.Address
		fn   func(status *CDRStatus) (string, error)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   bool
		want2   bool
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdr2}, addr2.String(): {cdr1}},
				Status: SettlementStatusSuccess,
			},
			args: args{
				addr: &addr1,
				fn:   customerFn,
			},
			want:    "Tencent",
			want1:   true,
			want2:   true,
			wantErr: false,
		}, {
			name: "ok",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdr2}, addr2.String(): {cdr1}},
				Status: SettlementStatusSuccess,
			},
			args: args{
				addr: &addr1,
				fn:   accountFn,
			},
			want:    "Tencent_DIRECTS",
			want1:   true,
			want2:   true,
			wantErr: false,
		}, {
			name: "ok2",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdr1, cdr2}, addr2.String(): {cdr2}},
				Status: SettlementStatusSuccess,
			},
			args: args{
				addr: &addr1,
				fn:   customerFn,
			},
			want:    "Tencent",
			want1:   true,
			want2:   false,
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				Params: nil,
				Status: 0,
			},
			args: args{
				addr: &addr1,
				fn:   customerFn,
			},
			want:    "",
			want1:   false,
			want2:   false,
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdrParam}},
				Status: SettlementStatusSuccess,
			},
			args: args{
				addr: &addr2,
				fn:   customerFn,
			},
			want:    "",
			want1:   false,
			want2:   false,
			wantErr: true,
		}, {
			name: "f3",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdr1, cdr2}},
				Status: SettlementStatusStage1,
			},
			args: args{
				addr: &addr1,
				fn:   customerFn,
			},
			want:    "Tencent",
			want1:   false,
			want2:   false,
			wantErr: false,
		}, {
			name: "f4",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {}},
				Status: SettlementStatusSuccess,
			},
			args: args{
				addr: &addr1,
				fn:   customerFn,
			},
			want:    "",
			want1:   false,
			want2:   false,
			wantErr: false,
		}, {
			name: "f5",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {}, addr2.String(): {}},
				Status: SettlementStatusSuccess,
			},
			args: args{
				addr: &addr1,
				fn:   customerFn,
			},
			want:    "",
			want1:   true,
			want2:   false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRStatus{
				Params: tt.fields.Params,
				Status: tt.fields.Status,
			}
			got, got1, got2, err := z.State(tt.args.addr, tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("State() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("State() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("State() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("State() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestCDRStatus_ExtractID(t *testing.T) {
	addr1 := mock.Address()
	//addr2 := mock.Address()

	cdr1 := cdrParam
	cdr1.Sender = "WeChat"
	cdr2 := cdrParam
	cdr2.Sender = "WeChat"
	cdr2.Customer = "Tencent"

	type fields struct {
		Params map[string][]CDRParam
		Status SettlementStatus
	}
	tests := []struct {
		name            string
		fields          fields
		wantDt          int64
		wantSender      string
		wantDestination string
		wantErr         bool
	}{
		{
			name: "ok",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdr1, cdr2}, mock.Address().String(): {cdr1}},
				Status: SettlementStatusSuccess,
			},
			wantDt:          cdrParam.SmsDt,
			wantSender:      "Tencent",
			wantDestination: cdrParam.Destination,
			wantErr:         false,
		}, {
			name: "fail",
			fields: fields{
				Params: nil,
				Status: SettlementStatusSuccess,
			},
			wantDt:          0,
			wantSender:      "",
			wantDestination: "",
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRStatus{
				Params: tt.fields.Params,
				Status: tt.fields.Status,
			}
			gotDt, gotSender, gotDestination, err := z.ExtractID()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDt != tt.wantDt {
				t.Errorf("ExtractID() gotDt = %v, want %v", gotDt, tt.wantDt)
			}
			if gotSender != tt.wantSender {
				t.Errorf("ExtractID() gotSender = %v, want %v", gotSender, tt.wantSender)
			}
			if gotDestination != tt.wantDestination {
				t.Errorf("ExtractID() gotDestination = %v, want %v", gotDestination, tt.wantDestination)
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
