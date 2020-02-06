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
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/mock"
)

var (
	createContractParam = CreateContractParam{
		PartyA:      mock.Address(),
		PartyAName:  "c1",
		PartyB:      mock.Address(),
		PartyBName:  "c2",
		Previous:    mock.Hash(),
		ServiceId:   mock.Hash().String(),
		Mcc:         1,
		Mnc:         2,
		TotalAmount: 100,
		UnitPrice:   2,
		Currency:    "USD",
		SignDate:    time.Now().Unix(),
		SignatureA:  types.ZeroSignature,
	}
)

type contractContainer struct {
	contract *ContractParam
	a1       *types.Account
	a2       *types.Account
}

func setupSettlementTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "settlement", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)

	return func(t *testing.T) {
		//err := l.Store.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l
}

func buildContractParam() (param *ContractParam, a1 *types.Account, a2 *types.Account) {
	cp := createContractParam
	a1 = mock.Account()
	a2 = mock.Account()
	cp.PartyA = a1.Address()
	cp.PartyB = a2.Address()
	if err := cp.Sign(a1); err != nil {
		fmt.Println(err)
	}

	cd := time.Now().Unix()
	param = &ContractParam{
		CreateContractParam: cp,
		ConfirmDate:         cd,
		SignatureB:          nil,
	}
	_ = param.Sign(a2)
	return
}

