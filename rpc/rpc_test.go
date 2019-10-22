package rpc

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	jsonrpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *RPC) {
	rpcDir := filepath.Join(config.QlcTestDataDir(), "rpc")
	var rpc *RPC

	cfg, _ := config.DefaultConfig(rpcDir)
	cfg.RPC = &config.RPCConfigV2{
		Enable:       true,
		HTTPEndpoint: "tcp4://0.0.0.0:9735",
		WSEndpoint:   "tcp4://0.0.0.0:9736",
		//IPCEndpoint:  defaultIPCEndpoint(filepath.Join(rpcDir, "qlc_test.ipc")),
		WSEnabled:   true,
		IPCEnabled:  true,
		HTTPEnabled: true,
	}

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cm.Load()
	var err error
	rpc, err = NewRPC(cm.ConfigFile)
	rpc.config.RPC = &config.RPCConfigV2{
		Enable:       true,
		HTTPEndpoint: "tcp4://0.0.0.0:19735",
		WSEndpoint:   "tcp4://0.0.0.0:19736",
		//IPCEndpoint:  defaultIPCEndpoint(filepath.Join(rpcDir, "qlc_test.ipc")),
		WSEnabled:   true,
		IPCEnabled:  true,
		HTTPEnabled: true,
	}
	if err != nil {
		t.Fatal(err)
	}
	err = rpc.StartRPC()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("rpc started")
	return func(t *testing.T) {
		rpc.StopRPC()
		rpc.ledger.Close()
		_ = rpc.wallet.Close()
		err := os.RemoveAll(rpcDir)
		if err != nil {
			t.Fatal(err)
		}
	}, rpc
}

func TestRPC_HTTP(t *testing.T) {
	teardownTestCase, rpc := setupTestCase(t)
	defer teardownTestCase(t)

	_, address, _ := scheme(rpc.config.RPC.HTTPEndpoint)
	client, err := jsonrpc.Dial(fmt.Sprintf("http://%s", address))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	addr := mock.Address()
	var resp bool
	err = client.Call(&resp, "account_validate", addr)
	if err != nil {
		t.Fatal(err)
	}
	if !resp {
		t.Fatal()
	}

	blk := new(types.StateBlock)
	blk.Token = common.ChainToken()
	rpc.ledger.AddStateBlock(blk)
	var resp2 types.Hash
	err = client.Call(&resp2, "ledger_blockHash", blk)
	if err != nil {
		t.Fatal(err)
	}
	if blk.GetHash() != resp2 {
		t.Fatal()
	}
}

func TestRPC_WebSocket(t *testing.T) {
	teardownTestCase, rpc := setupTestCase(t)
	defer teardownTestCase(t)

	_, address, _ := scheme(rpc.config.RPC.WSEndpoint)
	client, err := jsonrpc.Dial(fmt.Sprintf("ws://%s", address))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	blk := new(types.StateBlock)
	blk.Token = common.ChainToken()
	rpc.ledger.AddStateBlock(blk)
	var resp2 types.Hash
	err = client.Call(&resp2, "ledger_blockHash", blk)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp2)
}

//
//func TestRPC_IPC(t *testing.T) {
//	teardownTestCase := setupTestCase(t)
//	defer teardownTestCase(t)
//
//	client, err := Dial(rpc.config.RPC.IPCEndpoint)
//	defer func() {
//		if client != nil {
//			client.Close()
//		}
//	}()
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	addr := mock.Address()
//	var resp bool
//	err = client.Call(&resp, "account_validate", addr)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if !resp {
//		t.Fatal()
//	}
//}
//
//func TestRPC_Attach(t *testing.T) {
//	teardownTestCase := setupTestCase(t)
//	defer teardownTestCase(t)
//
//	client, err := rpc.Attach()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	defer func() {
//		if client != nil {
//			client.Close()
//		}
//	}()
//
//	addr := mock.Address()
//	var resp bool
//	err = client.Call(&resp, "account_validate", addr)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if !resp {
//		t.Fatal()
//	}
//}
//
//func defaultIPCEndpoint(str string) string {
//	if runtime.GOOS == "windows" {
//		return `\\.\pipe\gqlc_test.ipc`
//	}
//	return str
//}
