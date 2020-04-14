/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/ledger/process"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/types"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

var (
	account1 = mock.Account1
	account2 = mock.Account2
)

func setupLedgerForTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger) {
	t.Parallel()

	dir := filepath.Join(cfg.QlcTestDataDir(), "contract", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := cfg.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	l := ledger.NewLedger(cm.ConfigFile)

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(mock.MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}
	verifier := process.NewLedgerVerifier(l)
	for i := range blocks {
		block := blocks[i]
		if err := verifier.BlockProcess(block); err != nil {
			t.Fatal(err)
		}
		//if err := updateBlock(l, block); err != nil {
		//	t.Fatal(err)
		//}
	}

	return func(t *testing.T) {
		//err := l.DBStore.Erase()
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
