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
	"testing"
	"time"

	"github.com/google/uuid"
	ss "github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/rpc/api"
)

var (
	services1 = new(Service)
	services2 = new(Service)
)

type Service struct {
	ledgerService    *ss.LedgerService
	netService       *ss.P2PService
	consensusService *ss.ConsensusService
	rPCService       *ss.RPCService
	sqliteService    *ss.SqliteService
}

func TestTransaction(t *testing.T) {
	//node1
	dir1 := filepath.Join(config.QlcTestDataDir(), "consensus", uuid.New().String())
	fmt.Println(dir1)
	cfgFile1, _ := config.DefaultConfig(dir1)
	cfgFile1.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
	cfgFile1.P2P.Discovery.DiscoveryInterval = 3
	cfgFile1.P2P.SyncInterval = 30
	cfgFile1.LogLevel = "info"
	cfgFile1.RPC.Enable = true
	b1 := "/ip4/0.0.0.0/tcp/19741/ipfs/" + cfgFile1.P2P.ID.PeerID
	cfgFile1Byte, _ := json.Marshal(cfgFile1)
	fmt.Println("node1 config \n", string(cfgFile1Byte))

	// node1
	dir2 := filepath.Join(config.QlcTestDataDir(), "consensus", uuid.New().String())
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

	cfgFile2Byte, _ := json.Marshal(cfgFile1)
	fmt.Println("node2 config \n", string(cfgFile2Byte))

	fmt.Println(" start node1....")
	bytes, _ := hex.DecodeString(testPrivateKey)
	account := types.NewAccount(bytes)
	initNode(services1, cfgFile1, []*types.Account{account}, t)

	fmt.Println(" start node2....")
	initNode(services2, cfgFile2, nil, t)

	time.Sleep(10 * time.Second)

	err := services1.ledgerService.Ledger.GetStateBlocks(func(block *types.StateBlock) error {
		fmt.Println(block)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	client, err := services1.rPCService.RPC().Attach()

	var sendBlock types.StateBlock
	para := api.APISendBlockPara{
		From:      account.Address(),
		To:        mock.Address(),
		Amount:    types.Balance{Int: big.NewInt(100)},
		TokenName: "QLC",
	}
	if err := client.Call(&sendBlock, "ledger_generateSendBlock", &para, testPrivateKey); err != nil {
		t.Fatal(err)
	}
	var h types.Hash
	if err := client.Call(&h, "ledger_process", &sendBlock); err != nil {
		t.Fatal(err)
	}
	//select {}
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
	service.sqliteService, err = ss.NewSqliteService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	service.netService, err = ss.NewP2PService(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := service.sqliteService.Init(); err != nil {
		t.Fatal(err)
	}
	if err := service.ledgerService.Init(); err != nil {
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
	if err := service.sqliteService.Start(); err != nil {
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
}

func TestLedger_HashConvertToAddress(t *testing.T) {
	s := "6c0b2cdd533ee3a21668f199e111f6c8614040e60e70a73ab6c8da036f2a7ad7"
	h := new(types.Hash)
	h.Of(s)
	fmt.Println(types.Address(*h))
}
