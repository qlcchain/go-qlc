/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"encoding/hex"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "rewards", uuid.New().String())
	_ = os.RemoveAll(dir)
	l := ledger.NewLedger(dir)

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

func TestGetRewardsDetail(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	vmContext := vmstore.NewVMContext(l)

	data := mockRewards(4, Rewards)

	for _, md := range data {
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
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	vmContext := vmstore.NewVMContext(l)

	data := mockRewards(4, Confidant)

	for _, md := range data {
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
	Param *RewardsParam `json:"Param"`
	Info  *RewardsInfo  `json:"info"`
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
		param := &RewardsParam{
			Id:         txIds[i%2],
			Beneficial: mock.Address(),
			TxHeader:   mock.Hash(),
			RxHeader:   mock.Hash(),
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
			Param: param,
			Info:  info,
		})
	}
	return result
}
