// +build integrate

package test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"math/big"
	"path/filepath"
	"testing"
	"time"
)

func TestBlockCacheRollback(t *testing.T) {
	//node1
	dir1 := filepath.Join(config.QlcTestDataDir(), "cache", uuid.New().String())
	cfgFile1, _ := config.DefaultConfig(dir1)
	cfgFile1.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
	cfgFile1.P2P.Discovery.DiscoveryInterval = 3
	cfgFile1.P2P.SyncInterval = 30
	cfgFile1.LogLevel = "info"
	cfgFile1.RPC.Enable = true
	b1 := "/ip4/0.0.0.0/tcp/19741/ipfs/" + cfgFile1.P2P.ID.PeerID

	// node1
	dir2 := filepath.Join(config.QlcTestDataDir(), "cache", uuid.New().String())
	cfgFile2, _ := config.DefaultConfig(dir2)
	cfgFile2.P2P.Listen = "/ip4/127.0.0.1/tcp/19742"
	cfgFile2.P2P.SyncInterval = 30
	cfgFile2.LogLevel = "info"
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
	s1 := sendTransaction(client1, *testAccount, *tAccount, sAmount, t)
	fmt.Println(s1.String())
	fmt.Println(" start node2....")
	node2.dir = dir2

	// start node2
	initNode(node2, cfgFile2, nil, t)
	client2, err := node2.rPCService.RPC().Attach()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	s2 := sendTransaction(client2, *testAccount, *tAccount, sAmount, t)
	fmt.Println(s2.String())
	count, err := node2.ledgerService.Ledger.CountBlockCache()
	if count != 1 {
		t.Fatal("block cache add error")
	}
	time.Sleep(35 * time.Second)
	count, err = node2.ledgerService.Ledger.CountBlockCache()
	if count != 0 {
		t.Fatal("block cache rollback error")
	}
}
