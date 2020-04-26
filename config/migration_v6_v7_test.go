// +build !testnet

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationV6ToV7_Migration(t *testing.T) {
	dir := filepath.Join(QlcTestDataDir(), "config")
	defer func() { _ = os.RemoveAll(dir) }()
	cfg6, err := DefaultConfigV6(dir)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(cfg6)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMigrationV6ToV7()
	migration, i, err := m.Migration(bytes, 6)
	if err != nil || i != 7 {
		t.Fatal("migration failed")
	}

	cfg7 := &Config{}
	err = json.Unmarshal(migration, cfg7)
	if err != nil {
		t.Fatal(err)
	}
	if cfg7.WhiteList.Enable {
		t.Fatal("default whitelist is false")
	}
}
