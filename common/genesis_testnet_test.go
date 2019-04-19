// +build  testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func TestGenesisBlock2(t *testing.T) {
	h, _ := types.NewHash("5594c690c3618a170a77d2696688f908efec4da2b94363fcb96749516307031d")

	h2 := testGenesisBlock.GetHash()
	if h2 != h {
		t.Log(util.ToString(testGenesisBlock))
		t.Fatal("invalid genesis block", h2.String(), h.String())
	}

	h3, _ := types.NewHash("8b54787c668dddd4f22ad64a8b0d241810871b9a52a989eb97670f345ad5dc90")
	h4 := testGenesisMintageBlock.GetHash()
	if h3 != h4 {
		t.Log(util.ToIndentString(testGenesisMintageBlock))
		t.Fatal("invalid genesis mintage block", h3.String(), h4.String())
	}
}

func TestGasBlock2(t *testing.T) {
	h, _ := types.NewHash("424b367da2e0ff991d3086f599ce26547b80ae948b209f1cb7d63e19231ab213")

	h2 := testGasBlock.GetHash()
	if h2 != h {
		t.Log(util.ToString(testGasBlock))
		t.Fatal("invalid gas block", h2.String(), h.String())
	}

	h3, _ := types.NewHash("f798089896ffdf45ccce2e039666014b8c666ea0f47f0df4ee7e73b49dac0945")
	h4 := testGasMintageBlock.GetHash()
	if h3 != h4 {
		t.Log(util.ToIndentString(testGasMintageBlock))
		t.Fatal("invalid gas mintage block", h3.String(), h4.String())
	}
}

func TestIsGenesisToken(t *testing.T) {
	h1, _ := types.NewHash("327531148b1a6302632aa7ad6eb369437d8269a08a55b344bd06b514e4e6ae97")
	h2, _ := types.NewHash("89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4")
	b1 := IsGenesisToken(h1)
	if b1 {
		t.Fatal("h1 should not be Genesis Token")
	}
	b2 := IsGenesisToken(h2)
	if !b2 {
		t.Fatal("h2 should be Genesis Token")
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
