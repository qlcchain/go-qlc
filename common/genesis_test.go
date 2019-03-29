/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"math/big"
	"reflect"
	"testing"
)

func TestGenesisBlock(t *testing.T) {
	t.Skip()
	h, _ := types.NewHash("6fb8b79a65d77b9d54fc0a8d44e2f522afe426426b54f8680125c71d84122d1b")
	genesis := GenesisBlock()
	h2 := genesis.GetHash()
	if h2 != h {
		t.Log(util.ToString(genesis))
		t.Fatal("invalid genesis block", h2.String(), h.String())
	}

	h3, _ := types.NewHash("8407598dfb87ebc48a0926272599d99bf47a6447dc19bfd1de61e2789eeea8a8")
	mintage := GenesisMintageBlock()
	h4 := mintage.GetHash()
	if h3 != h4 {
		t.Fatal(h3.String(), h4.String())
	}
}

func TestGenesisBlock2(t *testing.T) {
	t.Skip()
	h, _ := types.NewHash("d40525ddc07849fdfda84679acc2b279207fa711f0e20d96ee115158b8061bf4")
	genesis := testGenesisBlock
	h2 := genesis.GetHash()
	if h2 != h {
		t.Log(util.ToString(genesis))
		t.Fatal("invalid genesis block", h2.String(), h.String())
	}

	h3, _ := types.NewHash("989b152a7bbfcc7e120772f563a9edaa087d655fab6daec1ddf411ad1e080943")
	mintage := testGenesisMintageBlock
	h4 := mintage.GetHash()
	if h3 != h4 {
		t.Fatal(h3.String(), h4.String())
	}
}

func TestBalanceToRaw(t *testing.T) {
	b1 := types.Balance{Int: big.NewInt(2)}
	i, _ := new(big.Int).SetString("200000000", 10)
	b2 := types.Balance{Int: i}

	type args struct {
		b    types.Balance
		unit string
	}
	tests := []struct {
		name    string
		args    args
		want    types.Balance
		wantErr bool
	}{
		{"Mqlc", args{b: b1, unit: "QLC"}, b2, false},
		//{"Mqn1", args{b: b1, unit: "QN1"}, b2, false},
		//{"Mqn3", args{b: b1, unit: "QN3"}, b2, false},
		//{"Mqn5", args{b: b1, unit: "QN5"}, b2, false},
		//{"Mqn6", args{b: b1, unit: "QN6"}, b1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BalanceToRaw(tt.args.b, tt.args.unit)
			if (err != nil) != tt.wantErr {
				t.Errorf("BalanceToRaw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BalanceToRaw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawToBalance(t *testing.T) {
	b1 := types.Balance{Int: big.NewInt(2)}
	i, _ := new(big.Int).SetString("200000000", 10)
	b2 := types.Balance{Int: i}
	type args struct {
		b    types.Balance
		unit string
	}
	tests := []struct {
		name    string
		args    args
		want    types.Balance
		wantErr bool
	}{
		{"Mqlc", args{b: b2, unit: "QLC"}, b1, false},
		//{"Mqn1", args{b: b2, unit: "QN1"}, b1, false},
		//{"Mqn3", args{b: b2, unit: "QN3"}, b1, false},
		//{"Mqn5", args{b: b2, unit: "QN5"}, b1, false},
		//{"Mqn6", args{b: b2, unit: "QN6"}, b2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RawToBalance(tt.args.b, tt.args.unit)
			if (err != nil) != tt.wantErr {
				t.Errorf("RawToBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RawToBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}