func TestSettlement_CreateContractParam(t *testing.T) {
	a1 := mock.Account()
	if err := createContractParam.Sign(a1); err != nil {
		t.Fatal(err)
	}
	t.Log(createContractParam.String())
	addr1, err := createContractParam.Address()
	if err != nil {
		t.Fatal(err)
	}

	cp := ContractParam{
		CreateContractParam: createContractParam,
		ConfirmDate:         time.Now().Unix() + 100,
	}

	a2 := mock.Account()
	if err = cp.Sign(a2); err != nil {
		t.Fatal(err)
	}
	addr2, err := cp.Address()
	if err != nil {
		t.Fatal(err)
	}
	if addr1 != addr2 {
		t.Fatalf("invalid addr,%s==>%s", addr1, addr2)
	}
	t.Log(cp.String())

	msg, err := createContractParam.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	param2 := &CreateContractParam{}
	_, err = param2.UnmarshalMsg(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(createContractParam, *param2) {
		t.Fatalf("invaid createContractParam,%s,%s", createContractParam.String(), param2.String())
	}

	msg, err = cp.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	cp2 := &ContractParam{}
	_, err = cp2.UnmarshalMsg(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cp, *cp2) {
		t.Fatalf("invaid createContractParam,%s,%s", cp.String(), cp2.String())
	}
}

func TestSettlement_Pack_UnPack(t *testing.T) {
	a1 := mock.Account()
	param := createContractParam
	if err := param.Sign(a1); err != nil {
		t.Fatal(err)
	}
	//t.Log(param.String())
	if abi, err := param.ToABI(); err == nil {
		// test convert to json
		//var js bytes.Buffer
		//if _, err := msgp.UnmarshalAsJSON(&js, abi); err == nil {
		//	tmp := &CreateContractParam{}
		//	if err := json.Unmarshal(js.Bytes(), tmp); err == nil {
		//		t.Log(tmp.String())
		//	} else {
		//		t.Fatal(err)
		//	}
		//} else {
		//	t.Fatal(err)
		//}

		param2 := new(CreateContractParam)
		if err := param2.FromABI(abi); err != nil {
			t.Fatal(err)
		} else {
			a1, _ := param.Address()
			a2, _ := param2.Address()
			if a1 != a2 {
				t.Fatalf("invalid pack/unpack exp: %v, act: %v", param, param2)
			}
		}
	} else {
		t.Fatal(err)
	}
}

func TestSettlement_ContractParam_Equal(t *testing.T) {
	param, a1, a2 := buildContractParam()
	cp := param.CreateContractParam
	cp.PartyA = a1.Address()
	cp.PartyB = a2.Address()

	cp2 := cp
	cp2.PartyA = mock.Address()

	cp3 := cp
	cp3.SignatureA = types.ZeroSignature

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
				SignatureB:          param.SignatureB,
			},
			args: args{
				cp: &cp,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "invalid signB",
			fields: fields{
				CreateContractParam: cp,
				ConfirmDate:         param.ConfirmDate,
				SignatureB:          &types.ZeroSignature,
			},
			args: args{
				cp: &cp,
			},
			want:    true,
			wantErr: false,
		}, {
			name: "invalid signA",
			fields: fields{
				CreateContractParam: cp2,
				ConfirmDate:         param.ConfirmDate,
				SignatureB:          param.SignatureB,
			},
			args: args{
				cp: &cp,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "zero signA",
			fields: fields{
				CreateContractParam: cp3,
				ConfirmDate:         param.ConfirmDate,
				SignatureB:          param.SignatureB,
			},
			args: args{
				cp: &cp,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "nil",
			fields: fields{
				CreateContractParam: cp3,
				ConfirmDate:         param.ConfirmDate,
				SignatureB:          param.SignatureB,
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
				SignatureB:          tt.fields.SignatureB,
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

func TestSettlement_ContractParam_Sign(t *testing.T) {
	param, _, a2 := buildContractParam()

	type fields struct {
		CreateContractParam CreateContractParam
		ConfirmDate         int64
		SignatureB          *types.Signature
	}
	type args struct {
		account *types.Account
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				CreateContractParam: param.CreateContractParam,
				ConfirmDate:         param.ConfirmDate,
				SignatureB:          param.SignatureB,
			},
			args:    args{account: a2},
			wantErr: false,
		},
		{
			name: "invalid confirm data",
			fields: fields{
				CreateContractParam: param.CreateContractParam,
				ConfirmDate:         0,
				SignatureB:          param.SignatureB,
			},
			args:    args{account: a2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				ConfirmDate:         tt.fields.ConfirmDate,
				SignatureB:          tt.fields.SignatureB,
			}
			if err := z.Sign(tt.args.account); (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSettlement_ContractParam_String(t *testing.T) {
	param, _, _ := buildContractParam()
	s := param.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}
}

func TestSettlement_CreateContractParam_Address(t *testing.T) {
	param, _, _ := buildContractParam()
	cp := param.CreateContractParam

	address, err := cp.Address()
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		PartyA      types.Address
		PartyAName  string
		PartyB      types.Address
		PartyBName  string
		Previous    types.Hash
		ServiceId   string
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		UnitPrice   uint64
		Currency    string
		SignDate    int64
		SignatureA  types.Signature
	}
	tests := []struct {
		name    string
		fields  fields
		want    types.Address
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    address,
			wantErr: false,
		},
		{
			name: "invalid name",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  "",
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    types.ZeroAddress,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:      tt.fields.PartyA,
				PartyAName:  tt.fields.PartyAName,
				PartyB:      tt.fields.PartyB,
				PartyBName:  tt.fields.PartyBName,
				Previous:    tt.fields.Previous,
				ServiceId:   tt.fields.ServiceId,
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
				SignDate:    tt.fields.SignDate,
				SignatureA:  tt.fields.SignatureA,
			}
			got, err := z.Address()
			if (err != nil) != tt.wantErr {
				t.Errorf("Address() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Address() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateContractParam_Balance(t *testing.T) {
	param, _, _ := buildContractParam()
	cp := param.CreateContractParam

	type fields struct {
		PartyA      types.Address
		PartyAName  string
		PartyB      types.Address
		PartyBName  string
		Previous    types.Hash
		ServiceId   string
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		UnitPrice   uint64
		Currency    string
		SignDate    int64
		SignatureA  types.Signature
	}
	tests := []struct {
		name    string
		fields  fields
		want    types.Balance
		wantErr bool
	}{
		{
			name: "equal",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    types.Balance{Int: new(big.Int).Mul(new(big.Int).SetUint64(cp.TotalAmount), new(big.Int).SetUint64(cp.UnitPrice))},
			wantErr: false,
		},
		{
			name: "overflow",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: math.MaxUint64,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    types.ZeroBalance,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:      tt.fields.PartyA,
				PartyAName:  tt.fields.PartyAName,
				PartyB:      tt.fields.PartyB,
				PartyBName:  tt.fields.PartyBName,
				Previous:    tt.fields.Previous,
				ServiceId:   tt.fields.ServiceId,
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
				SignDate:    tt.fields.SignDate,
				SignatureA:  tt.fields.SignatureA,
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

func TestSettlement_CreateContractParam_FromABI(t *testing.T) {
	param, _, _ := buildContractParam()
	cp := param.CreateContractParam
	abi, err := cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		PartyA      types.Address
		PartyAName  string
		PartyB      types.Address
		PartyBName  string
		Previous    types.Hash
		ServiceId   string
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		UnitPrice   uint64
		Currency    string
		SignDate    int64
		SignatureA  types.Signature
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
			name: "equal",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			args: args{
				data: abi,
			},
			wantErr: false,
		}, {
			name: "invalid data",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			args: args{
				data: abi[5:],
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:      tt.fields.PartyA,
				PartyAName:  tt.fields.PartyAName,
				PartyB:      tt.fields.PartyB,
				PartyBName:  tt.fields.PartyBName,
				Previous:    tt.fields.Previous,
				ServiceId:   tt.fields.ServiceId,
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
				SignDate:    tt.fields.SignDate,
				SignatureA:  tt.fields.SignatureA,
			}
			if err := z.FromABI(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("FromABI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSettlement_CreateContractParam_Sign(t *testing.T) {
	param, a1, _ := buildContractParam()
	cp := param.CreateContractParam

	type fields struct {
		PartyA      types.Address
		PartyAName  string
		PartyB      types.Address
		PartyBName  string
		Previous    types.Hash
		ServiceId   string
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		UnitPrice   uint64
		Currency    string
		SignDate    int64
		SignatureA  types.Signature
	}
	type args struct {
		account *types.Account
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			args: args{
				account: a1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:      tt.fields.PartyA,
				PartyAName:  tt.fields.PartyAName,
				PartyB:      tt.fields.PartyB,
				PartyBName:  tt.fields.PartyBName,
				Previous:    tt.fields.Previous,
				ServiceId:   tt.fields.ServiceId,
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
				SignDate:    tt.fields.SignDate,
				SignatureA:  tt.fields.SignatureA,
			}
			if err := z.Sign(tt.args.account); (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSettlement_CreateContractParam_String(t *testing.T) {
	param, _, _ := buildContractParam()
	cp := param.CreateContractParam
	s := cp.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}
}

func TestSettlement_CreateContractParam_ToABI(t *testing.T) {
	param, _, _ := buildContractParam()
	cp := param.CreateContractParam

	abi, err := cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		PartyA      types.Address
		PartyAName  string
		PartyB      types.Address
		PartyBName  string
		Previous    types.Hash
		ServiceId   string
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		UnitPrice   uint64
		Currency    string
		SignDate    int64
		SignatureA  types.Signature
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    abi,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:      tt.fields.PartyA,
				PartyAName:  tt.fields.PartyAName,
				PartyB:      tt.fields.PartyB,
				PartyBName:  tt.fields.PartyBName,
				Previous:    tt.fields.Previous,
				ServiceId:   tt.fields.ServiceId,
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
				SignDate:    tt.fields.SignDate,
				SignatureA:  tt.fields.SignatureA,
			}
			got, err := z.ToABI()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToABI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToABI() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSettlement_CreateContractParam_ToContractParam(t *testing.T) {
	param, _, _ := buildContractParam()
	cp := param.CreateContractParam
	exp := &ContractParam{
		CreateContractParam: cp,
		ConfirmDate:         0,
		SignatureB:          nil,
	}
	type fields struct {
		PartyA      types.Address
		PartyAName  string
		PartyB      types.Address
		PartyBName  string
		Previous    types.Hash
		ServiceId   string
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		UnitPrice   uint64
		Currency    string
		SignDate    int64
		SignatureA  types.Signature
	}
	tests := []struct {
		name   string
		fields fields
		want   *ContractParam
	}{
		{
			name: "normal",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want: exp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:      tt.fields.PartyA,
				PartyAName:  tt.fields.PartyAName,
				PartyB:      tt.fields.PartyB,
				PartyBName:  tt.fields.PartyBName,
				Previous:    tt.fields.Previous,
				ServiceId:   tt.fields.ServiceId,
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
				SignDate:    tt.fields.SignDate,
				SignatureA:  tt.fields.SignatureA,
			}
			if got := z.ToContractParam(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToContractParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSettlement_GetContractsByAddress(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	var contracts []*contractContainer
	for i := 0; i < 4; i++ {
		param, a1, a2 := buildContractParam()
		contracts = append(contracts, &contractContainer{
			contract: param,
			a1:       a1,
			a2:       a2,
		})
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

	a := contracts[0].a1.Address()
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
			want:    []*ContractParam{contracts[0].contract},
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

func mockContractData(size int) []*contractContainer {
	var contracts []*contractContainer
	accounts := []*types.Account{mock.Account(), mock.Account()}

	for i := 0; i < size; i++ {
		cp := createContractParam
		var a1 *types.Account
		var a2 *types.Account
		if i%2 == 0 {
			a1 = accounts[0]
			a2 = accounts[1]
		} else {
			a1 = accounts[1]
			a2 = accounts[0]
		}
		cp.PartyA = a1.Address()
		cp.PartyB = a2.Address()
		if err := cp.Sign(a1); err != nil {
			fmt.Println(err)
		}

		cd := time.Now().Unix()
		param := &ContractParam{
			CreateContractParam: cp,
			ConfirmDate:         cd,
			SignatureB:          nil,
		}
		_ = param.Sign(a2)
		contracts = append(contracts, &contractContainer{
			contract: param,
			a1:       a1,
			a2:       a2,
		})
	}
	return contracts
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
		a, _ := d.contract.Address()
		abi, _ := d.contract.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	a1 := data[0].a1.Address()

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
			want:    []*ContractParam{data[0].contract},
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
		a, _ := d.contract.Address()
		abi, _ := d.contract.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	a1 := mock.Address()
	a2 := data[0].a2.Address()

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
			want:    []*ContractParam{data[0].contract},
			wantErr: false,
		},
		{
			name: "empty",
			args: args{
				ctx:  ctx,
				addr: &a1,
			},
			want:    []*ContractParam{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsIDByAddressAsPartyB(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContractsIDByAddressAsPartyB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("GetContractsIDByAddressAsPartyB() len(go) != len(tt.want), %d,%d", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				a1, _ := got[i].Address()
				a2, _ := tt.want[i].Address()
				if a1 != a2 {
					t.Fatalf("GetContractsIDByAddressAsPartyB() i[%d] %v,%v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSettlement_IsContractAvailable(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)

	var contracts []*contractContainer

	for i := 0; i < 2; i++ {
		param, a1, a2 := buildContractParam()
		if i%2 == 1 {
			param.SignatureB = &types.ZeroSignature
		}
		contracts = append(contracts, &contractContainer{
			contract: param,
			a1:       a1,
			a2:       a2,
		})
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

	a1, _ := contracts[0].contract.Address()
	a2, _ := contracts[1].contract.Address()

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

func TestSettlement_ParseContractParam(t *testing.T) {
	param, _, _ := buildContractParam()
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

func TestSettlement_SignContract_ToABI(t *testing.T) {
	a := mock.Account()
	sc := &SignContractParam{
		ContractAddress: mock.Address(),
		ConfirmDate:     time.Now().Unix(),
		SignatureB:      types.Signature{},
	}

	if h, err := types.HashBytes(sc.ContractAddress[:], util.BE_Int2Bytes(sc.ConfirmDate)); err != nil {
		t.Fatal(err)
	} else {
		sc.SignatureB = a.Sign(h)
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

	addr := a.Address()
	if verify, err := sc.Verify(addr); err != nil {
		t.Fatal(err)
	} else {
		if !verify {
			t.Fatalf("veirfy signature failed, got: %s", sc.SignatureB.String())
		}
	}
}

func TestSettlement_CreateContractParam_Verify(t *testing.T) {
	param, _, _ := buildContractParam()
	cp := param.CreateContractParam

	type fields struct {
		PartyA      types.Address
		PartyAName  string
		PartyB      types.Address
		PartyBName  string
		Previous    types.Hash
		ServiceId   string
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		UnitPrice   uint64
		Currency    string
		SignDate    int64
		SignatureA  types.Signature
	}

	cp2 := cp
	cp2.SignatureA = types.ZeroSignature

	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "failure",
			fields: fields{
				PartyA:      cp2.PartyA,
				PartyAName:  cp2.PartyAName,
				PartyB:      cp2.PartyB,
				PartyBName:  cp2.PartyBName,
				Previous:    cp2.Previous,
				ServiceId:   cp2.ServiceId,
				Mcc:         cp2.Mcc,
				Mnc:         cp2.Mnc,
				TotalAmount: cp2.TotalAmount,
				UnitPrice:   cp2.UnitPrice,
				Currency:    cp2.Currency,
				SignDate:    cp2.SignDate,
				SignatureA:  cp2.SignatureA,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:      tt.fields.PartyA,
				PartyAName:  tt.fields.PartyAName,
				PartyB:      tt.fields.PartyB,
				PartyBName:  tt.fields.PartyBName,
				Previous:    tt.fields.Previous,
				ServiceId:   tt.fields.ServiceId,
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
				SignDate:    tt.fields.SignDate,
				SignatureA:  tt.fields.SignatureA,
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

func TestCreateContractParam_verifyParam(t *testing.T) {
	param, _, _ := buildContractParam()
	cp := param.CreateContractParam

	type fields struct {
		PartyA      types.Address
		PartyAName  string
		PartyB      types.Address
		PartyBName  string
		Previous    types.Hash
		ServiceId   string
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		UnitPrice   uint64
		Currency    string
		SignDate    int64
		SignatureA  types.Signature
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
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "f1",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  "",
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "f2",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      types.ZeroAddress,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "f3",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    types.ZeroHash,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "f4",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   "",
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "f5",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: 0,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "f6",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   0,
				Currency:    cp.Currency,
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f7",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    "",
				SignDate:    cp.SignDate,
				SignatureA:  cp.SignatureA,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f8",
			fields: fields{
				PartyA:      cp.PartyA,
				PartyAName:  cp.PartyAName,
				PartyB:      cp.PartyB,
				PartyBName:  cp.PartyBName,
				Previous:    cp.Previous,
				ServiceId:   cp.ServiceId,
				Mcc:         cp.Mcc,
				Mnc:         cp.Mnc,
				TotalAmount: cp.TotalAmount,
				UnitPrice:   cp.UnitPrice,
				Currency:    cp.Currency,
				SignDate:    0,
				SignatureA:  cp.SignatureA,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:      tt.fields.PartyA,
				PartyAName:  tt.fields.PartyAName,
				PartyB:      tt.fields.PartyB,
				PartyBName:  tt.fields.PartyBName,
				Previous:    tt.fields.Previous,
				ServiceId:   tt.fields.ServiceId,
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
				SignDate:    tt.fields.SignDate,
				SignatureA:  tt.fields.SignatureA,
			}
			got, err := z.verifyParam()
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("verifyParam() got = %v, want %v", got, tt.want)
			}
		})
	}
}
