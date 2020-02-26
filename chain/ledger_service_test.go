/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
)

func TestNewLedgerService(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	cfg, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	ls := NewLedgerService(cm.ConfigFile)
	for _, v := range cfg.Genesis.GenesisBlocks {
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: v.Mintage,
			GenesisBlock:        v.Genesis,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
	}
	err = ls.Init()
	if err != nil {
		t.Fatal(err)
	}
	if ls.State() != 2 {
		t.Fatal("ledger init failed")
	}
	_ = ls.Start()
	err = ls.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if ls.Status() != 6 {
		t.Fatal("stop failed.")
	}
}
