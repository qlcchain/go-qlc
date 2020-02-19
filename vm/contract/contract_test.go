package contract

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	qcfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"os"
	"path/filepath"
)

func getTestLedger() (func(), *ledger.Ledger) {
	dir := filepath.Join(qcfg.QlcTestDataDir(), "contract", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := qcfg.NewCfgManager(dir)
	cfg, _ := cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)

	var mintageBlock, genesisBlock types.StateBlock
	for _, v := range cfg.Genesis.GenesisBlocks {
		_ = json.Unmarshal([]byte(v.Genesis), &genesisBlock)
		_ = json.Unmarshal([]byte(v.Mintage), &mintageBlock)
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: mintageBlock,
			GenesisBlock:        genesisBlock,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
	}

	return func() {
		l.Close()
		os.RemoveAll(dir)
	}, l
}

func addLatestPovBlock(pb *types.PovBlock, td *types.PovTD, l *ledger.Ledger) error {
	err := l.AddPovBlock(pb, td)
	if err != nil {
		return err
	}

	err = l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	if err != nil {
		return err
	}

	err = l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
	if err != nil {
		return err
	}

	return nil
}
