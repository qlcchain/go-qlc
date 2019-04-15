package rpc

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/test/mock"
)

var (
	rpc   *RPC
	count int
	lock  = sync.RWMutex{}
	lock2 = sync.RWMutex{}
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Parallel()
	lock.Lock()
	defer lock.Unlock()
	count = count + 1
	rpcDir := filepath.Join(config.QlcTestDataDir(), "rpc")
	if rpc == nil {
		config := &config.Config{
			DataDir: rpcDir,
			RPC: &config.RPCConfigV2{
				Enable:       true,
				HTTPEndpoint: "tcp4://0.0.0.0:19735",
				WSEndpoint:   "tcp4://0.0.0.0:19736",
				IPCEndpoint:  defaultIPCEndpoint(filepath.Join(rpcDir, "qlc_test.ipc")),
				WSEnabled:    true,
				IPCEnabled:   true,
				HTTPEnabled:  true,
			},
		}
		eb := event.New()
		var err error
		rpc, err = NewRPC(config, eb)
		if err != nil {
			t.Fatal(err)
		}
		err = rpc.StartRPC()
		if err != nil {
			t.Fatal(err)
		}
		t.Log("rpc started")
	}
	return func(t *testing.T) {
		lock2.Lock()
		defer lock2.Unlock()
		count = count - 1
		if count == 0 {
			rpc.StopRPC()
			rpc.ledger.Close()
			rpc.wallet.Close()
			err := os.RemoveAll(rpcDir)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestRPC_HTTP(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	_, address, _ := scheme(rpc.config.RPC.HTTPEndpoint)
	client, err := Dial(fmt.Sprintf("http://%s", address))
	defer client.Close()

	if err != nil {
		t.Fatal(err)
	}

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
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	_, address, _ := scheme(rpc.config.RPC.WSEndpoint)
	client, err := Dial(fmt.Sprintf("ws://%s", address))
	defer client.Close()
	if err != nil {
		t.Fatal(err)
	}

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

func TestRPC_IPC(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	client, err := Dial(rpc.config.RPC.IPCEndpoint)
	defer client.Close()

	if err != nil {
		t.Fatal(err)
	}

	addr := mock.Address()
	var resp bool
	err = client.Call(&resp, "account_validate", addr)
	if err != nil {
		t.Fatal(err)
	}
	if !resp {
		t.Fatal()
	}
}

func TestRPC_Attach(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	client, err := rpc.Attach()
	defer client.Close()

	if err != nil {
		t.Fatal(err)
	}

	addr := mock.Address()
	var resp bool
	err = client.Call(&resp, "account_validate", addr)
	if err != nil {
		t.Fatal(err)
	}
	if !resp {
		t.Fatal()
	}
}

func defaultIPCEndpoint(str string) string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\gqlc_test.ipc`
	}
	return str
}
