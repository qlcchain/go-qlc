/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package process

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger, *LedgerVerifier) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	l := ledger.NewLedger(dir)

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

var bc, _ = mock.BlockChain()

//func TestProcess_BlockCheck(t *testing.T) {
//	teardownTestCase, l, lv := setupTestCase(t)
//	defer teardownTestCase(t)
//	blockCheck(t, lv)
//	checkInfo(t, l)
//}

func TestLedger_Rollback(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)
	if err := lv.BlockProcess(bc[0]); err != nil {
		t.Fatal(err)
	}
	for i, b := range bc[1:] {
		fmt.Println(i + 1)
		if p, err := lv.Process(b); err != nil || p != Progress {
			t.Fatal(p, err)
		}
	}

	h := bc[5].GetHash()
	if err := l.Rollback(h); err != nil {
		t.Fatal(err)
	}
	checkInfo(t, l)
}

func checkInfo(t *testing.T, l *ledger.Ledger) {
	addrs := make(map[types.Address]int)
	fmt.Println("----blocks----")
	err := l.GetStateBlocks(func(block *types.StateBlock) error {
		fmt.Println(block)
		if _, ok := addrs[block.GetAddress()]; !ok {
			addrs[block.GetAddress()] = 0
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(addrs)
	fmt.Println("----frontiers----")
	fs, _ := l.GetFrontiers()
	for _, f := range fs {
		fmt.Println(f)
	}

	fmt.Println("----account----")
	for k, _ := range addrs {
		ac, _ := l.GetAccountMeta(k)
		fmt.Println("   account", ac.Address)
		for _, token := range ac.Tokens {
			fmt.Println("      token, ", token)
		}
	}

	fmt.Println("----representation----")
	for k, _ := range addrs {
		b, err := l.GetRepresentation(k)
		if err != nil {
			if err == ledger.ErrRepresentationNotFound {
			}
		} else {
			fmt.Println(k, b)
		}
	}
}
