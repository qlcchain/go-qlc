package config

import (
	"github.com/qlcchain/go-qlc/common/util"
	"testing"
)

func TestConfig_LogDir(t *testing.T) {
	cfg, err := DefaultConfigV1(DefaultDataDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.ToIndentString(cfg))

	t.Log(cfg.DataDir)
	//t.Log(cfg.LogDir())
	//t.Log(cfg.LedgerDir())
	//t.Log(cfg.WalletDir())
	t.Log(QlcTestDataDir())
}
