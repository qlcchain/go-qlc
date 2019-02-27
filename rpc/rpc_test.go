package rpc

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/rpc/api"
	"github.com/qlcchain/go-qlc/test/mock"
)

var (
	rs    *RPCService
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
	if rs == nil {
		//dir := filepath.Join(rpcDir, "config.json")
		cm := config.NewCfgManager(rpcDir)
		cfg, err := cm.Load()
		cfg.DataDir = rpcDir
		dpos := consensus.DposService{}
		rs = NewRPCService(cfg, &dpos)

		cfg.RPC = new(config.RPCConfig)
		cfg.RPC.HTTPEndpoint = "tcp4://0.0.0.0:19735"
		cfg.RPC.WSEndpoint = "tcp4://0.0.0.0:19736"
		cfg.RPC.IPCEndpoint = defaultIPCEndpoint(filepath.Join(cfg.DataDir, "qlc_test.ipc"))
		cfg.RPC.WSEnabled = true
		cfg.RPC.IPCEnabled = true
		cfg.RPC.HTTPEnabled = true
		rs.rpc.config = cfg

		err = rs.Start()
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
			rs.Stop()
			rs.rpc.ledger.Close()
			rs.rpc.dpos.Stop()
			rs.rpc.wallet.Close()
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

	_, address, _ := scheme(rs.rpc.config.RPC.HTTPEndpoint)
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
	blk.Token = mock.GetChainTokenType()
	rs.rpc.ledger.AddStateBlock(blk)
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

	_, address, _ := scheme(rs.rpc.config.RPC.WSEndpoint)
	client, err := Dial(fmt.Sprintf("ws://%s", address))
	defer client.Close()
	if err != nil {
		t.Fatal(err)
	}
	acc := mock.Account()
	blk := new(types.StateBlock)
	blk.Address = acc.Address()
	blk.Previous = types.ZeroHash
	blk.Token = mock.GetChainTokenType()
	rs.rpc.ledger.AddStateBlock(blk)

	var resp []api.APIBlock
	err = client.Call(&resp, "ledger_blocksInfo", []types.Hash{blk.GetHash()})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}

func TestRPC_IPC(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	client, err := Dial(rs.rpc.config.RPC.IPCEndpoint)
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

	client, err := rs.rpc.Attach()
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
