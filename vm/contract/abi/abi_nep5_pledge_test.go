/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestStringToPledgeType(t *testing.T) {
	type args struct {
		sType string
	}
	tests := []struct {
		name    string
		args    args
		want    PledgeType
		wantErr bool
	}{
		{
			name: "network",
			args: args{
				sType: "network",
			},
			want:    Network,
			wantErr: false,
		}, {
			name: "confidant",
			args: args{
				sType: "confidant",
			},
			want:    Network,
			wantErr: false,
		}, {
			name: "oracle",
			args: args{
				sType: "oracle",
			},
			want:    Oracle,
			wantErr: false,
		}, {
			name: "vote",
			args: args{
				sType: "vote",
			},
			want:    Vote,
			wantErr: false,
		}, {
			name: "invalid",
			args: args{
				sType: "",
			},
			want:    Invalid,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringToPledgeType(tt.args.sType)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToPledgeType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StringToPledgeType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBeneficialInfos(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	infos, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) == 0 {
		t.Fatal("invalid generate data...")
	}

	b := infos[0].Beneficial
	info, balance := GetBeneficialInfos(ctx, b)
	if info == nil {
		t.Fatal("invalid pledge info")
	}
	if balance.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("invalid amount, exp: 100, act: %d", balance)
	}
}

func TestGetBeneficialPledgeInfos(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	infos, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) == 0 {
		t.Fatal("invalid generate data...")
	}

	b := infos[0].Beneficial
	info2, balance := GetBeneficialPledgeInfos(ctx, b, Network)
	if info2 == nil {
		t.Fatal("invalid pledge info")
	}
	if balance.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("invalid amount, exp: 100, act: %d", balance)
	}

	info3, balance3 := GetBeneficialPledgeInfos(ctx, b, Oracle)
	if info3 != nil {
		t.Fatal("invalid pledge info")
	}
	if balance3.Cmp(big.NewInt(0)) != 0 {
		t.Fatalf("invalid amount, exp: 0, act: %d", balance3)
	}
}

func TestGetPledgeBeneficialAmount(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	infos, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}

	if len(infos) == 0 {
		t.Fatal("invalid generate data...")
	}
	b := infos[0].Beneficial
	if amount := GetPledgeBeneficialAmount(ctx, b, uint8(Network)); amount.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("invalid amount, exp: 100, act: %d", amount)
	}
	if amount := GetPledgeBeneficialAmount(ctx, b, uint8(Oracle)); amount.Cmp(big.NewInt(0)) != 0 {
		t.Fatalf("invalid amount, exp: 0, act: %d", amount)
	}
}

func TestGetPledgeBeneficialTotalAmount(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	_, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}

	if amount, err := GetPledgeBeneficialTotalAmount(ctx, a); err != nil {
		t.Fatal(err)
	} else if amount.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("invalid amount, exp: 100, act: %d", amount)
	}
}

func TestGetPledgeInfos(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	_, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}

	infos, b := GetPledgeInfos(ctx, a)
	if len(infos) != 4 {
		t.Fatalf("invalid infos len, exp: 4, act: %d", len(infos))
	}

	if b.Cmp(big.NewInt(400)) != 0 {
		t.Fatalf("invalid total balance, exp: 400, act: %d", b)
	}
}

func TestGetPledgeKey(t *testing.T) {
	if key := GetPledgeKey(mock.Address(), mock.Address(), mock.Hash().String()); key == nil {
		t.Fatal()
	}
}

