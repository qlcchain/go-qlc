/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/types"
	qcfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

func getTestLedger() (func(), *ledger.Ledger) {
	dir := filepath.Join(qcfg.QlcTestDataDir(), "contract", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := qcfg.NewCfgManager(dir)
	_, _ = cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)

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
