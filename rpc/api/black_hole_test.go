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
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

func setupBlackHoleAPI(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *BlackHoleAPI) {
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewBlackHoleApi(l, cc)

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(MockBlocks), &blocks); err != nil {
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

func TestBlackHoleAPI_GetSendBlock(t *testing.T) {
	testcase, verifier, api := setupBlackHoleAPI(t)
	defer testcase(t)

	a1 := account1.Address()
	tm, err := api.l.GetTokenMeta(a1, config.GasToken())
	if err != nil {
		t.Fatal(err)
	}

	param := &cabi.DestroyParam{
		Owner:    a1,
		Previous: tm.Header,
		Token:    config.GasToken(),
		Amount:   big.NewInt(100),
	}
	if param.Sign, err = param.Signature(account1); err != nil {
		t.Fatal(err)
	}

	if blk, err := api.GetSendBlock(param); err != nil {
		t.Fatal(err)
	} else {
		txHash := blk.GetHash()
		blk.Signature = account1.Sign(txHash)
		if err := verifier.BlockProcess(blk); err != nil {
			t.Fatal(err)
		}

		if blk, err := api.GetRewardsBlock(&txHash); err != nil {
			t.Fatal(err)
		} else {
			txHash := blk.GetHash()
			blk.Signature = account1.Sign(txHash)
			if err := verifier.BlockProcess(blk); err != nil {
				t.Fatal(err)
			}
		}

		if details, err := api.GetDestroyInfoDetail(&a1); err != nil {
			t.Fatal(err)
		} else if len(details) != 1 {
			t.Fatalf("invalid details len, exp: 1, act: %d", len(details))
		} else {
			t.Log(util.ToString(details))
		}

		if info, err := api.GetTotalDestroyInfo(&a1); err != nil {
			t.Fatal(err)
		} else if info.Compare(types.ZeroBalance) == types.BalanceCompEqual {
			t.Fatal("invalid balance")
		} else {
			t.Log(info)
		}
	}
}
