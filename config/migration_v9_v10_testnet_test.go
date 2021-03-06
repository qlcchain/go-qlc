// +build testnet

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationV9ToV10_Migration(t *testing.T) {
	dir := filepath.Join(QlcTestDataDir(), "config")
	defer func() { _ = os.RemoveAll(dir) }()
	cfg9, err := DefaultConfigV9(dir)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(cfg9)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMigrationV9ToV10()
	migration, i, err := m.Migration(bytes, 9)
	if err != nil || i != 10 {
		t.Fatal("migration failed")
	}

	cfg10 := &Config{}
	err = json.Unmarshal(migration, cfg10)
	if err != nil {
		t.Fatal(err)
	}

	if m.StartVersion() != 9 {
		t.Fatal("start version error")
	}
	if m.EndVersion() != 10 {
		t.Fatal("end version error")
	}

	if cfg10.DBOptimize == nil {
		t.Fatal("trie clean config is nil")
	}
}
