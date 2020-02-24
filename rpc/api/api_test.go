package api

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	qcfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

func getTestLedger() (func(), *ledger.Ledger, string) {
	dir := filepath.Join(qcfg.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := qcfg.NewCfgManager(dir)
	cfg, _ := cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)

	for _, v := range cfg.Genesis.GenesisBlocks {
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: v.Mintage,
			GenesisBlock:        v.Genesis,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
	}

	return func() {
		l.Close()
		os.RemoveAll(dir)
	}, l, cm.ConfigFile
}
