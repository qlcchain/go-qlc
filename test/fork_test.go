// +build integrate

package test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var povSyncState common.SyncState

func onPovSyncState(state common.SyncState) {
	povSyncState = state
}

func envSetup(t *testing.T, ctx1 *context.ChainContext, ctx2 *context.ChainContext) {
	var sendBlock, recvBlock, changeBlock types.StateBlock

	_ = json.Unmarshal([]byte(jsonTestSend), &sendBlock)
	_ = json.Unmarshal([]byte(jsonTestReceive), &recvBlock)
	_ = json.Unmarshal([]byte(jsonTestChangeRepresentative), &changeBlock)

	l1 := ledger.NewLedger(ctx1.ConfigFile())
	lv1 := process.NewLedgerVerifier(l1)

	result, _ := lv1.Process(&sendBlock)
	if result != process.Progress {
		t.Fatal("process result", result)
	}

	result, _ = lv1.Process(&recvBlock)
	if result != process.Progress {
		t.Fatal("process result", result)
	}

	result, _ = lv1.Process(&changeBlock)
	if result != process.Progress {
		t.Fatal("process result", result)
	}

	l2 := ledger.NewLedger(ctx2.ConfigFile())
	lv2 := process.NewLedgerVerifier(l2)

	result, _ = lv2.Process(&sendBlock)
	if result != process.Progress {
		t.Fatal("process result", result)
	}

	result, _ = lv2.Process(&recvBlock)
	if result != process.Progress {
		t.Fatal("process result", result)
	}

	result, _ = lv2.Process(&changeBlock)
	if result != process.Progress {
		t.Fatal("process result", result)
	}

	eb := ctx1.EventBus()
	_, _ = eb.SubscribeSync(common.EventPovSyncState, onPovSyncState)
}

