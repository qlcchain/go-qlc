/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common/event"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/config"
)

func TestNewRPCService(t *testing.T) {
	eventBus := event.New()
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	cfg, err := config.DefaultConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	ls := NewRPCService(cfg, eventBus)
	err = ls.Init()
	if err != nil {
		t.Fatal(err)
	}
	if ls.State() != 2 {
		t.Fatal("rpc init failed")
	}
	err = ls.Start()
	if err != nil {
		t.Fatal(err)
	}
	if ls.State() != 4 {
		t.Fatal("rpc start failed")
	}
	err = ls.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if ls.Status() != 6 {
		t.Fatal("stop failed.")
	}
}
