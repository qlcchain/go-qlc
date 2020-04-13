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

func TestNewPrivacyService(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	cfg, _ := cm.Config()
	cfg.Privacy.Enable = true
	cfg.Privacy.PtmNode = "unix:/tmp/__UnitTestCase__.ipc"
	_ = cm.CommitAndSave()

	ls := NewPrivacyService(cm.ConfigFile)

	err = ls.Init()
	if err != nil {
		t.Fatal(err)
	}
	if ls.State() != 2 {
		t.Fatal("privacy init failed")
	}
	_ = ls.Start()
	err = ls.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if ls.Status() != 6 {
		t.Fatal("privacy stop failed.")
	}
}
