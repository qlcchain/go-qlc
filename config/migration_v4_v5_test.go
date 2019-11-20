/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationV4ToV5_Migration(t *testing.T) {
	dir := filepath.Join(QlcTestDataDir(), "config")
	defer func() { _ = os.RemoveAll(dir) }()
	cfg4, err := DefaultConfigV4(dir)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(cfg4)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMigrationV4ToV5()
	migration, i, err := m.Migration(bytes, 4)
	if err != nil || i != 5 {
		t.Fatal("migration failed")
	}

	cfg5 := &ConfigV5{}
	err = json.Unmarshal(migration, cfg5)
	if err != nil {
		t.Fatal(err)
	}

	if !cfg5.PoV.PovEnabled {
		t.Fatal("pov enabled failed")
	}

	if cfg5.Metrics == nil {
		t.Fatal("metrics config failed")
	}

	if !cfg5.Metrics.Enable {
		t.Fatal("cfg5.Metrics.Enable failed")
	}

	//if cfg5.Version != 5 {
	//	t.Fatal("invalid cfg version", cfg5.Version)
	//}

	if cfg5.Manager == nil {
		t.Fatal("invalid manager")
	} else {
		t.Log(cfg5.Manager.AdminToken)
	}
}
