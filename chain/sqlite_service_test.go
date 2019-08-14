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

func TestSqliteService(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	ls, err := NewSqliteService(cm.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	err = ls.Init()
	if err != nil {
		t.Fatal(err)
	}
	if ls.State() != 2 {
		t.Fatal("service init failed")
	}
	err = ls.Start()
	if err != nil {
		t.Fatal(err)
	}
	if ls.State() != 4 {
		t.Fatal("service start failed")
	}
	err = ls.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if ls.Status() != 6 {
		t.Fatal("stop failed.")
	}
}