func TestFork(t *testing.T) {
	povSyncState = common.SyncNotStart
	testBytes, _ := hex.DecodeString(testPrivateKey)
	testAccount := types.NewAccount(testBytes)
	rootDir := filepath.Join(config.QlcTestDataDir(), "fork")

	//node1
	dir1 := filepath.Join(rootDir, uuid.New().String())
	ctx1, err := NewChain(dir1, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
		cfg.P2P.Discovery.DiscoveryInterval = 3
		cfg.P2P.SyncInterval = 3000
		cfg.LogLevel = "error"
		cfg.RPC.Enable = true
		cfg.RPC.HTTPEndpoint = "tcp4://0.0.0.0:29735"
		cfg.RPC.WSEnabled = false
		cfg.RPC.IPCEnabled = false
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
	dir2 := filepath.Join(rootDir, uuid.New().String())
	ctx2, err := NewChain(dir2, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19742"
		cfg.P2P.SyncInterval = 3000
		cfg.LogLevel = "error"
		cfg.RPC.Enable = true
		cfg.RPC.HTTPEndpoint = "tcp4://0.0.0.0:39735"
		cfg.RPC.WSEnabled = false
		cfg.RPC.IPCEnabled = false
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

	fmt.Println("start node1....")
	ctx1.SetAccounts([]*types.Account{testAccount})
	_ = ctx1.Init(func() error {
		return chain.RegisterServices(ctx1)
	})
	err = ctx1.Start()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("start node2....")
	_ = ctx2.Init(func() error {
		return chain.RegisterServices(ctx2)
	})
	err = ctx2.Start()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		fmt.Println("closing server1...")
		_ = ctx1.Stop()
		fmt.Println("closing server2...")
		_ = ctx2.Stop()
		fmt.Println("removing files...")
		os.RemoveAll(rootDir)
	}()

	//waiting all service start successful
	ctx1.WaitForever()
	ctx2.WaitForever()
	envSetup(t, ctx1, ctx2)

	timerCheckState := time.NewTicker(3 * time.Second)
	checkTimes := 0
	checkQuit := false

	for !checkQuit {
		select {
		case <-timerCheckState.C:
			if povSyncState == common.SyncDone {
				checkQuit = true
				break
			} else {
				checkTimes++
				if checkTimes >= 30 {
					t.Fatal("pov not ready")
				}
			}
		}
	}

	// qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s
	tPrivateKey := "a2bc57c1d9dc433a411d6bfff1a24538a4fbd7026edb13b2909d9a60dff5f3c6d0503c72a9bd4df1b6cb3c6c4806a505acf0c69a1e2e790b6e61774da5c5653b"
	tBytes, _ := hex.DecodeString(tPrivateKey)
	tAccount := types.NewAccount(tBytes)

	// generate sendBlock on node2, but dont publish it
	var n2Sendblock types.StateBlock
	n2Sendblock.Address = testAccount.Address()
	n2Sendblock.Link = types.Hash(tAccount.Address())
	n2Sendblock.Token = common.ChainToken()
	amount1 := types.Balance{Int: big.NewInt(1000)}

	l2 := ledger.NewLedger(ctx2.ConfigFile())
	lv2 := process.NewLedgerVerifier(l2)
	blk2, err := l2.GenerateSendBlock(&n2Sendblock, amount1, testAccount.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("blk2", blk2)

	r, _ := lv2.Process(blk2)
	if r != process.Progress {
		t.Fatal(r)
	}

	// generate sendBlock on node1 and publish it, will rollback node2's forked block, node1 is rep
	var n1Sendblock types.StateBlock
	n1Sendblock.Address = testAccount.Address()
	n1Sendblock.Link = types.Hash(tAccount.Address())
	n1Sendblock.Token = common.ChainToken()
	amount2 := types.Balance{Int: big.NewInt(2000)}

	l1 := ledger.NewLedger(ctx1.ConfigFile())
	blk1, err := l1.GenerateSendBlock(&n1Sendblock, amount2, testAccount.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("blk1", blk1)

	rpc1, _ := RPC(ctx1)
	client1, err := rpc1.Attach()
	if err != nil {
		t.Fatal(err)
	}

	var h types.Hash
	err = client1.Call(&h, "ledger_process", &blk1)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3 * time.Second)

	if has, _ := l1.HasStateBlockConfirmed(blk1.GetHash()); !has {
		t.Fatal("node1 block not confirmed")
	}

	if has, _ := l2.HasStateBlockConfirmed(blk1.GetHash()); !has {
		t.Fatal("node2 block not confirmed")
	}
}

func TestPovFork(t *testing.T) {
	povSyncState = common.SyncNotStart
	testBytes, _ := hex.DecodeString(testPrivateKey)
	testAccount := types.NewAccount(testBytes)
	rootDir := filepath.Join(config.QlcTestDataDir(), "fork")

	//node1
	dir1 := filepath.Join(rootDir, uuid.New().String())
	ctx1, err := NewChain(dir1, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
		cfg.P2P.Discovery.DiscoveryInterval = 3
		cfg.P2P.SyncInterval = 3000
		cfg.LogLevel = "error"
		cfg.RPC.Enable = true
		cfg.RPC.HTTPEndpoint = "tcp4://0.0.0.0:29735"
		cfg.RPC.WSEnabled = false
		cfg.RPC.IPCEnabled = false
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
	dir2 := filepath.Join(rootDir, uuid.New().String())
	ctx2, err := NewChain(dir2, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19742"
		cfg.P2P.SyncInterval = 3000
		cfg.LogLevel = "error"
		cfg.RPC.Enable = true
		cfg.RPC.HTTPEndpoint = "tcp4://0.0.0.0:39735"
		cfg.RPC.WSEnabled = false
		cfg.RPC.IPCEnabled = false
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

	fmt.Println("start node1....")
	ctx1.SetAccounts([]*types.Account{testAccount})
	_ = ctx1.Init(func() error {
		return chain.RegisterServices(ctx1)
	})
	err = ctx1.Start()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("start node2....")
	_ = ctx2.Init(func() error {
		return chain.RegisterServices(ctx2)
	})
	err = ctx2.Start()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		fmt.Println("closing server1...")
		_ = ctx1.Stop()
		fmt.Println("closing server2...")
		_ = ctx2.Stop()
		fmt.Println("removing files...")
		os.RemoveAll(rootDir)
	}()

	//waiting all service start successful
	ctx1.WaitForever()
	ctx2.WaitForever()
	envSetup(t, ctx1, ctx2)

	timerCheckState := time.NewTicker(3 * time.Second)
	checkTimes := 0
	checkQuit := false

	for !checkQuit {
		select {
		case <-timerCheckState.C:
			if povSyncState == common.SyncDone {
				checkQuit = true
				break
			} else {
				checkTimes++
				if checkTimes >= 30 {
					t.Fatal("pov not ready")
				}
			}
		}
	}

	// qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s
	tPrivateKey := "a2bc57c1d9dc433a411d6bfff1a24538a4fbd7026edb13b2909d9a60dff5f3c6d0503c72a9bd4df1b6cb3c6c4806a505acf0c69a1e2e790b6e61774da5c5653b"
	tBytes, _ := hex.DecodeString(tPrivateKey)
	tAccount := types.NewAccount(tBytes)

	// generate sendBlock on node2, but dont publish it
	var n2Sendblock types.StateBlock
	n2Sendblock.Address = testAccount.Address()
	n2Sendblock.Link = types.Hash(tAccount.Address())
	n2Sendblock.Token = common.ChainToken()
	amount1 := types.Balance{Int: big.NewInt(1000)}

	l2 := ledger.NewLedger(ctx2.ConfigFile())
	lv2 := process.NewLedgerVerifier(l2)
	blk2, err := l2.GenerateSendBlock(&n2Sendblock, amount1, testAccount.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("blk2", blk2)

	r, _ := lv2.Process(blk2)
	if r != process.Progress {
		t.Fatal(r)
	}

	povBlock, td := mock.GeneratePovBlock(nil, 0)
	l2.AddPovBlock(povBlock, td)
	l2.AddPovBestHash(povBlock.GetHeight(), povBlock.GetHash())
	l2.SetPovLatestHeight(povBlock.GetHeight())

	txlookup := &types.PovTxLookup{
		BlockHash:   povBlock.GetHash(),
		BlockHeight: povBlock.GetHeight(),
		TxIndex:     1,
	}
	err = l2.AddPovTxLookup(blk2.GetHash(), txlookup)
	if err != nil {
		t.Fatal(err)
	}

	// generate sendBlock on node1 and publish it, will rollback node2's forked block, node1 is rep
	var n1Sendblock types.StateBlock
	n1Sendblock.Address = testAccount.Address()
	n1Sendblock.Link = types.Hash(tAccount.Address())
	n1Sendblock.Token = common.ChainToken()
	amount2 := types.Balance{Int: big.NewInt(2000)}

	l1 := ledger.NewLedger(ctx1.ConfigFile())
	blk1, err := l1.GenerateSendBlock(&n1Sendblock, amount2, testAccount.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("blk1", blk1)

	rpc1, _ := RPC(ctx1)
	client1, err := rpc1.Attach()
	if err != nil {
		t.Fatal(err)
	}

	var h types.Hash
	err = client1.Call(&h, "ledger_process", &blk1)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3 * time.Second)

	if has, _ := l1.HasStateBlockConfirmed(blk1.GetHash()); !has {
		t.Fatal("node1 block not confirmed")
	}

	if has, _ := l2.HasStateBlockConfirmed(blk1.GetHash()); has {
		t.Fatal("node2 block rollback")
	}
}

func TestCacheFork(t *testing.T) {
	povSyncState = common.SyncNotStart
	testBytes, _ := hex.DecodeString(testPrivateKey)
	testAccount := types.NewAccount(testBytes)
	rootDir := filepath.Join(config.QlcTestDataDir(), "fork")

	//node1
	dir1 := filepath.Join(rootDir, uuid.New().String())
	ctx1, err := NewChain(dir1, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
		cfg.P2P.Discovery.DiscoveryInterval = 3
		cfg.P2P.SyncInterval = 3000
		cfg.LogLevel = "error"
		cfg.RPC.Enable = true
		cfg.RPC.HTTPEndpoint = "tcp4://0.0.0.0:29735"
		cfg.RPC.WSEnabled = false
		cfg.RPC.IPCEnabled = false
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
	dir2 := filepath.Join(rootDir, uuid.New().String())
	ctx2, err := NewChain(dir2, func(cfg *config.Config) {
		cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19742"
		cfg.P2P.SyncInterval = 3000
		cfg.LogLevel = "error"
		cfg.RPC.Enable = true
		cfg.RPC.HTTPEndpoint = "tcp4://0.0.0.0:39735"
		cfg.RPC.WSEnabled = false
		cfg.RPC.IPCEnabled = false
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

	fmt.Println("start node1....")
	ctx1.SetAccounts([]*types.Account{testAccount})
	_ = ctx1.Init(func() error {
		return chain.RegisterServices(ctx1)
	})
	err = ctx1.Start()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("start node2....")
	_ = ctx2.Init(func() error {
		return chain.RegisterServices(ctx2)
	})
	err = ctx2.Start()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		fmt.Println("closing server1...")
		_ = ctx1.Stop()
		fmt.Println("closing server2...")
		_ = ctx2.Stop()
		fmt.Println("removing files...")
		os.RemoveAll(rootDir)
	}()

	//waiting all service start successful
	ctx1.WaitForever()
	ctx2.WaitForever()
	envSetup(t, ctx1, ctx2)

	timerCheckState := time.NewTicker(3 * time.Second)
	checkTimes := 0
	checkQuit := false

	for !checkQuit {
		select {
		case <-timerCheckState.C:
			if povSyncState == common.SyncDone {
				checkQuit = true
				break
			} else {
				checkTimes++
				if checkTimes >= 30 {
					t.Fatal("pov not ready")
				}
			}
		}
	}

	// qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s
	tPrivateKey := "a2bc57c1d9dc433a411d6bfff1a24538a4fbd7026edb13b2909d9a60dff5f3c6d0503c72a9bd4df1b6cb3c6c4806a505acf0c69a1e2e790b6e61774da5c5653b"
	tBytes, _ := hex.DecodeString(tPrivateKey)
	tAccount := types.NewAccount(tBytes)

	// generate sendBlock on node2, but dont publish it
	var n1Sendblock1 types.StateBlock
	n1Sendblock1.Address = testAccount.Address()
	n1Sendblock1.Link = types.Hash(tAccount.Address())
	n1Sendblock1.Token = common.ChainToken()
	amount1 := types.Balance{Int: big.NewInt(1000)}

	l1 := ledger.NewLedger(ctx1.ConfigFile())
	lv1 := process.NewLedgerVerifier(l1)
	blk1, err := l1.GenerateSendBlock(&n1Sendblock1, amount1, testAccount.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("blk1", blk1)

	// generate sendBlock on node1 and publish it, will rollback node2's forked block, node1 is rep
	var n1Sendblock2 types.StateBlock
	n1Sendblock2.Address = testAccount.Address()
	n1Sendblock2.Link = types.Hash(tAccount.Address())
	n1Sendblock2.Token = common.ChainToken()
	amount2 := types.Balance{Int: big.NewInt(2000)}

	blk2, err := l1.GenerateSendBlock(&n1Sendblock2, amount2, testAccount.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("blk2", blk2)

	r, _ := lv1.Process(blk1)
	if r != process.Progress {
		t.Fatal(r)
	}

	l1.EB.Publish(common.EventGenerateBlock, process.Progress, blk2)

	time.Sleep(3 * time.Second)

	if has, _ := l1.HasStateBlockConfirmed(blk1.GetHash()); !has {
		t.Fatal("node1 block rollback")
	}
}