func mockPledgeInfo(ctx *vmstore.VMContext, addr types.Address, size int) ([]*NEP5PledgeInfo, error) {
	var infos []*NEP5PledgeInfo
	for i := 0; i < size; i++ {
		b := mock.Address()
		if i == 0 {
			b = addr
		}
		info := &NEP5PledgeInfo{
			PType:         uint8(Network),
			Amount:        big.NewInt(100),
			WithdrawTime:  time.Now().Unix(),
			Beneficial:    b,
			PledgeAddress: mock.Address(),
			NEP5TxId:      mock.Hash().String(),
		}
		if data, err := info.ToABI(); err != nil {
			return nil, err
		} else {
			pledgeKey := GetPledgeKey(addr, info.Beneficial, info.NEP5TxId)
			if err := ctx.SetStorage(types.NEP5PledgeAddress[:], pledgeKey, data); err != nil {
				return nil, err
			}
			infos = append(infos, info)
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		return nil, err
	}
	return infos, nil
}

func TestGetTotalPledgeAmount(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	_, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}

	amount := GetTotalPledgeAmount(ctx)
	if amount.Cmp(big.NewInt(400)) != 0 {
		t.Fatalf("invalid total balance, exp: 400, act: %d", amount)
	}
}

func TestParsePledgeInfo(t *testing.T) {
	info := &NEP5PledgeInfo{
		PType:         uint8(Network),
		Amount:        big.NewInt(100),
		WithdrawTime:  time.Now().Unix(),
		Beneficial:    mock.Address(),
		PledgeAddress: mock.Address(),
		NEP5TxId:      mock.Hash().String(),
	}
	if data, err := info.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		if info2, err := ParsePledgeInfo(data); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(info, info2) {
			t.Fatalf("exp: %v, act: %v", info, info2)
		}
	}
}

func TestParsePledgeParam(t *testing.T) {
	param := &PledgeParam{
		Beneficial:    mock.Address(),
		PledgeAddress: mock.Address(),
		PType:         uint8(Network),
		NEP5TxId:      mock.Hash().String(),
	}
	if data, err := param.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		if p2, err := ParsePledgeParam(data); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(param, p2) {
			t.Fatalf("exp: %v, act: %v", param, p2)
		}
	}
}

func TestSearchAllPledgeInfos(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	_, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}

	if infos, err := SearchAllPledgeInfos(ctx); err != nil {
		t.Fatal(err)
	} else if len(infos) != 4 {
		t.Fatalf("invalid infos len, exp: 4, act: %d", len(infos))
	}
}

