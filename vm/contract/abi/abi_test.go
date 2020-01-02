package abi

import (
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"os"
	"path/filepath"
	"testing"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "abi", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)

	return func(t *testing.T) {
		//err := l.Store.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l
}
