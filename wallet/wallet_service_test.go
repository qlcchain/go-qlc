/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"testing"

	"github.com/qlcchain/go-qlc/config"
)

func TestWalletService_Init(t *testing.T) {
	cfg, err := config.DefaultConfig(config.DefaultDataDir())
	if err != nil {
		t.Fatal(err)
	}
	s := NewWalletService(cfg)
	err = s.Init()
	if err != nil {
		t.Fatal(err)
	}
	if s.State() != 2 {
		t.Fatal("ledger init failed")
	}
	_ = s.Start()
	err = s.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if s.Status() != 6 {
		t.Fatal("stop failed.")
	}
}
