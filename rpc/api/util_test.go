/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
)

func setupUtilAPI(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *UtilAPI) {
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewUtilAPI(l)

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(mock.MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}

	for i := range blocks {
		block := blocks[i]
		if err := verifier.BlockProcess(block); err != nil {
			t.Fatal(err)
		}
	}

	return func(t *testing.T) {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
		_ = cc.Stop()
	}, verifier, api
}

func TestNewUtilApi(t *testing.T) {
	testcase, _, api := setupUtilAPI(t)
	defer testcase(t)

	raw := mock.Hash().String()
	pw := "qlcchain"

	if encrypt, err := api.Encrypt(raw, pw); err != nil {
		t.Fatal(err)
	} else {
		if decrypt, err := api.Decrypt(encrypt, pw); err != nil {
			t.Fatal(err)
		} else if decrypt != raw {
			t.Fatalf("invalid raw, exp: %s, act: %s", raw, pw)
		}
	}
	n := "QGAS"
	if raw, err := api.BalanceToRaw(types.Balance{Int: big.NewInt(100)}, "QGAS", &n); err != nil {
		t.Fatal(err)
	} else {
		t.Log(raw)
	}

	if raw, err := api.RawToBalance(types.Balance{Int: big.NewInt(100)}, "QGAS", &n); err != nil {
		t.Fatal(err)
	} else {
		t.Log(raw)
	}

	if raw, err := api.BalanceToRaw(types.Balance{Int: big.NewInt(100)}, "QLC", nil); err != nil {
		t.Fatal(err)
	} else {
		t.Log(raw)
	}

	if raw, err := api.RawToBalance(types.Balance{Int: big.NewInt(100)}, "QLC", nil); err != nil {
		t.Fatal(err)
	} else {
		t.Log(raw)
	}
}

func TestAPIBalance_MarshalText(t *testing.T) {
	b := &APIBalance{big.NewFloat(11)}
	if text, err := b.MarshalText(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(text))
	}
	t.Log(b.String())
}
