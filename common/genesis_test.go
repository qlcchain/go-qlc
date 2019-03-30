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

func TestGenesisBlock(t *testing.T) {
	h, _ := types.NewHash("8858a2b2563f6a702690beb4f29b61a88fc5a56dc50f3a26f2c97db1bf99c114")

	h2 := genesisBlock.GetHash()
	if h2 != h {
		t.Log(util.ToString(genesisBlock))
		t.Fatal("invalid genesis block", h2.String(), h.String())
	}

	h3, _ := types.NewHash("90f28436423396887ccb08362b62061ca4b3c5a297a84e30f405e8973f652484")
	h4 := genesisMintageBlock.GetHash()
	if h3 != h4 {
		t.Log(util.ToIndentString(genesisMintageBlock))
		t.Fatal("invalid genesis mintage block", h3.String(), h4.String())
	}
}

func TestGenesisBlock2(t *testing.T) {
	h, _ := types.NewHash("b14e95d66841ea82f77d5293a1e477691fe66e9c1a68db92d2bb040a2b67ba71")

	h2 := testGenesisBlock.GetHash()
	if h2 != h {
		t.Log(util.ToString(testGenesisBlock))
		t.Fatal("invalid genesis block", h2.String(), h.String())
	}

	h3, _ := types.NewHash("67513e803863279bc62d8e49a087b623895c8e2b21160a874f337ce147c859f1")
	h4 := testGenesisMintageBlock.GetHash()
	if h3 != h4 {
		t.Log(util.ToIndentString(testGenesisMintageBlock))
		t.Fatal("invalid genesis mintage block", h3.String(), h4.String())
	}
}

func TestGasBlock1(t *testing.T) {
	h, _ := types.NewHash("f72abf493f9b9378e67d4b24aa470be5a2cb71c22cf8d1a60e52b1f0e222a5d4")

	h2 := gasBlock.GetHash()
	if h2 != h {
		t.Log(util.ToString(testGasBlock))
		t.Fatal("invalid gas block", h2.String(), h.String())
	}

	h3, _ := types.NewHash("3b8c3acfbef2a93d9ba506073976f293cc1cca98892b7c545603945dd78f824f")
	h4 := gasMintageBlock.GetHash()
	if h3 != h4 {
		t.Log(util.ToIndentString(testGasMintageBlock))
		t.Fatal("invalid gas mintage block", h3.String(), h4.String())
	}
}

func TestGasBlock2(t *testing.T) {
	h, _ := types.NewHash("10043836573fdc1a4250913008c844a3572c2724ccc813e87bc2c341814d1afd")

	h2 := testGasBlock.GetHash()
	if h2 != h {
		t.Log(util.ToString(testGasBlock))
		t.Fatal("invalid gas block", h2.String(), h.String())
	}

	h3, _ := types.NewHash("327531148b1a6302632aa7ad6eb369437d8269a08a55b344bd06b514e4e6ae97")
	h4 := testGasMintageBlock.GetHash()
	if h3 != h4 {
		t.Log(util.ToIndentString(testGasMintageBlock))
		t.Fatal("invalid gas mintage block", h3.String(), h4.String())
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
