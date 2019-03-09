package consensus

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/test/mock"
	"github.com/qlcchain/go-qlc/wallet"
)

var seed = "EE8B2F389880D23D1656652EDE146149A7A2E9DDCE2A95D511C55354CBB5ED50"
var password = "123456"
var address types.Address
var ac *types.Account
var token = common.QLCChainToken
var genesisBlock *types.StateBlock

func creatGenesisBlock(l *ledger.Ledger) {
	sByte, _ := hex.DecodeString(seed)
	seed, _ := types.BytesToSeed(sByte)
	ac, _ = seed.Account(0)
	fmt.Println(ac.Address())
	genesisBlock = createBlock(*ac, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(60000000000000000))}, types.Hash(ac.Address()), ac.Address())
	l.BlockProcess(genesisBlock)
}

func importWallet(cfg *config.Config) error {
	if len(seed) == 0 {
		return errors.New("invalid seed")
	}
	w := wallet.NewWalletService(cfg)
	addr, err := w.Wallet.NewWalletBySeed(seed, password)
	if err != nil {
		fmt.Println(err)
	}
	address = addr
	fmt.Printf("import seed[%s]  => %s success", seed, addr.String())
	return nil
}

func Test_Trx_Confirmed(t *testing.T) {
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "consensus", uuid.New().String())
	cfgFile, _ := config.DefaultConfig(dir)
	cfgFile.P2P.Listen = "/ip4/0.0.0.0/tcp/19740"
	cfgFile.Discovery.MDNS.Enabled = false
	cfgFile.P2P.BootNodes = []string{}
	b := "/ip4/0.0.0.0/tcp/19740/ipfs/" + cfgFile.ID.PeerID
	//new ledger
	l := ledger.NewLedgerService(cfgFile)

	//start bootNode
	node, err := p2p.NewQlcService(cfgFile)
	err = node.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "consensus", uuid.New().String())
	cfgFile1, _ := config.DefaultConfig(dir1)
	cfgFile1.P2P.Listen = "/ip4/0.0.0.0/tcp/19741"
	cfgFile1.P2P.BootNodes = []string{b}
	cfgFile1.Discovery.MDNS.Enabled = false
	cfgFile1.Discovery.DiscoveryInterval = 3
	//new ledger
	ledger1 := ledger.NewLedgerService(cfgFile1)

	//storage genesisBlock
	creatGenesisBlock(ledger1.Ledger)

	//start node1
	node1, err := p2p.NewQlcService(cfgFile1)
	err = node1.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}
	//import wallet
	importWallet(cfgFile1)
	//new dpos service
	consensusService1, err := NewDposService(cfgFile1, node1, address, password)
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
	cfgFile2.Discovery.MDNS.Enabled = false
	cfgFile2.Discovery.DiscoveryInterval = 3
	cfgFile2.PerformanceTest.Enabled = true

	//new ledger
	ledger2 := ledger.NewLedgerService(cfgFile2)
	//storage genesisBlock
	creatGenesisBlock(ledger2.Ledger)
	//start node2
	node2, err := p2p.NewQlcService(cfgFile2)
	err = node2.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = node2.Start()
	if err != nil {
		t.Fatal(err)
	}

	consensusService2, err := NewDposService(cfgFile2, node2, types.ZeroAddress, "")
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
		err = consensusService1.wallet.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = ledger2.Ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = consensusService2.wallet.Close()
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
			t.Fatal("connect peer timeout")
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
	send, err := ledger2.Ledger.GenerateSendBlock(ac.Address(), token, dst.Address(), types.Balance{Int: big.NewInt(int64(1000))}, ac.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	consensusService1.ledger.Process(send)
	node1.Broadcast(p2p.PublishReq, send)
	time.Sleep(30 * time.Second)
	c, err := consensusService2.ledger.CountStateBlocks()
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
