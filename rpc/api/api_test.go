package api

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"

	qcfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

func getTestLedger() (func(), *ledger.Ledger, string) {
	dir := filepath.Join(qcfg.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := qcfg.NewCfgManager(dir)
	_, _ = cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)

	return func() {
		l.Close()
		os.RemoveAll(dir)
	}, l, cm.ConfigFile
}
