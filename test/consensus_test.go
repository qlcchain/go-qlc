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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/event"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/test/mock"
)

var (
	seed         = "EE8B2F389880D23D1656652EDE146149A7A2E9DDCE2A95D511C55354CBB5ED50"
	ac           *types.Account
	token        = common.ChainToken()
	genesisBlock *types.StateBlock
)

func TestConsensus(t *testing.T) {
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "consensus", uuid.New().String())
	cfgFile, _ := config.DefaultConfig(dir)
	cfgFile.P2P.Listen = "/ip4/0.0.0.0/tcp/19740"
	cfgFile.P2P.Discovery.MDNSEnabled = false
	cfgFile.P2P.BootNodes = []string{}
	b := "/ip4/0.0.0.0/tcp/19740/ipfs/" + cfgFile.P2P.ID.PeerID
	//new ledger
	l := services.NewLedgerService(cfgFile)

	eventBus := event.New()
	//start bootNode
	node, err := p2p.NewQlcService(cfgFile, eventBus)
	err = node.Start()
	if err != nil {
		//t.Fatal(err)
	}

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "consensus", uuid.New().String())
	cfgFile1, _ := config.DefaultConfig(dir1)
	cfgFile1.P2P.Listen = "/ip4/0.0.0.0/tcp/19741"
	cfgFile1.P2P.BootNodes = []string{b}
	cfgFile1.P2P.Discovery.MDNSEnabled = false
	cfgFile1.P2P.Discovery.DiscoveryInterval = 3
	//new ledger
	ledger1 := services.NewLedgerService(cfgFile1)

	//storage genesisBlock
	creatGenesisBlock(ledger1.Ledger)

	//start node1
	node1, err := p2p.NewQlcService(cfgFile1, eventBus)
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}
	//import wallet
	//importWallet(cfgFile1)
	//new dpos service
	var accs []*types.Account
	accs = append(accs, ac)
	consensusService1, err := consensus.NewDPoS(cfgFile1, accs, eventBus)
	//start node1 dpos service
	err = consensusService1.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = consensusService1.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "consensus", uuid.New().String())
	cfgFile2, _ := config.DefaultConfig(dir2)
	cfgFile2.P2P.Listen = "/ip4/0.0.0.0/tcp/19742"
	cfgFile2.P2P.BootNodes = []string{b}
	cfgFile2.P2P.Discovery.MDNSEnabled = false
	cfgFile2.P2P.Discovery.DiscoveryInterval = 3
	cfgFile2.PerformanceEnabled = true

	//new ledger
	ledger2 := services.NewLedgerService(cfgFile2)
	//storage genesisBlock
	creatGenesisBlock(ledger2.Ledger)
	//start node2
	node2, err := p2p.NewQlcService(cfgFile2, eventBus)
	err = node2.Start()
	if err != nil {
		t.Fatal(err)
	}

	consensusService2 := services.NewDPosService(cfgFile2, nil, eventBus)
	//start node2 dpos service
	err = consensusService2.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = consensusService2.Start()
	if err != nil {
		t.Fatal(err)
	}

	//remove test file
	defer func() {
		err := l.Ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = ledger1.Ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = consensusService1.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = ledger2.Ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = consensusService2.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = node.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = node1.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = node2.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(config.QlcTestDataDir())
		if err != nil {
			t.Fatal(err)
		}
	}()

	ticker1 := time.NewTicker(60 * time.Second)
	for {

		select {
		case <-ticker1.C:
			fmt.Println("connect peer timeout")
			return
		default:
			time.Sleep(5 * time.Millisecond)
		}
		_, err = node1.Node().StreamManager().RandomPeer()
		if err != nil {
			continue
		}
		break
	}
	dst := mock.Account()
	addr := dst.Address()
	sb := types.StateBlock{
		Address: ac.Address(),
		Token:   token,
		Link:    addr.ToHash(),
		Message: types.ZeroHash,
	}
	send, err := ledger1.Ledger.GenerateSendBlock(&sb, types.Balance{Int: big.NewInt(int64(1000))}, ac.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	verifier1 := process.NewLedgerVerifier(ledger1.Ledger)
	/**/ verifier1.Process(send)
	node1.Broadcast(common.PublishReq, send)
	time.Sleep(30 * time.Second)
	c, err := ledger2.Ledger.CountStateBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if c != 2 {
		t.Fatal("node2 block count not correct")
	}

	p, err := ledger2.Ledger.GetPerformanceTime(send.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	if p.T0 == 0 || p.T1 == 0 || p.T2 == 0 || p.T3 == 0 {
		t.Fatal("send block confirmed error")
	}
}

func creatGenesisBlock(l *ledger.Ledger) {
	sByte, _ := hex.DecodeString(seed)
	seed, _ := types.BytesToSeed(sByte)
	ac, _ = seed.Account(0)
	fmt.Println(ac.Address())
	genesisBlock = createBlock(*ac, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(60000000000000000))}, types.Hash(ac.Address()), ac.Address())
	verifier := process.NewLedgerVerifier(l)
	verifier.BlockProcess(genesisBlock)
}

func createBlock(ac types.Account, pre types.Hash, token types.Hash, balance types.Balance, link types.Hash, rep types.Address) *types.StateBlock {
	blk := new(types.StateBlock)
	blk.Type = types.State
	blk.Address = ac.Address()
	blk.Previous = pre
	blk.Token = token
	blk.Balance = balance
	blk.Link = link
	blk.Representative = rep
	blk.Signature = ac.Sign(blk.GetHash())
	var w types.Work
	worker, _ := types.NewWorker(w, blk.Root())
	blk.Work = worker.NewWork()
	return blk
}
