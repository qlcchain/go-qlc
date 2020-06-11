package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationV7ToV8_Migration(t *testing.T) {
	dir := filepath.Join(QlcTestDataDir(), "config")
	defer func() { _ = os.RemoveAll(dir) }()
	cfg7, err := DefaultConfigV7(dir)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(cfg7)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMigrationV7ToV8()
	migration, i, err := m.Migration(bytes, 7)
	if err != nil || i != 8 {
		t.Fatal("migration failed")
	}

	cfg8 := &Config{}
	err = json.Unmarshal(migration, cfg8)
	if err != nil {
		t.Fatal(err)
	}
	if cfg8.RPC.GRPCConfig == nil {
		t.Fatal("grpc config is nil")
	}
	if m.startVersion != 7 {
		t.Fatal("start version error")
	}
	if m.endVersion != 8 {
		t.Fatal("end version error")
	}
}
