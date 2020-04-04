package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
)

func TestNewDB(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "relation", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cfg, _ := cm.Load()
	db, err := NewDB(cfg)
	if err != nil {
		t.Fatal(err)
	}

	//err := l.DBStore.Erase()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	//CloseLedger()
	err = os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
}
