package config

import "testing"

func TestConfig_LogDir(t *testing.T) {
	cfg, err := DefaultConfig(DefaultDataDir())
	if err != nil {
		t.Fatal(err)
	}

	t.Log(cfg.DataDir)
	t.Log(cfg.LogDir())
	t.Log(cfg.LedgerDir())
	t.Log(cfg.WalletDir())
	t.Log(QlcTestDataDir())
}
