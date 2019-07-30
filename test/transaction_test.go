// +build  testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	ss "github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"
)

var (
	node1        = new(Service)
	node2        = new(Service)
	testBytes, _ = hex.DecodeString(testPrivateKey)
	testAccount  = types.NewAccount(testBytes)
	//testAddress = testAccount.Address()
)

type Service struct {
	ledgerService    *ss.LedgerService
	netService       *ss.P2PService
	consensusService *ss.ConsensusService
	rPCService       *ss.RPCService
	dir              string
}

func TestTransaction(t *testing.T) {
	fmt.Println("transaction start ")
	nodeConfig(t)
	defer func() {
		fmt.Println(" close servers")
		closeServer(node1, t)
		closeServer(node2, t)
	}()

	// qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s
	tPrivateKey := "a2bc57c1d9dc433a411d6bfff1a24538a4fbd7026edb13b2909d9a60dff5f3c6d0503c72a9bd4df1b6cb3c6c4806a505acf0c69a1e2e790b6e61774da5c5653b"
	tBytes, _ := hex.DecodeString(tPrivateKey)
	tAccount := types.NewAccount(tBytes)

	client1, err := node1.rPCService.RPC().Attach()
	if err != nil {
		t.Fatal(err)
	}
	// prepare two account to send transaction
	sAmount := types.Balance{Int: big.NewInt(10000000000000)}
	sendBlock := sendTransaction(client1, *testAccount, *tAccount, sAmount, t)

	b := false
	start := time.Now()
	for !b {
		if checkConfirmed(sendBlock.GetHash(), t) {
			b = true
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("confirmed use ", time.Now().Sub(start))

	tAccountOpenBlock := receiveTransaction(client1, *tAccount, sendBlock, t)
	b = false
	for !b {
		if checkConfirmed(tAccountOpenBlock.GetHash(), t) {
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
	go func() {
		for i := 0; i < m1; i++ {
			wg1.Add(1)
			headerBlock1 = sendTransaction(client1, *testAccount, *cAccount, types.Balance{Int: big.NewInt(int64(amount1))}, t)
			wg1.Done()
		}
	}()

	var headerBlock2 types.StateBlock
	wg2 := sync.WaitGroup{}
	go func() {
		for j := 0; j < m2; j++ {
			wg2.Add(1)
			headerBlock2 = sendTransaction(client1, *tAccount, *cAccount, types.Balance{Int: big.NewInt(int64(amount2))}, t)
			wg2.Done()
		}
	}()
	wg2.Wait()
	wg1.Wait()

	fmt.Println("transaction finish ")
	b = false
	for !b {
		b1 := checkConfirmed(headerBlock1.GetHash(), t)
		b2 := checkConfirmed(headerBlock2.GetHash(), t)
		if b1 && b2 {
			b = true
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("consensus finish ")
	// check result
	fmt.Println("check node1")
	checkBlock(node1, 9+m1+m2, t)
	checkAccount(node1, testAccount.Address(), testReceiveBlock.Balance.Sub(sAmount).Sub(types.Balance{Int: big.NewInt(int64(m1 * amount1))}),
		m1+3, headerBlock1.GetHash(), testReceiveBlock.GetHash(), t)
	checkAccount(node1, tAccount.Address(), sAmount.Sub(types.Balance{Int: big.NewInt(int64(m2 * amount2))}),
		m2+1, headerBlock2.GetHash(), tAccountOpenBlock.GetHash(), t)
	checkRepresentation(node1, testAccount.Address(), testReceiveBlock.Balance.Sub(types.Balance{Int: big.NewInt(int64(m1*amount1 + m2*amount2))}), t)

	// check node2
	fmt.Println("check node2")
	checkBlock(node2, 9+m1+m2, t)
	checkAccount(node2, testAccount.Address(), testReceiveBlock.Balance.Sub(sAmount).Sub(types.Balance{Int: big.NewInt(int64(m1 * amount1))}),
		m1+3, headerBlock1.GetHash(), testReceiveBlock.GetHash(), t)
	checkAccount(node2, tAccount.Address(), sAmount.Sub(types.Balance{Int: big.NewInt(int64(m2 * amount2))}),
		m2+1, headerBlock2.GetHash(), tAccountOpenBlock.GetHash(), t)
	checkRepresentation(node2, testAccount.Address(), testReceiveBlock.Balance.Sub(types.Balance{Int: big.NewInt(int64(m1*amount1 + m2*amount2))}), t)

	fmt.Println("transaction successfully ")
}

func checkConfirmed(hash types.Hash, t *testing.T) bool {
	b1, err := node1.ledgerService.Ledger.HasStateBlockConfirmed(hash)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := node2.ledgerService.Ledger.HasStateBlockConfirmed(hash)
	if err != nil {
		t.Fatal(err)
	}

	if b1 && b2 {
		return true
	} else {
		return false
	}
}

func checkBlock(service *Service, blockCount int, t *testing.T) {
	bc, err := service.ledgerService.Ledger.CountStateBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if bc != uint64(blockCount) {
		t.Fatal("block count error", bc, blockCount)
	}
}

func checkAccount(service *Service, address types.Address, amount types.Balance, blockCount int, header types.Hash, open types.Hash, t *testing.T) {
	tm, err := service.ledgerService.Ledger.GetTokenMeta(address, common.ChainToken())
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

func checkRepresentation(service *Service, address types.Address, amount types.Balance, t *testing.T) {
	r, err := service.ledgerService.Ledger.GetRepresentation(address)
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

func nodeConfig(t *testing.T) {
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

	fmt.Println(" start node1....")
	node1.dir = dir1
	initNode(node1, cfgFile1, []*types.Account{testAccount}, t)

	fmt.Println(" start node2....")
	node2.dir = dir2
	initNode(node2, cfgFile2, nil, t)
	time.Sleep(10 * time.Second)

	//err := node1.ledgerService.Ledger.GetStateBlocks(func(block *types.StateBlock) error {
	//	fmt.Println(block)
	//	return nil
	//})
	//if err != nil {
	//	t.Fatal(err)
	//}
}

func initNode(service *Service, cfg *config.Config, accounts []*types.Account, t *testing.T) {
	logService := log.NewLogService(cfg)
	_ = logService.Init()
	var err error
	service.ledgerService = ss.NewLedgerService(cfg)
	service.consensusService = ss.NewConsensusService(cfg, accounts)
	service.rPCService, err = ss.NewRPCService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	service.netService, err = ss.NewP2PService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ledgerService.Init(); err != nil {
		t.Fatal(err)
	}

	_ = json.Unmarshal([]byte(jsonTestSend), &testSendBlock)
	_ = json.Unmarshal([]byte(jsonTestReceive), &testReceiveBlock)
	_ = json.Unmarshal([]byte(jsonTestChangeRepresentative), &testChangeRepresentative)

	verfiyfy := process.NewLedgerVerifier(service.ledgerService.Ledger)
	if err := verfiyfy.BlockProcess(&testSendBlock); err != nil {
		t.Fatal(err)
	}
	if err := verfiyfy.BlockProcess(&testReceiveBlock); err != nil {
		t.Fatal(err)
	}
	if err := verfiyfy.BlockProcess(&testChangeRepresentative); err != nil {
		t.Fatal(err)
	}

	if err := service.netService.Init(); err != nil {
		t.Fatal(err)
	}
	if err := service.consensusService.Init(); err != nil {
		t.Fatal(err)
	}
	if err := service.rPCService.Init(); err != nil {
		t.Fatal(err)
	}
	if err := service.ledgerService.Start(); err != nil {
		t.Fatal(err)
	}
	if err := service.netService.Start(); err != nil {
		t.Fatal(err)
	}
	if err := service.consensusService.Start(); err != nil {
		t.Fatal(err)
	}
	if err := service.rPCService.Start(); err != nil {
		t.Fatal(err)
	}
}

func closeServer(service *Service, t *testing.T) {
	if err := service.rPCService.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := service.consensusService.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := service.netService.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := service.ledgerService.Stop(); err != nil {
		t.Fatal(err)
	}
	//if err := os.RemoveAll(service.dir); err != nil {
	//	t.Fatal(err)
	//}
}
