/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"math/big"
	"testing"

	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestDestroyParam_Signature(t *testing.T) {
	a := mock.Account()
	addr := a.Address()

	param := &DestroyParam{
		Owner:    addr,
		Previous: mock.Hash(),
		Token:    config.GasToken(),
		Amount:   big.NewInt(10),
		Sign:     types.Signature{},
	}
	var err error
	if param.Sign, err = param.Signature(a); err != nil {
		t.Fatal(err)
	} else {
		if b, err := param.Verify(); err != nil {
			t.Fatal(err)
		} else if !b {
			t.Fatal()
		}
	}
}

func TestPackSendBlock(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l, &contractaddress.BlackHoleAddress)
	a1 := account1.Address()

	if tm, err := l.GetTokenMeta(a1, config.GasToken()); err != nil {
		t.Fatal(err)
	} else {
		param := &DestroyParam{
			Owner:    a1,
			Previous: tm.Header,
			Token:    config.GasToken(),
			Amount:   big.NewInt(10),
			Sign:     types.Signature{},
		}
		param.Sign, _ = param.Signature(account1)
		t.Log(param.String())

		if blk, err := PackSendBlock(ctx, param); err != nil {
			t.Fatal(err)
		} else {
			t.Log(blk)
		}
	}
}

func TestGetTotalDestroyInfo(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l, &contractaddress.BlackHoleAddress)

	a := mock.Address()
	h := mock.Hash()

	if data, err := BlackHoleABI.PackVariable(VariableDestroyInfo, a, h,
		config.GasToken(), big.NewInt(10), common.TimeNow().Unix()); err != nil {
		t.Fatal(err)
	} else {
		if err := ctx.SetStorage(a[:], h[:], data); err != nil {
			t.Fatal(err)
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	exp := types.Balance{Int: big.NewInt(10)}

	if balance, err := GetTotalDestroyInfo(ctx, &a); err != nil {
		t.Fatal(err)
	} else if balance.Compare(exp) != types.BalanceCompEqual {
		t.Fatalf("exp:%s, act: %s", exp.String(), balance.String())
	}
	if balance, err := GetTotalDestroyInfo(ctx, &a); err != nil {
		t.Fatal(err)
	} else if balance.Compare(exp) != types.BalanceCompEqual {
		t.Fatalf("exp:%s, act: %s", exp.String(), balance.String())
	}
}

func TestGetDestroyInfoDetail(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l, &contractaddress.BlackHoleAddress)

	a := mock.Address()
	h := mock.Hash()

	if data, err := BlackHoleABI.PackVariable(VariableDestroyInfo, a, h,
		config.GasToken(), big.NewInt(10), common.TimeNow().Unix()); err != nil {
		t.Fatal(err)
	} else {
		if err := ctx.SetStorage(a[:], h[:], data); err != nil {
			t.Fatal(err)
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	if details, err := GetDestroyInfoDetail(ctx, &a); err != nil {
		t.Fatal(err)
	} else {
		if len(details) != 1 {
			t.Fatal()
		}
		for _, detail := range details {
			t.Log(detail)
		}
	}

}

func TestDestroyParam_Verify(t *testing.T) {
	a := mock.Account()
	addr := a.Address()

	param := &DestroyParam{
		Owner:    addr,
		Previous: mock.Hash(),
		Token:    config.GasToken(),
		Amount:   big.NewInt(10),
		Sign:     types.Signature{},
	}
	param.Sign, _ = param.Signature(a)
	type fields struct {
		Owner    types.Address
		Previous types.Hash
		Token    types.Hash
		Amount   *big.Int
		Sign     types.Signature
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
				Owner:    param.Owner,
				Previous: param.Previous,
				Token:    param.Token,
				Amount:   param.Amount,
				Sign:     param.Sign,
			},
			want:    true,
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				Owner:    types.Address{},
				Previous: param.Previous,
				Token:    param.Token,
				Amount:   param.Amount,
				Sign:     param.Sign,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				Owner:    param.Owner,
				Previous: types.ZeroHash,
				Token:    param.Token,
				Amount:   param.Amount,
				Sign:     param.Sign,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f3",
			fields: fields{
				Owner:    param.Owner,
				Previous: param.Previous,
				Token:    mock.Hash(),
				Amount:   param.Amount,
				Sign:     param.Sign,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f4",
			fields: fields{
				Owner:    param.Owner,
				Previous: param.Previous,
				Token:    param.Token,
				Amount:   big.NewInt(0),
				Sign:     param.Sign,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f5",
			fields: fields{
				Owner:    param.Owner,
				Previous: param.Previous,
				Token:    param.Token,
				Amount:   param.Amount,
				Sign:     types.Signature{},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := &DestroyParam{
				Owner:    tt.fields.Owner,
				Previous: tt.fields.Previous,
				Token:    tt.fields.Token,
				Amount:   tt.fields.Amount,
				Sign:     tt.fields.Sign,
			}
			got, err := param.Verify()
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
