/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestGetRewardsDetail(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	vmContext := vmstore.NewVMContext(l)

	data := mockRewards(4, Rewards)

	for _, md := range data {
		if err := l.AddStateBlock(md.Block1); err != nil {
			t.Fatal(err)
		}
		if err := l.AddStateBlock(md.Block2); err != nil {
			t.Fatal(err)
		}
		key := GetRewardsKey(md.Param.Id[:], md.Param.TxHeader[:], md.Param.RxHeader[:])
		if data, err := RewardsABI.PackVariable(VariableNameRewards, md.Info.Type, md.Info.From, md.Info.To,
			md.Info.TxHeader, md.Info.RxHeader, md.Info.Amount); err == nil {
			if err := vmContext.SetStorage(types.RewardsAddress[:], key, data); err != nil {
				t.Error(err)
			} else {
				//t.Log(util.ToIndentString(md))
				//t.Log(hex.EncodeToString(key))
				//t.Log(hex.EncodeToString(data))
			}
		} else {
			t.Fatal(err)
		}
	}
	err := vmContext.SaveStorage()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("prepare data successful")

	txId := hex.EncodeToString(data[0].Param.Id[:])
	t.Log("query ", txId)
	infos, err := GetRewardsDetail(vmContext, txId)
	if err != nil {
		t.Fatal(err)
	} else {
		total := new(big.Int)
		for _, value := range infos {
			total.Add(total, value.Amount)
			t.Log("find ", util.ToIndentString(value))
		}

		rewards, err := GetTotalRewards(vmContext, txId)
		if err != nil {
			t.Fatal(err)
		}

		if rewards.Cmp(total) != 0 {
			t.Fatal("invalid total", rewards, total)
		}
	}
}

func TestGetConfidantDetail(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	vmContext := vmstore.NewVMContext(l)

	data := mockRewards(4, Confidant)

	for _, md := range data {
		if err := l.AddStateBlock(md.Block1); err != nil {
			t.Fatal(err)
		}
		if err := l.AddStateBlock(md.Block2); err != nil {
			t.Fatal(err)
		}
		key := GetConfidantKey(md.Param.Beneficial, md.Param.Id[:], md.Param.TxHeader[:], md.Param.RxHeader[:])
		if data, err := RewardsABI.PackVariable(VariableNameRewards, md.Info.Type, md.Info.From, md.Info.To,
			md.Info.TxHeader, md.Info.RxHeader, md.Info.Amount); err == nil {
			if err := vmContext.SetStorage(types.RewardsAddress[:], key, data); err != nil {
				t.Error(err)
			} else {
				t.Log(util.ToIndentString(md))
				t.Log(hex.EncodeToString(key))
				t.Log(hex.EncodeToString(data))
			}
		} else {
			t.Fatal(err)
		}
	}
	err := vmContext.SaveStorage()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("prepare data successful")
	rxAddress := data[0].Param.Beneficial
	t.Log("query ", rxAddress.String())
	infos, err := GetConfidantRewordsDetail(vmContext, rxAddress)
	if err != nil {
		t.Fatal(err)
	} else {
		for k, values := range infos {
			for idx, v := range values {
				t.Logf("%s: %d==>%v", k, idx, v)
			}
		}

		rewards, err := GetConfidantRewords(vmContext, rxAddress)
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log(util.ToIndentString(rewards))
		}
	}
}

type mockData struct {
	Param  *RewardsParam     `json:"Param"`
	Info   *RewardsInfo      `json:"info"`
	Block1 *types.StateBlock `json:"b1"`
	Block2 *types.StateBlock `json:"b2"`
}

