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
		}, {
			name: "ok",
			fields: fields{
				CreateContractParam: param.CreateContractParam,
				PreStops:            []string{},
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

func TestContractParam_IsContractor(t *testing.T) {
	cp := buildContractParam()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
		Terminator          *Terminator
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
		Terminator          *Terminator
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
				status:   ContractStatusActivated,
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
				Status:              ContractStatusActivated,
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
		Terminator          *Terminator
	}
	type args struct {
		operator *Terminator
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
				operator: &Terminator{Request: true, Address: cp.PartyA.Address},
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
				operator: &Terminator{Request: true, Address: cp.PartyB.Address},
				status:   ContractStatusRejected,
			},
			wantErr: false,
		}, {
			name: "partyA_destroy_activated",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActivated,
			},
			args: args{
				operator: &Terminator{Request: true, Address: cp.PartyA.Address},
				status:   ContractStatusDestroyStage1,
			},
			wantErr: false,
		}, {
			name: "partyA_cancel_destroyed",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyStage1,
				Terminator:          &Terminator{Request: true, Address: cp.PartyA.Address},
			},
			args: args{
				operator: &Terminator{Request: false, Address: cp.PartyA.Address},
				status:   ContractStatusActivated,
			},
			wantErr: false,
		}, {
			name: "partyB_confirm_destroy",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyStage1,
				Terminator:          &Terminator{Request: true, Address: cp.PartyA.Address},
			},
			args: args{
				operator: &Terminator{Request: true, Address: cp.PartyB.Address},
				status:   ContractStatusDestroyed,
			},
			wantErr: false,
		}, {
			name: "partyB_reject_destroy",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyStage1,
				Terminator:          &Terminator{Request: true, Address: cp.PartyA.Address},
			},
			args: args{
				operator: &Terminator{Request: false, Address: cp.PartyB.Address},
				status:   ContractStatusActivated,
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
				Terminator:          &Terminator{Request: true, Address: cp.PartyA.Address},
			},
			args: args{
				operator: &Terminator{Request: true, Address: cp.PartyA.Address},
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
				Terminator:          &Terminator{Request: true, Address: cp.PartyA.Address},
			},
			args: args{
				operator: &Terminator{Request: true, Address: mock.Address()},
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
				Terminator:          &Terminator{Request: true, Address: cp.PartyA.Address},
			},
			args: args{
				operator: &Terminator{Request: true, Address: cp.PartyB.Address},
				status:   ContractStatusDestroyed,
			},
			wantErr: true,
		}, {
			name: "f4",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyed,
				Terminator:          &Terminator{Request: true, Address: cp.PartyA.Address},
			},
			args: args{
				operator: nil,
				status:   ContractStatusDestroyed,
			},
			wantErr: true,
		}, {
			name: "f5",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActivated,
			},
			args: args{
				operator: &Terminator{Request: false, Address: cp.PartyA.Address},
				status:   ContractStatusDestroyed,
			},
			wantErr: true,
		},
		{
			name: "f6",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusRejected,
			},
			args: args{
				operator: &Terminator{Request: true, Address: cp.PartyA.Address},
				status:   ContractStatusRejected,
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

func TestContractParam_IsAvailable(t *testing.T) {
	cp := buildContractParam()
	cp.StartDate = time.Now().AddDate(0, 0, -1).Unix()
	cp.EndDate = time.Now().AddDate(0, 0, 5).Unix()
	cp.Status = ContractStatusActivated

	cp2 := buildContractParam()
	cp2.StartDate = time.Now().AddDate(0, 0, -3).Unix()
	cp2.EndDate = time.Now().AddDate(0, 0, -2).Unix()

	cp3 := buildContractParam()
	cp3.StartDate = time.Now().AddDate(0, 0, 1).Unix()
	cp3.EndDate = time.Now().AddDate(0, 0, 2).Unix()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
		Terminator          *Terminator
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "ok",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              cp.Status,
				Terminator:          nil,
			},
			want: true,
		}, {
			name: "f1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              ContractStatusActiveStage1,
				Terminator:          nil,
			},
			want: false,
		}, {
			name: "f2",
			fields: fields{
				CreateContractParam: cp2.CreateContractParam,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              cp.Status,
				Terminator:          nil,
			},
			want: false,
		}, {
			name: "f3",
			fields: fields{
				CreateContractParam: cp2.CreateContractParam,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              cp.Status,
				Terminator:          nil,
			},
			want: false,
		}, {
			name: "f4",
			fields: fields{
				CreateContractParam: cp3.CreateContractParam,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              cp.Status,
				Terminator:          nil,
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
			if got := z.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContractParam_IsExpired(t *testing.T) {
	cp := buildContractParam()
	cp.StartDate = time.Now().Unix()
	cp.EndDate = time.Now().AddDate(0, 0, -5).Unix()
	cp.Status = ContractStatusActivated

	cp2 := buildContractParam()
	cp2.StartDate = time.Now().AddDate(0, 0, -1).Unix()
	cp2.EndDate = time.Now().AddDate(0, 0, 2).Unix()

	cp3 := buildContractParam()
	cp3.StartDate = time.Now().AddDate(0, 0, 1).Unix()
	cp3.EndDate = time.Now().AddDate(0, 0, 2).Unix()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
		Terminator          *Terminator
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "ok",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              cp.Status,
				Terminator:          nil,
			},
			want: true,
		}, {
			name: "f1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              ContractStatusActiveStage1,
				Terminator:          nil,
			},
			want: false,
		}, {
			name: "f2",
			fields: fields{
				CreateContractParam: cp2.CreateContractParam,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              cp.Status,
				Terminator:          nil,
			},
			want: false,
		}, {
			name: "f3",
			fields: fields{
				CreateContractParam: cp2.CreateContractParam,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              cp.Status,
				Terminator:          nil,
			},
			want: false,
		}, {
			name: "f4",
			fields: fields{
				CreateContractParam: cp3.CreateContractParam,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              cp.Status,
				Terminator:          nil,
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
			if got := z.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
