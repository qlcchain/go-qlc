// +build !testnet

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationV8ToV10_Migration(t *testing.T) {
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
	m := NewMigrationV8ToV10()
	migration, i, err := m.Migration(bytes, 8)
	if err != nil || i != 10 {
		t.Fatal("migration failed")
	}

	cfg10 := &Config{}
	err = json.Unmarshal(migration, cfg10)
	if err != nil {
		t.Fatal(err)
	}
	if cfg10.TrieClean == nil {
		t.Fatal("trie clean config is nil")
	}
}
