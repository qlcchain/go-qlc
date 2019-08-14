/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/config"
)

func TestWalletService_Init(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	s := NewWalletService(cm.ConfigFile)
	err = s.Init()
	if err != nil {
		t.Fatal(err)
	}
	if s.State() != 2 {
		t.Fatal("wallet init failed")
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
