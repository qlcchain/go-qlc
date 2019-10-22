// +build integrate

package test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/chain"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/process"
)

func TestOpenFork(t *testing.T) {
	//node1
	dir1 := filepath.Join(config.QlcTestDataDir(), "transaction", uuid.New().String())
	ctx1, err := NewChain(dir1, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
		cfg.P2P.Discovery.DiscoveryInterval = 3
		cfg.P2P.SyncInterval = 30
		cfg.LogLevel = "error"
		cfg.RPC.Enable = true
		cfg.PoV.PovEnabled = false
	})
	if err != nil {
		t.Fatal(err)
	}

	cfg1, err := ctx1.Config()
	if err != nil {
		t.Fatal(err)
	}
	b1 := "/ip4/0.0.0.0/tcp/19741/ipfs/" + cfg1.P2P.ID.PeerID

	// node2
	dir2 := filepath.Join(config.QlcTestDataDir(), "transaction", uuid.New().String())
	ctx2, err := NewChain(dir2, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19742"
		cfg.P2P.SyncInterval = 30
		cfg.LogLevel = "error"
		cfg.RPC.Enable = true
		cfg.RPC.HTTPEndpoint = "tcp4://0.0.0.0:29735"
		cfg.RPC.WSEnabled = false
		cfg.RPC.IPCEnabled = false
		cfg.PoV.PovEnabled = false
	})
	if err != nil {
		t.Fatal(err)
	}

	cfg2, err := ctx2.Config()
	if err != nil {
		t.Fatal(err)
	}

	b2 := "/ip4/0.0.0.0/tcp/19742/ipfs/" + cfg2.P2P.ID.PeerID

	cfg1.P2P.BootNodes = []string{b2}
	cfg2.P2P.BootNodes = []string{b1}

	fmt.Println(" start node1....")
	ctx1.SetAccounts([]*types.Account{testAccount})
	_ = ctx1.Init(func() error {
		return chain.RegisterServices(ctx1)
	})
	err = ctx1.Start()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(" start node2....")
	_ = ctx2.Init(func() error {
		return chain.RegisterServices(ctx2)
	})
	err = ctx2.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		fmt.Println(" close servers")
		_ = ctx1.Stop()
		_ = ctx2.Stop()
	}()

	ctx1.WaitForever()
	ctx2.WaitForever()

	//waiting all service start successful

	// qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s
	tPrivateKey := "a2bc57c1d9dc433a411d6bfff1a24538a4fbd7026edb13b2909d9a60dff5f3c6d0503c72a9bd4df1b6cb3c6c4806a505acf0c69a1e2e790b6e61774da5c5653b"
	tBytes, _ := hex.DecodeString(tPrivateKey)
	tAccount := types.NewAccount(tBytes)
	rpc1, _ := RPC(ctx1)
	client1, err := rpc1.Attach()
	if err != nil {
		t.Fatal(err)
	}
	sAmount := types.Balance{Int: big.NewInt(1000)}
	sendBlock1 := sendTransaction(client1, *testAccount, *tAccount, sAmount, t)
	sendBlock2 := sendTransaction(client1, *testAccount, *tAccount, sAmount, t)
	fmt.Println(sendBlock2)

	var openblock1 types.StateBlock
	if err := client1.Call(&openblock1, "ledger_generateReceiveBlock", &sendBlock1, hex.EncodeToString(tAccount.PrivateKey())); err != nil {
		t.Fatal(err)
	}
	fmt.Println("openblock1 ", openblock1.String())
	var h1 types.Hash
	if err := client1.Call(&h1, "ledger_process", &openblock1); err != nil {
		t.Fatal(err)
	}
	b := false

	l1, _ := Ledger(ctx1)
	for !b {
		b1, err := l1.HasStateBlockConfirmed(openblock1.GetHash())
		if err != nil {
			t.Fatal(err)
		}
		if b1 {
			b = true
		}
		time.Sleep(1 * time.Second)
	}

	rpc2, _ := RPC(ctx2)
	client2, err := rpc2.Attach()
	if err != nil {
		t.Fatal(err)
	}
	l2, _ := Ledger(ctx2)
	verifier := process.NewLedgerVerifier(l2)
	if err := verifier.BlockProcess(&sendBlock1); err != nil {
		t.Fatal(err)
	}
	if err := verifier.BlockProcess(&sendBlock2); err != nil {
		t.Fatal(err)
	}
	var openblock2 types.StateBlock
	if err := client2.Call(&openblock2, "ledger_generateReceiveBlock", &sendBlock2, hex.EncodeToString(tAccount.PrivateKey())); err != nil {
		t.Fatal(err)
	}
	var h2 types.Hash
	if err := client2.Call(&h2, "ledger_process", &openblock2); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Second)
	tm1, err := l1.GetTokenMeta(tAccount.Address(), common.ChainToken())
	if err != nil {
		t.Fatal(err)
	}
	tm2, err := l2.GetTokenMeta(tAccount.Address(), common.ChainToken())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(tm1)
	fmt.Println(tm2)
	if tm1.Header != tm2.Header || tm1.OpenBlock != tm2.OpenBlock || tm1.BlockCount != tm2.BlockCount ||
		!tm1.Balance.Equal(tm2.Balance) {
		t.Fatal("account error")
	}
}
