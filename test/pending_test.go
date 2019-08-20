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
	"github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
)

type Services struct {
	ledgerService    *services.LedgerService
	netService       *services.P2PService
	consensusService *services.ConsensusService
	rPCService       *services.RPCService
	dir              string
}

func TestPending(t *testing.T) {

	testBytes, _ := hex.DecodeString(testPrivateKey)
	testAccount := types.NewAccount(testBytes)

	//node1
	dir1 := filepath.Join(config.QlcTestDataDir(), "transaction", uuid.New().String())
	cfgFile1, _ := config.DefaultConfig(dir1)
	cfgFile1.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
	cfgFile1.P2P.Discovery.DiscoveryInterval = 3
	cfgFile1.P2P.SyncInterval = 30
	cfgFile1.LogLevel = "info"
	cfgFile1.RPC.Enable = true
	cfgFile1.RPC.HTTPEndpoint = "tcp4://0.0.0.0:19735"
	cfgFile1.P2P.BootNodes = []string{}

	cfgFile1Byte, _ := json.Marshal(cfgFile1)
	fmt.Println("node1 config \n", string(cfgFile1Byte))
	node := new(Services)
	initNode(node, cfgFile1, nil, t)
	client, err := node.rPCService.RPC().Attach()
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
		Token:   common.ChainToken(),
		Link:    tAddress.ToHash(),
	}

	sendBlock, err := node.ledgerService.Ledger.GenerateSendBlock(&sb, sAmount, testAccount.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("sendBlock ", sendBlock.GetHash())
	lv := process.NewLedgerVerifier(node.ledgerService.Ledger)
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
	if err := client.Call(&receiverBlock2, "ledger_generateReceiveBlock", &sendBlock); err != nil {
		t.Fatal(err)
	}

	if err := node.ledgerService.Ledger.AddBlockCache(&receiverBlock2); err != nil {
		t.Fatal(err)
	}

	count1 := 0
	err = node.ledgerService.Ledger.SearchPending(receiverBlock2.Address, func(key *types.PendingKey, value *types.PendingInfo) error {
		count1++
		return nil
	})
	if err != nil || count1 != 0 {
		t.Fatal("error pending")
	}
	count2 := 0
	err = node.ledgerService.Ledger.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		count2++
		return nil
	})
	if err != nil || count2 != 0 {
		t.Fatal("error pending")
	}

	// wait roll back
	time.Sleep(70 * time.Second)

	count3 := 0
	err = node.ledgerService.Ledger.SearchPending(receiverBlock2.Address, func(key *types.PendingKey, value *types.PendingInfo) error {
		count3++
		return nil
	})
	if err != nil || count3 != 1 {
		t.Fatal("error pending")
	}

	count4 := 0
	err = node.ledgerService.Ledger.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		count4++
		return nil
	})
	if err != nil || count4 != 1 {
		t.Fatal("error pending")
	}
}

func initNode(service *Services, cfg *config.Config, accounts []*types.Account, t *testing.T) {
	logService := log.NewLogService(cfg)
	_ = logService.Init()
	var err error
	service.ledgerService = services.NewLedgerService(cfg)
	service.consensusService = services.NewConsensusService(cfg, accounts)
	service.rPCService, err = services.NewRPCService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	service.netService, err = services.NewP2PService(cfg)
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

func closeServer(service *Services, t *testing.T) {
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

func signAndWork(block *types.StateBlock, account *types.Account) error {
	var w types.Work
	worker, err := types.NewWorker(w, block.Root())
	if err != nil {
		return err
	}
	block.Work = worker.NewWork()
	block.Signature = account.Sign(block.GetHash())

	return nil
}
