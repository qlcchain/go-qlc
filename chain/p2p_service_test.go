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

func TestNewP2PService(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	cfg, err := config.DefaultConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewP2PService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Init()
	if err != nil {
		t.Fatal(err)
	}
	if p.State() != 2 {
		t.Fatal("p2p init failed")
	}
	err = p.Start()
	if err != nil {
		t.Fatal(err)
	}
	if p.State() != 4 {
		t.Fatal("p2p start failed")
	}
	err = p.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if p.Status() != 6 {
		t.Fatal("stop failed.")
	}
}
