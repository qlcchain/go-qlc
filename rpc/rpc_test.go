package rpc

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/test/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *RPC) {
	t.Parallel()

	rpcDir := filepath.Join(config.DefaultDataDir(), "rpc")
	dir := filepath.Join(rpcDir, "config.json")
	cm := config.NewCfgManager(dir)
	cfg, err := cm.Load()
	cfg.DataDir = rpcDir
	dpos := consensus.DposService{}
	rs := NewRPCService(cfg, &dpos)

	cfg.RPC = new(config.RPCConfig)
	cfg.RPC.HTTPEndpoint = "0.0.0.0:29734"
	cfg.RPC.WSEndpoint = "0.0.0.0:29735"
	cfg.RPC.IPCEndpoint = "29736"
	cfg.RPC.WSEnabled = true
	cfg.RPC.IPCEnabled = true
	cfg.RPC.HTTPEnabled = true
	rs.rpc.config = cfg

	err = rs.Start()
	if err != nil {
		logger.Info(err)
	}
	logger.Info("rpc started")
	return func(t *testing.T) {
		rs.Stop()
		err = os.RemoveAll(rpcDir)
		if err != nil {
			t.Fatal(err)
		}
	}, rs.rpc
}

func TestRPC_Client(t *testing.T) {
	teardownTestCase, r := setupTestCase(t)
	defer teardownTestCase(t)

	//client, err := Dial(fmt.Sprintf("http://%s", r.config.RPC.HTTPEndpoint))
	client, err := Dial(fmt.Sprintf("ws://%s", r.config.RPC.WSEndpoint))
	//client, err := Dial(r.config.RPC.IPCEndpoint)
	if err != nil {
		logger.Info(err)
	}

	addr := mock.Address()
	ac := mock.AccountMeta(addr)
	r.ledger.AddAccountMeta(ac)
	addr2 := mock.Address()
	ac2 := mock.AccountMeta(addr2)
	r.ledger.AddAccountMeta(ac2)
	var resp map[types.Address]map[types.Hash]types.Balance
	err = client.Call(&resp, "qlcclassic_accountsBalances", []types.Address{addr, addr2})
	//err = client.Call(&resp, "qlcclassic_accountsBalances", []string{addr.String(), addr2.String()})
	if err != nil {
		t.Fatal(err)
	}
	logger.Info(resp)
}

func TestRPC_Client2(t *testing.T) {
	teardownTestCase, r := setupTestCase(t)
	defer teardownTestCase(t)

	client, err := Dial(fmt.Sprintf("http://%s", r.config.RPC.HTTPEndpoint))
	//client, err := Dial(fmt.Sprintf("ws://%s", r.config.RPC.WSEndpoint))
	//client, err := Dial(r.config.RPC.IPCEndpoint)
	if err != nil {
		logger.Info(err)
	}

	b1 := mock.StateBlock()
	r.ledger.AddBlock(b1)
	b2 := mock.StateBlock()
	r.ledger.AddBlock(b2)
	var resp []*types.StateBlock
	err = client.Call(&resp, "qlcclassic_blocksInfo", []types.Hash{b1.GetHash(), b2.GetHash()})
	if err != nil {
		t.Fatal(err)
	}
	logger.Info(resp)
	for _, b := range resp {
		fmt.Println(b.GetHash())
	}
}

func TestRPC_Client3(t *testing.T) {
	teardownTestCase, r := setupTestCase(t)
	defer teardownTestCase(t)

	//client, err := Dial(fmt.Sprintf("http://%s", r.config.RPC.HTTPEndpoint))
	//client, err := Dial(fmt.Sprintf("ws://%s", r.config.RPC.WSEndpoint))
	client, err := Dial(r.config.RPC.IPCEndpoint)
	if err != nil {
		logger.Info(err)
	}

	var resp ledger.ProcessResult
	b := mock.StateBlock()
	sb, _ := b.(*types.StateBlock)
	err = client.Call(&resp, "qlcclassic_process", sb)
	if err != nil {
		t.Fatal(err)
	}
	logger.Info(resp)
}
