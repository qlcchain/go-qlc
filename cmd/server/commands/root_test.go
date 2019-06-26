/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/config"
)

func Test_updateConfig(t *testing.T) {
	cfgPathP = filepath.Join(config.QlcTestDataDir(), "config")
	configParamsP = "rpc.rpcEnabled=true:rpc.httpCors=localhost,localhost2:p2p.syncInterval=200:rpc.rpcEnabled="
	manager := config.NewCfgManager(cfgPathP)
	defer func() {
		_ = os.RemoveAll(cfgPathP)
	}()
	cfg, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfgPathN := filepath.Join(cfgPathP, config.QlcConfigFile)
	err = updateConfig(cfg, cfgPathN)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	for len(cfg.RPC.HTTPCors) != 2 || cfg.RPC.HTTPCors[0] != "localhost" || cfg.RPC.HTTPCors[1] != "localhost2" {
		t.Fatal("invalid rpc.httpCors", cfg.RPC.HTTPCors)
	}

	if cfg.P2P.SyncInterval != 200 {
		t.Fatal("invalid p2p.syncInterval", cfg.P2P.SyncInterval)
	}
}
