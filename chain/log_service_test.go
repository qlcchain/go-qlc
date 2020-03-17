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

func TestLogService_Start(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	ls := NewLogService(cm.ConfigFile)

	if err = ls.Init(); err != nil {
		t.Fatal(err)
	}

	if ls.State() != 2 {
		t.Fatal("logger init failed")
	}
	if err = ls.Start(); err != nil {
		t.Fatal(err)
	}
	//if err = ls.Stop(); err != nil {
	//	t.Fatal(err)
	//}

	_ = ls.Stop()

	if ls.Status() != 6 {
		t.Fatal("stop failed.")
	}
}
