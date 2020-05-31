// +build integrate

package test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/process"
)

func TestPending(t *testing.T) {
	testBytes, _ := hex.DecodeString(testPrivateKey)
	testAccount := types.NewAccount(testBytes)

	//node1
	dir1 := filepath.Join(config.QlcTestDataDir(), "transaction", uuid.New().String())
	ctx, err := NewChain(dir1, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
		cfg.P2P.Discovery.DiscoveryInterval = 3
		cfg.P2P.SyncInterval = 30
		cfg.LogLevel = "info"
		cfg.RPC.Enable = true
		cfg.RPC.HTTPEndpoint = "tcp4://0.0.0.0:19735"
		cfg.P2P.BootNodes = []string{}
		cfg.PoV.PovEnabled = false
	})

	if err != nil {
		t.Fatal(err)
	}
	_ = ctx.Init(func() error {
		return chain.RegisterServices(ctx)
	})
	err = ctx.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		fmt.Println(" close servers")
		_ = ctx.Stop()
	}()

	ctx.WaitForever()

	rpc, err := RPC(ctx)
	if err != nil {
		t.Fatal(err)
	}
	client, err := rpc.Attach()
	if err != nil {
		t.Fatal(err)
	}

	// qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s
	tPrivateKey := "a2bc57c1d9dc433a411d6bfff1a24538a4fbd7026edb13b2909d9a60dff5f3c6d0503c72a9bd4df1b6cb3c6c4806a505acf0c69a1e2e790b6e61774da5c5653b"
	tBytes, _ := hex.DecodeString(tPrivateKey)
	tAccount := types.NewAccount(tBytes)
	tAddress := tAccount.Address()

	sAmount := types.Balance{Int: big.NewInt(10000000000000)}
	sb := types.StateBlock{
		Address: testAccount.Address(),
		Token:   config.ChainToken(),
		Link:    tAddress.ToHash(),
	}

	l, err := Ledger(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sendBlock, err := l.GenerateSendBlock(&sb, sAmount, testAccount.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("sendBlock ", sendBlock.GetHash())
	lv := process.NewLedgerVerifier(l)
	if r, err := lv.BlockCheck(sendBlock); r != process.Progress || err != nil {
		t.Fatal(err)
	}
	if err := lv.BlockProcess(sendBlock); err != nil {
		t.Fatal(err)
	}

	// -------- check receive block repeatedly ------------

	// generate ReceiveBlock
	var receiverBlock types.StateBlock
	if err := client.Call(&receiverBlock, "ledger_generateReceiveBlock", &sendBlock, hex.EncodeToString(tAccount.PrivateKey())); err != nil {
		t.Fatal(err)
	}
	var h types.Hash
	if err := client.Call(&h, "ledger_process", &receiverBlock); err != nil {
		t.Fatal(err)
	}
	fmt.Println(receiverBlock)

	// generate ReceiveBlock repeatedly
	var receiverBlock1 types.StateBlock
	if err := client.Call(&receiverBlock1, "ledger_generateReceiveBlock", &sendBlock, hex.EncodeToString(tAccount.PrivateKey())); err != nil {
		t.Fatal(err)
	}
	var h2 types.Hash
	if err := client.Call(&h2, "ledger_process", &receiverBlock1); err != nil {
		t.Log(err)
	} else {
		t.Fatal("repeatedly check error")
	}

	// -------- check receive block cache rollback ------------

	var receiverBlock2 types.StateBlock
	if err := client.Call(&receiverBlock2, "ledger_generateReceiveBlock", &sendBlock, hex.EncodeToString(tAccount.PrivateKey())); err != nil {
		t.Fatal(err)
	}

	if err := l.AddBlockCache(&receiverBlock2); err != nil {
		t.Fatal(err)
	}

	count1 := 0
	err = l.GetPendingsByAddress(receiverBlock2.Address, func(key *types.PendingKey, value *types.PendingInfo) error {
		count1++
		return nil
	})
	if err != nil || count1 != 0 {
		t.Fatal("error pending")
	}
	count2 := 0
	err = l.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		count2++
		return nil
	})
	if err != nil || count2 != 0 {
		t.Fatal("error pending")
	}

	// wait roll back
	time.Sleep(70 * time.Second)

	count3 := 0
	err = l.GetPendingsByAddress(receiverBlock2.Address, func(key *types.PendingKey, value *types.PendingInfo) error {
		count3++
		return nil
	})
	if err != nil || count3 != 1 {
		t.Fatal("error pending")
	}

	count4 := 0
	err = l.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		count4++
		return nil
	})
	if err != nil || count4 != 1 {
		t.Fatal("error pending")
	}
	t.Log("done")
}
