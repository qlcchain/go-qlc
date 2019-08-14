// +build integrate

package test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/process"
)

func TestOpenFork(t *testing.T) {
	//node1
	dir1 := filepath.Join(config.QlcTestDataDir(), "transaction", uuid.New().String())
	cfgFile1, _ := config.DefaultConfig(dir1)
	cfgFile1.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
	cfgFile1.P2P.Discovery.DiscoveryInterval = 3
	cfgFile1.P2P.SyncInterval = 30
	cfgFile1.LogLevel = "error"
	cfgFile1.RPC.Enable = true
	b1 := "/ip4/0.0.0.0/tcp/19741/ipfs/" + cfgFile1.P2P.ID.PeerID

	// node1
	dir2 := filepath.Join(config.QlcTestDataDir(), "transaction", uuid.New().String())
	cfgFile2, _ := config.DefaultConfig(dir2)
	cfgFile2.P2P.Listen = "/ip4/127.0.0.1/tcp/19742"
	cfgFile2.P2P.SyncInterval = 30
	cfgFile2.LogLevel = "error"
	cfgFile2.RPC.Enable = true
	cfgFile2.RPC.HTTPEndpoint = "tcp4://0.0.0.0:29735"
	cfgFile2.RPC.WSEnabled = false
	cfgFile2.RPC.IPCEnabled = false
	b2 := "/ip4/0.0.0.0/tcp/19742/ipfs/" + cfgFile2.P2P.ID.PeerID

	cfgFile1.P2P.BootNodes = []string{b2}
	cfgFile2.P2P.BootNodes = []string{b1}

	cfgFile1Byte, _ := json.Marshal(cfgFile1)
	fmt.Println("node1 config \n", string(cfgFile1Byte))
	cfgFile2Byte, _ := json.Marshal(cfgFile2)
	fmt.Println("node2 config \n", string(cfgFile2Byte))

	defer func() {
		fmt.Println(" close servers")
		closeServer(node1, t)
		closeServer(node2, t)
	}()
	fmt.Println(" start node1....")
	node1.dir = dir1

	// start node1
	initNode(node1, cfgFile1, []*types.Account{testAccount}, t)

	// qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s
	tPrivateKey := "a2bc57c1d9dc433a411d6bfff1a24538a4fbd7026edb13b2909d9a60dff5f3c6d0503c72a9bd4df1b6cb3c6c4806a505acf0c69a1e2e790b6e61774da5c5653b"
	tBytes, _ := hex.DecodeString(tPrivateKey)
	tAccount := types.NewAccount(tBytes)
	client1, err := node1.rPCService.RPC().Attach()
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
	for !b {
		b1, err := node1.ledgerService.Ledger.HasStateBlockConfirmed(openblock1.GetHash())
		if err != nil {
			t.Fatal(err)
		}
		if b1 {
			b = true
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println(" start node2....")
	node2.dir = dir2

	// start node2
	initNode(node2, cfgFile2, nil, t)
	client2, err := node2.rPCService.RPC().Attach()
	if err != nil {
		t.Fatal(err)
	}
	verifier := process.NewLedgerVerifier(node2.ledgerService.Ledger)
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
	tm1, err := node1.ledgerService.Ledger.GetTokenMeta(tAccount.Address(), common.ChainToken())
	if err != nil {
		t.Fatal(err)
	}
	tm2, err := node2.ledgerService.Ledger.GetTokenMeta(tAccount.Address(), common.ChainToken())
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
