/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package process

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"os"
	"path/filepath"
	"testing"
)

var bc []*types.StateBlock

func setupTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger, *LedgerVerifier) {
	//t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	//dir := filepath.Join(config.DefaultDataDir()) // if want to test rollback contract and remove time sleep

	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)
	bc, _ = mock.BlockChain(false)

	return func(t *testing.T) {
		//err := l.db.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l, NewLedgerVerifier(l)
}

func TestProcess_BlockProcess(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	if err := lv.BlockProcess(bc[0]); err != nil {
		t.Fatal(err)
	}
	t.Log("bc hash", bc[0].GetHash())
	for i, b := range bc[1:] {
		fmt.Println(i + 1)
		fmt.Println("bc.previous", b.GetPrevious())
		if p, err := lv.Process(b); err != nil || p != Progress {
			t.Fatal(p, err)
		}
	}
}

func TestProcess_ContractBlockProcess(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	bs := mock.ContractBlocks()
	if err := lv.BlockProcess(bs[0]); err != nil {
		t.Fatal(err)
	}
	for i, b := range bs[1:] {
		fmt.Println(i)
		if r, err := lv.BlockCheck(b); r != Progress || err != nil {
			t.Fatal(r, err)
		}
		if err := lv.BlockProcess(b); err != nil {
			t.Fatal(Other, err)
		}
	}
}

func TestProcess_Exception(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	bc[0].Signature, _ = types.NewSignature("5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600")
	if r, err := lv.BlockCheck(bc[0]); err != nil || r != BadSignature {
		t.Fatal(r, err)
	}
	if r, err := lv.BlockCheck(bc[1]); err != nil || r != GapPrevious {
		t.Fatal(r, err)
	}

	if err := lv.BlockProcess(bc[0]); err != nil {
		t.Fatal(err)
	}
	if r, err := lv.BlockCheck(bc[0]); err != nil || r != Old {
		t.Fatal(r, err)
	}
	if r, err := lv.BlockCheck(bc[2]); err != nil || r != GapSource {
		t.Fatal(r, err)
	}

	// unReceivable
	if err := lv.BlockProcess(bc[1]); err != nil {
		t.Fatal(err)
	}
	if err := lv.l.DeletePending(&types.PendingKey{
		Address: bc[2].Address,
		Hash:    bc[1].GetHash(),
	}, lv.l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	if r, err := lv.BlockCheck(bc[2]); err != nil || r != UnReceivable {
		t.Fatal(r, err)
	}

	// contract block
	bc := mock.StateBlockWithoutWork()
	bc.Type = types.ContractReward
	if r, err := lv.BlockCheck(bc); err != nil || r != GapSource {
		t.Fatal(r, err)
	}
	bc.Type = types.ContractSend
	bc.Link = types.NEP5PledgeAddress.ToHash()
	if r, err := lv.BlockCheck(bc); err != nil || r != BadWork {
		t.Fatal(r, err)
	}

	bs := mock.ContractBlocks()
	if r, err := lv.BlockCheck(bs[1]); err != nil || r != GapPrevious {
		t.Fatal(r, err)
	}
}
