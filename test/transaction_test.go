// +build  integrate

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/ledger"

	"github.com/google/uuid"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/rpc/api"
)

var (
	testBytes, _ = hex.DecodeString(testPrivateKey)
	testAccount  = types.NewAccount(testBytes)
	//testAddress = testAccount.Address()
)

func TestTransaction(t *testing.T) {
	fmt.Println("transaction start ")
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

	t.Log(" start node1....")
	ctx1.SetAccounts([]*types.Account{testAccount})
	_ = ctx1.Init(func() error {
		return chain.RegisterServices(ctx1)
	})
	err = ctx1.Start()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(" start node2....")
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
	// qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s
	tPrivateKey := "a2bc57c1d9dc433a411d6bfff1a24538a4fbd7026edb13b2909d9a60dff5f3c6d0503c72a9bd4df1b6cb3c6c4806a505acf0c69a1e2e790b6e61774da5c5653b"
	tBytes, _ := hex.DecodeString(tPrivateKey)
	tAccount := types.NewAccount(tBytes)

	rpc, _ := RPC(ctx1)
	client1, err := rpc.Attach()
	if err != nil {
		t.Fatal(err)
	}

	l1, _ := Ledger(ctx1)
	l2, _ := Ledger(ctx2)

	// prepare two account to send transaction
	sAmount := types.Balance{Int: big.NewInt(10000000000000)}
	sendBlock := sendTransaction(client1, *testAccount, *tAccount, sAmount, t)

	b := false
	start := time.Now()
	for !b {
		if checkConfirmed(l1, l2, sendBlock.GetHash(), t) {
			b = true
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("confirmed use ", time.Now().Sub(start))

	tAccountOpenBlock := receiveTransaction(client1, *tAccount, sendBlock, t)
	b = false
	for !b {
		if checkConfirmed(l1, l2, tAccountOpenBlock.GetHash(), t) {
			b = true
		}
		time.Sleep(1 * time.Second)
	}

	// transaction
	// qlc_3hpt4k5hst4i1gdsn5o366owyndxdcoq3wtnrbsm8gw5edb4gatqjzbmwsc9
	cPrivateKey := "6ff74e6b363c87bfef33288ede0a1984126a4d771483d66ad9be24ff1ce6cd8bbeda1486fce85003979a0ea1212bcf517d5aab70f354c273333b8362d2272357"
	cBytes, _ := hex.DecodeString(cPrivateKey)
	cAccount := types.NewAccount(cBytes)

	m1 := 5000
	m2 := 5000
	amount1 := 10
	amount2 := 20
	var headerBlock1 types.StateBlock
	wg1 := sync.WaitGroup{}
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		for i := 0; i < m1; i++ {
			headerBlock1 = sendTransaction(client1, *testAccount, *cAccount, types.Balance{Int: big.NewInt(int64(amount1))}, t)
		}
	}()

	var headerBlock2 types.StateBlock
	wg2 := sync.WaitGroup{}
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		for j := 0; j < m2; j++ {
			headerBlock2 = sendTransaction(client1, *tAccount, *cAccount, types.Balance{Int: big.NewInt(int64(amount2))}, t)
		}
	}()
	wg2.Wait()
	wg1.Wait()

	fmt.Println("transaction finish ")
	b = false
	for !b {
		b1 := checkConfirmed(l1, l2, headerBlock1.GetHash(), t)
		b2 := checkConfirmed(l1, l2, headerBlock2.GetHash(), t)
		if b1 && b2 {
			b = true
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("consensus finish ")
	// check result
	fmt.Println("check node1")
	checkBlock(l1, 9+m1+m2, t)
	checkAccount(l1, testAccount.Address(), testReceiveBlock.Balance.Sub(sAmount).Sub(types.Balance{Int: big.NewInt(int64(m1 * amount1))}),
		m1+3, headerBlock1.GetHash(), testReceiveBlock.GetHash(), t)
	checkAccount(l1, tAccount.Address(), sAmount.Sub(types.Balance{Int: big.NewInt(int64(m2 * amount2))}),
		m2+1, headerBlock2.GetHash(), tAccountOpenBlock.GetHash(), t)
	checkRepresentation(l1, testAccount.Address(), testReceiveBlock.Balance.Sub(types.Balance{Int: big.NewInt(int64(m1*amount1 + m2*amount2))}), t)

	// check node2
	fmt.Println("check node2")
	checkBlock(l2, 9+m1+m2, t)
	checkAccount(l2, testAccount.Address(), testReceiveBlock.Balance.Sub(sAmount).Sub(types.Balance{Int: big.NewInt(int64(m1 * amount1))}),
		m1+3, headerBlock1.GetHash(), testReceiveBlock.GetHash(), t)
	checkAccount(l2, tAccount.Address(), sAmount.Sub(types.Balance{Int: big.NewInt(int64(m2 * amount2))}),
		m2+1, headerBlock2.GetHash(), tAccountOpenBlock.GetHash(), t)
	checkRepresentation(l2, testAccount.Address(), testReceiveBlock.Balance.Sub(types.Balance{Int: big.NewInt(int64(m1*amount1 + m2*amount2))}), t)

	fmt.Println("transaction successfully ")
}

func checkConfirmed(l1, l2 *ledger.Ledger, hash types.Hash, t *testing.T) bool {
	b1, err := l1.HasStateBlockConfirmed(hash)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := l2.HasStateBlockConfirmed(hash)
	if err != nil {
		t.Fatal(err)
	}

	if b1 && b2 {
		return true
	} else {
		return false
	}
}

func checkBlock(l *ledger.Ledger, blockCount int, t *testing.T) {
	bc, err := l.CountStateBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if bc != uint64(blockCount) {
		t.Fatal("block count error", bc, blockCount)
	}
}

func checkAccount(l *ledger.Ledger, address types.Address, amount types.Balance, blockCount int, header types.Hash, open types.Hash, t *testing.T) {
	tm, err := l.GetTokenMeta(address, config.ChainToken())
	if err != nil {
		t.Fatal(err)
	}
	if !tm.Balance.Equal(amount) {
		t.Fatal("balance error", address.String(), tm.Balance.String(), amount.String())
	}
	if tm.Header != header {
		t.Fatal("header block error", address.String(), tm.Header, header)
	}
	if tm.BlockCount != int64(blockCount) {
		t.Fatal("block count error,", address.String(), tm.BlockCount, blockCount)
	}

}

func checkRepresentation(l *ledger.Ledger, address types.Address, amount types.Balance, t *testing.T) {
	r, err := l.GetRepresentation(address)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Total.Equal(amount) {
		t.Fatal("representation amount error", address.String(), r.Total, amount)
	}
}

func sendTransaction(client *rpc.Client, from, to types.Account, amount types.Balance, t *testing.T) types.StateBlock {
	var sendBlock types.StateBlock
	para := api.APISendBlockPara{
		From:      from.Address(),
		To:        to.Address(),
		Amount:    amount,
		TokenName: "QLC",
	}
	if err := client.Call(&sendBlock, "ledger_generateSendBlock", &para, hex.EncodeToString(from.PrivateKey())); err != nil {
		t.Fatal(err)
	}
	var h types.Hash
	if err := client.Call(&h, "ledger_process", &sendBlock); err != nil {
		t.Fatal(err)
	}
	return sendBlock
}

func receiveTransaction(client *rpc.Client, acc types.Account, sendBlock types.StateBlock, t *testing.T) types.StateBlock {
	var receiverBlock types.StateBlock
	if err := client.Call(&receiverBlock, "ledger_generateReceiveBlock", &sendBlock, hex.EncodeToString(acc.PrivateKey())); err != nil {
		t.Fatal(err)
	}
	var h types.Hash
	if err := client.Call(&h, "ledger_process", &receiverBlock); err != nil {
		t.Fatal(err)
	}
	return receiverBlock
}
