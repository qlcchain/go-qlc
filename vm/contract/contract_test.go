package contract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract"
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

func TestDescribe_GetVersion(t *testing.T) {
	d := vmcontract.Describe{
		SpecVer:   1,
		Signature: true,
		Pending:   true,
		PovState:  true,
		Work:      true,
	}
	if d.GetVersion() != 1 {
		t.Fatal()
	}
	if !d.WithSignature() {
		t.Fatal()
	}

	if !d.WithPending() {
		t.Fatal()
	}
	if !d.WithPovState() {
		t.Fatal()
	}
	if !d.WithWork() {
		t.Fatal()
	}
}

func TestBaseContract_GetDescribe(t *testing.T) {
	// stupid
	b := &BaseContract{}
	_, _, _ = b.DoPending(nil)
	b.GetDescribe()
	_, _ = b.GetTargetReceiver(nil, nil)
	_, _ = b.GetFee(nil, nil)
	b.GetRefundData()
	_, _, _ = b.DoPending(nil)
	_, _, _ = b.ProcessSend(nil, nil)
	_, _, _ = b.DoGap(nil, nil)
	_ = b.DoReceiveOnPov(nil, nil, 0, nil, nil)
	_ = b.DoSend(nil, nil)
	_ = b.DoSendOnPov(nil, nil, 0, nil)
}
