/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	_ "net/http/pprof"
	"testing"

	"github.com/qlcchain/go-qlc/config"
)

func Test_updateConfig(t *testing.T) {
	cfgPathP = config.DefaultDataDir()
	configParamsP = []string{"rpc.rpcEnabled=true", "rpc.httpCors=localhost,localhost2", "p2p.syncInterval=200", "rpc.rpcEnabled="}
	v3, err := config.DefaultConfig(cfgPathP)
	if err != nil {
		t.Fatal(err)
	}

	err = updateConfig(v3)
	if err != nil {
		t.Fatal(err)
	}
	if !v3.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	for len(v3.RPC.HTTPCors) != 2 || v3.RPC.HTTPCors[0] != "localhost" || v3.RPC.HTTPCors[1] != "localhost2" {
		t.Fatal("invalid rpc.httpCors", v3.RPC.HTTPCors)
	}

	if v3.P2P.SyncInterval != 200 {
		t.Fatal("invalid p2p.syncInterval", v3.P2P.SyncInterval)
	}
}
