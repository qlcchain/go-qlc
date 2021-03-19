package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationV10ToV11_Migration(t *testing.T) {
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
	m := NewMigrationV10ToV11()
	migration, i, err := m.Migration(bytes, 10)
	if err != nil || i != 11 {
		t.Fatal("migration failed")
	}

	cfg11 := &Config{}
	err = json.Unmarshal(migration, cfg11)
	if err != nil {
		t.Fatal(err)
	}

	if m.StartVersion() != 10 {
		t.Fatal("start version error")
	}
	if m.EndVersion() != 11 {
		t.Fatal("end version error")
	}
}
