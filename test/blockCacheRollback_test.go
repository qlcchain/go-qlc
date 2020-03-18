// +build integrate

package test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/util"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
)

func TestBlockCacheRollback(t *testing.T) {
	//node1
	cachedir, _ := os.UserCacheDir()
	dir1 := filepath.Join(cachedir, uuid.New().String())
	defer func() {
		os.RemoveAll(dir1)
	}()
	ctx1, err := NewChain(dir1, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
		cfg.P2P.Discovery.DiscoveryInterval = 3
		cfg.P2P.SyncInterval = 30
		//cfg.LogLevel = "info"
		cfg.RPC.Enable = true
		cfg.RPC.HTTPEnabled = false
		cfg.RPC.WSEnabled = false
		cfg.RPC.IPCEnabled = true
		cfg.PoV.PovEnabled = false
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx1.SetAccounts([]*types.Account{testAccount})

	cfg1, err := ctx1.Config()
	if err != nil {
		t.Fatal(err)
	}
	b1 := "/ip4/0.0.0.0/tcp/19741/p2p/" + cfg1.P2P.ID.PeerID

	// node1
	dir2 := filepath.Join(cachedir, uuid.New().String())
	ctx2, err := NewChain(dir2, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19742"
		cfg.P2P.SyncInterval = 30
		//cfg.LogLevel = "info"
		cfg.RPC.Enable = true
		cfg.RPC.HTTPEnabled = false
		//cfg.RpcCall.HTTPEndpoint = "tcp4://0.0.0.0:29735"
		cfg.RPC.WSEnabled = false
		cfg.RPC.IPCEnabled = true
		cfg.PoV.PovEnabled = false
	})
	if err != nil {
		t.Fatal(err)
	}

	cfg2, err := ctx1.Config()
	if err != nil {
		t.Fatal(err)
	}
	b2 := "/ip4/0.0.0.0/tcp/19742/p2p/" + cfg2.P2P.ID.PeerID

	cfg1.P2P.BootNodes = []string{b2}
	cfg2.P2P.BootNodes = []string{b1}

	fmt.Println(" start node1....")
	fmt.Println(util.ToIndentString(cfg1))
	err = ctx1.Init(func() error {
		return chain.RegisterServices(ctx1)
	})
	if err != nil {
		t.Fatal(err)
	}
	err = ctx1.Start()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(" start node2....")
	fmt.Println(util.ToIndentString(cfg2))
	err = ctx2.Init(func() error {
		return chain.RegisterServices(ctx2)
	})
	if err != nil {
		t.Fatal(err)
	}
	err = ctx2.Start()
	if err != nil {
		t.Fatal(err)
	}
	ctx1.WaitForever()
	ctx2.WaitForever()

	defer func() {
		fmt.Println(" close servers")
		_ = ctx1.Stop()
		_ = ctx2.Stop()
		_ = os.RemoveAll(dir1)
		_ = os.RemoveAll(dir2)
	}()

	// qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s
	tPrivateKey := "a2bc57c1d9dc433a411d6bfff1a24538a4fbd7026edb13b2909d9a60dff5f3c6d0503c72a9bd4df1b6cb3c6c4806a505acf0c69a1e2e790b6e61774da5c5653b"
	tBytes, _ := hex.DecodeString(tPrivateKey)
	tAccount := types.NewAccount(tBytes)

	rpc1, err := RPC(ctx1)
	if err != nil {
		t.Fatal(err)
	}
	client1, err := rpc1.Attach()
	if err != nil {
		t.Fatal(err)
	}
	sAmount := types.Balance{Int: big.NewInt(1000)}
	s1 := sendTransaction(client1, *testAccount, *tAccount, sAmount, t)
	fmt.Println(s1.String())

	rpc2, err := RPC(ctx2)
	if err != nil {
		t.Fatal(err)
	}
	client2, err := rpc2.Attach()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	s2 := sendTransaction(client2, *testAccount, *tAccount, sAmount, t)
	fmt.Println(s2.String())
	l2, _ := Ledger(ctx2)
	count, err := l2.CountBlockCache()
	if count != 1 {
		t.Fatal("block cache add error")
	}
	time.Sleep(35 * time.Second)
	count, err = l2.CountBlockCache()
	if count != 0 {
		t.Fatal("block cache rollback error")
	}
}