func mockRewards(count int, t uint8) []*mockData {
	var result []*mockData
	txAddress := mock.Address()
	txIds := []types.Hash{mock.Hash(), mock.Hash()}

	for i := 0; i < count; i++ {
		sign := make([]byte, types.SignatureSize)
		_ = random.Bytes(sign)
		signature, _ := types.BytesToSignature(sign)
		i, _ := random.Intn(100)
		b1 := mock.StateBlockWithoutWork()
		b2 := mock.StateBlockWithoutWork()
		param := &RewardsParam{
			Id:         txIds[i%2],
			Beneficial: mock.Address(),
			TxHeader:   b1.GetHash(),
			RxHeader:   b2.GetHash(),
			Amount:     big.NewInt(int64(i + 1)),
			Sign:       signature,
		}

		info := &RewardsInfo{
			Type:     t,
			From:     txAddress,
			To:       param.Beneficial,
			TxHeader: param.TxHeader,
			RxHeader: param.RxHeader,
			Amount:   param.Amount,
		}

		result = append(result, &mockData{
			Param:  param,
			Info:   info,
			Block1: b1,
			Block2: b2,
		})
	}
	return result
}

func TestGetRewardsKey(t *testing.T) {
	id := mock.Hash()
	txHeader := mock.Hash()
	rxHeader := mock.Hash()
	key := GetRewardsKey(id[:], txHeader[:], rxHeader[:])

	//t.Logf("id: %v\ntxHeader:%v\nrxHeader:%v\nkey:%v", id[:], txHeader[:], rxHeader[:], key)
	len1 := len(key)
	len2 := len(id) + len(txHeader) + len(rxHeader)
	if len1 != len2 {
		t.Fatal("invalid len", len1, len2)
	}

	if !bytes.HasPrefix(key, id[:]) {
		t.Fatal("invalid key prefix")
	}

	if !bytes.HasSuffix(key, rxHeader[:]) {
		t.Fatal("invalid key suffix")
	}
}

func TestGetConfidantKey(t *testing.T) {
	addr := mock.Address()
	id := mock.Hash()
	txHeader := mock.Hash()
	rxHeader := mock.Hash()
	key := GetConfidantKey(addr, id[:], txHeader[:], rxHeader[:])

	//t.Logf("id: %v\ntxHeader:%v\nrxHeader:%v\nkey:%v", id[:], txHeader[:], rxHeader[:], key)
	len1 := len(key)
	len2 := len(id) + len(txHeader) + len(rxHeader) + types.AddressSize
	if len1 != len2 {
		t.Fatal("invalid len", len1, len2)
	}

	if !bytes.HasPrefix(key, addr[:]) {
		t.Fatal("invalid key addr")
	}

	if !bytes.EqualFold(key[types.AddressSize:types.AddressSize+types.HashSize], id[:]) {
		t.Fatal("invalid hash")
	}

	if !bytes.HasSuffix(key, rxHeader[:]) {
		t.Fatal("invalid key suffix")
	}
}

func TestParseRewardsInfo(t *testing.T) {
	tx, _ := hex.DecodeString("0d8e899aee3fd8acf707841e04463cf3d8f151cbe8870495febc5f47cfadfe1a")
	rx, _ := hex.DecodeString("52112688922647245083902739812993217025393153372347545601073710280271530289093")
	h1, _ := types.BytesToAddress(tx)
	h2, _ := types.BytesToHash(rx)
	a1, _ := types.HexToAddress("qlc_1kk5xst583y8hpn9c48ruizs5cxprdeptw6s5wm6ezz6i1h5srpz3mnjgxao")
	a2, _ := types.HexToAddress("qlc_3pj83yuemoegkn6ejskd8bustgunmfqpbhu3pnpa6jsdjf9isybzffwq7s4p")
	bytes, e := RewardsABI.PackVariable(VariableNameRewards, uint8(1), a1, a2, h1, h2, big.NewInt(83020000000))
	if e != nil {
		t.Fatal(e)
	} else {
		fmt.Println(bytes)
	}

	info, e := ParseRewardsInfo(bytes)
	if e != nil {
		t.Fatal(e)
	} else {
		t.Log(util.ToIndentString(info))
	}
}

