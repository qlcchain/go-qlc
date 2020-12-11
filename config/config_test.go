package config

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/util"
)

func TestConfig_Dir(t *testing.T) {
	cfg, err := DefaultConfigV10(DefaultDataDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.ToIndentString(cfg))
	t.Log(cfg.DataDir)
	c := Config(*cfg)
	t.Log(c.LogDir())
	t.Log(c.LedgerDir())
	t.Log(c.WalletDir())
	t.Log(QlcTestDataDir())
}