func TestSearchBeneficialPledgeInfo(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	infos, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 4 {
		t.Fatal()
	}

	type args struct {
		ctx   *vmstore.VMContext
		param *WithdrawPledgeParam
	}
	tests := []struct {
		name string
		args args
		want []*PledgeResult
	}{
		{
			name: "ok",
			args: args{
				ctx: ctx,
				param: &WithdrawPledgeParam{
					Beneficial: infos[0].Beneficial,
					Amount:     infos[0].Amount,
					PType:      infos[0].PType,
					NEP5TxId:   infos[0].NEP5TxId,
				},
			},
			want: []*PledgeResult{{
				Key:        []byte{},
				PledgeInfo: infos[0],
			}},
		}, {
			name: "f",
			args: args{
				ctx: ctx,
				param: &WithdrawPledgeParam{
					Beneficial: infos[0].Beneficial,
					Amount:     infos[0].Amount,
					PType:      infos[0].PType,
					NEP5TxId:   mock.Hash().String(),
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchBeneficialPledgeInfo(tt.args.ctx, tt.args.param); got == nil && tt.want != nil {
				t.Errorf("SearchBeneficialPledgeInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchBeneficialPledgeInfoByTxId(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	infos, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 4 {
		t.Fatal()
	}

	type args struct {
		ctx   *vmstore.VMContext
		param *WithdrawPledgeParam
	}
	tests := []struct {
		name string
		args args
		want *PledgeResult
	}{
		{
			name: "ok",
			args: args{
				ctx: ctx,
				param: &WithdrawPledgeParam{
					Beneficial: infos[0].Beneficial,
					Amount:     infos[0].Amount,
					PType:      infos[0].PType,
					NEP5TxId:   infos[0].NEP5TxId,
				},
			},
			want: &PledgeResult{
				Key:        []byte{},
				PledgeInfo: infos[0],
			},
		},
		{
			name: "f",
			args: args{
				ctx: ctx,
				param: &WithdrawPledgeParam{
					Beneficial: infos[0].Beneficial,
					Amount:     infos[0].Amount,
					PType:      uint8(Oracle),
					NEP5TxId:   infos[0].NEP5TxId,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchBeneficialPledgeInfoByTxId(tt.args.ctx, tt.args.param); got == nil && tt.want != nil {
				t.Errorf("SearchBeneficialPledgeInfoByTxId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchBeneficialPledgeInfoIgnoreWithdrawTime(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	infos, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 4 {
		t.Fatal()
	}

	type args struct {
		ctx   *vmstore.VMContext
		param *WithdrawPledgeParam
	}
	tests := []struct {
		name string
		args args
		want []*PledgeResult
	}{
		{
			name: "ok",
			args: args{
				ctx: ctx,
				param: &WithdrawPledgeParam{
					Beneficial: infos[0].Beneficial,
					Amount:     infos[0].Amount,
					PType:      infos[0].PType,
					NEP5TxId:   infos[0].NEP5TxId,
				},
			},
			want: []*PledgeResult{
				{
					Key:        []byte{},
					PledgeInfo: infos[0],
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchBeneficialPledgeInfoIgnoreWithdrawTime(tt.args.ctx, tt.args.param); got == nil && tt.want != nil {
				t.Errorf("SearchBeneficialPledgeInfoIgnoreWithdrawTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchPledgeInfoWithNEP5TxId(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	infos, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 4 {
		t.Fatal()
	}

	type args struct {
		ctx   *vmstore.VMContext
		param *WithdrawPledgeParam
	}
	tests := []struct {
		name string
		args args
		want *PledgeResult
	}{
		{
			name: "ok",
			args: args{
				ctx: ctx,
				param: &WithdrawPledgeParam{
					Beneficial: infos[0].Beneficial,
					Amount:     infos[0].Amount,
					PType:      infos[0].PType,
					NEP5TxId:   infos[0].NEP5TxId,
				},
			},
			want: &PledgeResult{
				Key:        []byte{},
				PledgeInfo: infos[0],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchPledgeInfoWithNEP5TxId(tt.args.ctx, tt.args.param); got == nil && tt.want != nil {
				t.Errorf("SearchPledgeInfoWithNEP5TxId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_searchBeneficialPledgeInfoByTxId(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l)
	a := mock.Address()

	infos, err := mockPledgeInfo(ctx, a, 4)
	if err != nil {
		t.Fatal(err)
	}

	if len(infos) != 4 {
		t.Fatal()
	}

	type args struct {
		ctx   *vmstore.VMContext
		param *WithdrawPledgeParam
	}
	tests := []struct {
		name string
		args args
		want *PledgeResult
	}{
		{
			name: "ok",
			args: args{
				ctx: ctx,
				param: &WithdrawPledgeParam{
					Beneficial: infos[0].Beneficial,
					Amount:     infos[0].Amount,
					PType:      infos[0].PType,
					NEP5TxId:   infos[0].NEP5TxId,
				},
			},
			want: &PledgeResult{
				Key:        []byte{},
				PledgeInfo: infos[0],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchBeneficialPledgeInfoByTxId(tt.args.ctx, tt.args.param); got == nil && tt.want != nil {
				t.Errorf("searchBeneficialPledgeInfoByTxId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseWithdrawPledgeParam(t *testing.T) {
	param := &WithdrawPledgeParam{
		Beneficial: mock.Address(),
		Amount:     big.NewInt(100),
		PType:      uint8(Network),
		NEP5TxId:   mock.Hash().String(),
	}

	if abi, err := param.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		if p2, err := ParseWithdrawPledgeParam(abi); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(param, p2) {
			t.Fatalf("invalid withdraw param, exp: %v, act: %v", param, p2)
		}
	}
}