func TestRewardsParam_Verify(t *testing.T) {
	a := mock.Account()
	param := &RewardsParam{
		Id:         mock.Hash(),
		Beneficial: mock.Address(),
		TxHeader:   mock.Hash(),
		RxHeader:   mock.Hash(),
		Amount:     big.NewInt(100),
		Sign:       types.Signature{},
	}

	if data, err := RewardsABI.PackMethod(MethodNameUnsignedConfidantRewards, param.Id, param.Beneficial, param.TxHeader,
		param.RxHeader, param.Amount); err == nil {
		param.Sign = a.Sign(types.HashData(data))
	} else {
		t.Fatal(err)
	}

	param2 := &RewardsParam{
		Id:         mock.Hash(),
		Beneficial: mock.Address(),
		TxHeader:   mock.Hash(),
		RxHeader:   mock.Hash(),
		Amount:     big.NewInt(200),
		Sign:       types.Signature{},
	}
	if data, err := RewardsABI.PackMethod(MethodNameUnsignedAirdropRewards, param2.Id, param2.Beneficial, param2.TxHeader,
		param2.RxHeader, param2.Amount); err == nil {
		param2.Sign = a.Sign(types.HashData(data))
	} else {
		t.Fatal(err)
	}

	type fields struct {
		Id         types.Hash
		Beneficial types.Address
		TxHeader   types.Hash
		RxHeader   types.Hash
		Amount     *big.Int
		Sign       types.Signature
	}
	type args struct {
		address    types.Address
		methodName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "MethodNameUnsignedConfidantRewards",
			fields: fields{
				Id:         param.Id,
				Beneficial: param.Beneficial,
				TxHeader:   param.TxHeader,
				RxHeader:   param.RxHeader,
				Amount:     param.Amount,
				Sign:       param.Sign,
			},
			args: args{
				address:    a.Address(),
				methodName: MethodNameUnsignedConfidantRewards,
			},
			want:    true,
			wantErr: false,
		}, {
			name: "MethodNameUnsignedAirdropRewards",
			fields: fields{
				Id:         param2.Id,
				Beneficial: param2.Beneficial,
				TxHeader:   param2.TxHeader,
				RxHeader:   param2.RxHeader,
				Amount:     param2.Amount,
				Sign:       param2.Sign,
			},
			args: args{
				address:    a.Address(),
				methodName: MethodNameUnsignedAirdropRewards,
			},
			want:    true,
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				Id:         types.Hash{},
				Beneficial: param2.Beneficial,
				TxHeader:   param2.TxHeader,
				RxHeader:   param2.RxHeader,
				Amount:     param2.Amount,
				Sign:       param2.Sign,
			},
			args: args{
				address:    a.Address(),
				methodName: MethodNameUnsignedAirdropRewards,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				Id:         param2.Id,
				Beneficial: types.ZeroAddress,
				TxHeader:   param2.TxHeader,
				RxHeader:   param2.RxHeader,
				Amount:     param2.Amount,
				Sign:       param2.Sign,
			},
			args: args{
				address:    a.Address(),
				methodName: MethodNameUnsignedAirdropRewards,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				Id:         param2.Id,
				Beneficial: param2.Beneficial,
				TxHeader:   param2.TxHeader,
				RxHeader:   param2.RxHeader,
				Amount:     param2.Amount,
				Sign:       types.ZeroSignature,
			},
			args: args{
				address:    a.Address(),
				methodName: MethodNameUnsignedAirdropRewards,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := &RewardsParam{
				Id:         tt.fields.Id,
				Beneficial: tt.fields.Beneficial,
				TxHeader:   tt.fields.TxHeader,
				RxHeader:   tt.fields.RxHeader,
				Amount:     tt.fields.Amount,
				Sign:       tt.fields.Sign,
			}
			got, err := ap.Verify(tt.args.address, tt.args.methodName)
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
