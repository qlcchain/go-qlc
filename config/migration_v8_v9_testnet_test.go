// +build  testnet

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationV8ToV9_Migration(t *testing.T) {
	dir := filepath.Join(QlcTestDataDir(), "config")
	defer func() { _ = os.RemoveAll(dir) }()
	cfg8, err := DefaultConfigV8(dir)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(cfg8)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMigrationV8ToV9()
	migration, i, err := m.Migration(bytes, 8)
	if err != nil || i != 9 {
		t.Fatal("migration failed")
	}

	cfg9 := &Config{}
	err = json.Unmarshal(migration, cfg9)
	if err != nil {
		t.Fatal(err)
	}

	if m.startVersion != 8 {
		t.Fatal("start version error")
	}
	if m.endVersion != 9 {
		t.Fatal("end version error")
	}

	found := false
	for _, pm := range cfg9.RPC.PublicModules {
		if pm == "KYC" {
			found = true
		}
	}

	if !found {
		t.Fatal()
	}
}
