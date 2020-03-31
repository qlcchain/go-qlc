/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestGetTokenById(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
		teardownTestCase(t)
	}()

	ctx := vmstore.NewVMContext(l, &contractaddress.MintageAddress)

	a := mock.Address()
	h := mock.Hash()
	b := mock.Address()
	tokenID := NewTokenHash(a, h, "TEST")
	nep5TxId := random.RandomHexString(32)

	if data, err := MintageABI.PackVariable(
		VariableNameToken, tokenID, "TEST", "TTT", big.NewInt(100), uint8(8), b,
		big.NewInt(100), time.Now().AddDate(0, 0, 5).Unix(),
		a, nep5TxId); err != nil {
		t.Fatal(err)
	} else {
		if err := ctx.SetStorage(contractaddress.MintageAddress[:], []byte(nep5TxId), data); err != nil {
			t.Fatal(err)
		}

		if info, err := ParseTokenInfo(data); err != nil {
			t.Fatal(err)
		} else {
			t.Log(info)
		}
		gb := config.GenesisBlock()
		if err := ctx.SetStorage(contractaddress.MintageAddress[:], gb.Token[:], gb.Data); err != nil {
			t.Fatal(err)
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	if ti, err := GetTokenById(ctx, tokenID); err != nil {
		t.Fatal(err)
	} else if ti == nil {
		t.Fatal()
	}

	if ti, err := GetTokenByName(ctx, "TEST"); err != nil {
		t.Fatal(err)
	} else if ti == nil {
		t.Fatal()
	}

	if tokens, err := ListTokens(ctx); err != nil {
		t.Fatal(err)
	} else {
		for _, token := range tokens {
			t.Log(token)
		}
	}

}

func TestParseGenesisTokenInfo(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	block := config.GenesisBlock()
	if info, err := ParseGenesisTokenInfo(block.Data); err != nil {
		t.Fatal(err)
	} else {
		t.Log(info)
	}
}
