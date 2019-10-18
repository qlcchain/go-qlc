/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/hex"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger) {
	t.Parallel()

	dir := filepath.Join(cfg.QlcTestDataDir(), "rewards", uuid.New().String())
	_ = os.RemoveAll(dir)

	cm := cfg.NewCfgManager(dir)
	cm.Load()
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

func TestRewardsApi_GetRewardData(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	api := NewRewardsApi(l)

	tx := mock.Account()
	rx := mock.Account()

	//mock tx data
	am := mock.AccountMeta(tx.Address())
	if err := l.AddAccountMeta(am); err == nil {
		token := mock.TokenMeta(tx.Address())
		token.Type = common.GasToken()
		if err := l.AddTokenMeta(token.BelongTo, token); err == nil {

		} else {
			t.Error(err)
		}
	} else {
		t.Error(err)
	}

	//mock rx data
	rxAm := mock.AccountMeta(rx.Address())
	if err := l.AddAccountMeta(rxAm); err == nil {

	} else {
		t.Error(err)
	}

	id := mock.Hash()
	param := &RewardsParam{
		Id:     hex.EncodeToString(id[:]),
		Amount: types.Balance{Int: big.NewInt(10e7)},
		Self:   tx.Address(),
		To:     rx.Address(),
	}
	hash, err := api.GetUnsignedRewardData(param)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(hash.String())
	}
	sign := tx.Sign(hash)
	t.Log(tx.Address(), "==>", sign.String())

	blk, err := api.GetSendRewardBlock(param, &sign)
	if err != nil {
		t.Fatal(err)
	} else {
		//t.Log(util.ToIndentString(blk))
	}

	t.Log(util.ToIndentString(param))
	t.Log(sign.String())

	if isAirdropRewards := api.IsAirdropRewards(blk.Data); !isAirdropRewards {
		t.Fatal("invalid data type")
	}

	vmContext := vmstore.NewVMContext(l)
	err = api.rewards.DoSend(vmContext, blk)
	if err != nil {
		t.Fatal(err)
	}

	key, info, err := api.rewards.DoPending(blk)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(key, info)
	}

	//verifier := process.NewLedgerVerifier(api.ledger)
	//err = verifier.BlockProcess(blk)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//hash = blk.GetHash()
	//
	//recv, err := api.GetReceiveRewardBlock(&hash)
	//if err != nil {
	//	t.Fatal(err)
	//} else {
	//	t.Log(recv)
	//}
}

func TestRewardsApi_GetConfidantRewordsRewardData(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	api := NewRewardsApi(l)

	tx := mock.Account()
	rx := mock.Account()

	//mock tx data
	am := mock.AccountMeta(tx.Address())
	if err := l.AddAccountMeta(am); err == nil {
		token := mock.TokenMeta(tx.Address())
		token.Type = common.GasToken()
		if err := l.AddTokenMeta(token.BelongTo, token); err == nil {

		} else {
			t.Error(err)
		}
	} else {
		t.Error(err)
	}

	//mock rx data
	rxAm := mock.AccountMeta(rx.Address())
	if err := l.AddAccountMeta(rxAm); err == nil {

	} else {
		t.Error(err)
	}

	param := &RewardsParam{
		Id:     mock.NewRandomMac().String(),
		Amount: types.Balance{Int: big.NewInt(10e7)},
		Self:   tx.Address(),
		To:     rx.Address(),
	}
	hash, err := api.GetUnsignedConfidantData(param)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(hash.String())
	}
	sign := tx.Sign(hash)
	t.Log(sign.String())

	blk, err := api.GetSendConfidantBlock(param, &sign)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(util.ToIndentString(blk))
	}

	if isAirdrop := api.IsAirdropRewards(blk.Data); isAirdrop {
		t.Fatal("invalid block data")
	}

	vmContext := vmstore.NewVMContext(l)
	err = api.confidantRewards.DoSend(vmContext, blk)
	if err != nil {
		t.Fatal(err)
	}
}
