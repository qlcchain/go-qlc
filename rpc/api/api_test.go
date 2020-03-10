package api

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
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

func addPovBlock(l *ledger.Ledger, prevBlock *types.PovBlock, height uint64) *types.PovBlock {
	pb, td := mock.GeneratePovBlock(prevBlock, 0)
	pb.Header.BasHdr.Height = height
	l.AddPovBlock(pb, td)
	l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
	return pb
}
