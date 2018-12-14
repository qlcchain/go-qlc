/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"github.com/qlcchain/go-qlc/config"
	"os"
	"path/filepath"
	"testing"
)

var (
	store *WalletStore
	dir   = filepath.Join(config.QlcTestDataDir(), "wallet")
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup store test case", dir)
	_ = os.RemoveAll(dir)
	store = NewWalletStore(dir)
	if store == nil {
		t.Fatal("create store failed")
	}
	return func(t *testing.T) {
		t.Log("teardown wallet test case")
		err := store.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}
}
