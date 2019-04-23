package test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"

	"github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/rpc"
)

var beneficialPledge = "dd20a386c735a077206619eca312072ad19266a161b8269d2f9b49785a3afde95d56683fb3f03c259dc0a703645ae0fb4f883d492d059665e4dee58c56c4e853"

func startService_Pledge(t *testing.T) (func(t *testing.T), *rpc.Client, *services.LedgerService) {

	dir := filepath.Join(config.DefaultDataDir(), "pledge")
	cfgFile, _ := config.DefaultConfig(dir)
	ls := services.NewLedgerService(cfgFile)
	err := ls.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = ls.Start()
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal([]byte(jsonTestSend), &testPledgeSendBlock)
	_ = json.Unmarshal([]byte(jsonTestReceive), &testPledgeReceiveBlock)
	l := ls.Ledger
	verifier := process.NewLedgerVerifier(l)
	p, _ := verifier.Process(&testPledgeSendBlock)
	if p != process.Progress {
		t.Fatal("process send block error")
	}
	p, _ = verifier.Process(&testPledgeReceiveBlock)
	if p != process.Progress {
		t.Fatal("process receive block error")
	}
	rPCService, err := services.NewRPCService(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	err = rPCService.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = rPCService.Start()
	if err != nil {
		t.Fatal(err)
	}
	client, err := rPCService.RPC().Attach()
	if err != nil {
		t.Fatal(err)
	}
	sqliteService, err := services.NewSqliteService(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	err = sqliteService.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = sqliteService.Start()
	if err != nil {
		t.Fatal(err)
	}
	walletService := services.NewWalletService(cfgFile)
	err = walletService.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = walletService.Start()
	if err != nil {
		t.Fatal(err)
	}
	return func(t *testing.T) {
		if client != nil {
			client.Close()
		}
		if err := ls.Stop(); err != nil {
			t.Fatal(err)
		}
		if err := sqliteService.Stop(); err != nil {
			t.Fatal(err)
		}
		if err := walletService.Stop(); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}, client, ls
}

func TestPledge(t *testing.T) {

	teardownTestCase, client, ls := startService_Pledge(t)
	defer teardownTestCase(t)
	pledgeBytes, err := hex.DecodeString(testPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	p := types.NewAccount(pledgeBytes)
	beneficialPledgeBytes, err := hex.DecodeString(beneficialPledge)
	if err != nil {
		t.Fatal(err)
	}
	b := types.NewAccount(beneficialPledgeBytes)
	am := types.StringToBalance("1000000000")
	NEP5tTxId := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	pledgeParam := api.PledgeParam{
		Beneficial:    b.Address(),
		PledgeAddress: p.Address(),
		Amount:        am,
		PType:         "vote",
		NEP5TxId:      NEP5tTxId,
	}

	send := types.StateBlock{}
	err = client.Call(&send, "pledge_getPledgeBlock", &pledgeParam)
	if err != nil {
		t.Fatal(err)
	}
	sendHash := send.GetHash()
	send.Signature = p.Sign(sendHash)
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()

	reward := types.StateBlock{}
	err = client.Call(&reward, "pledge_getPledgeRewardBlock", &send)

	if err != nil {
		t.Fatal(err)
	}
	reward.Signature = b.Sign(reward.GetHash())
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()

	err = client.Call(nil, "ledger_process", &send)
	if err != nil {
		t.Fatal(err)
	}

	err = client.Call(nil, "ledger_process", &reward)
	if err != nil {
		t.Fatal(err)
	}
	tm, err := ls.Ledger.GetAccountMeta(b.Address())
	if !tm.CoinVote.Equal(am) {
		t.Fatal("get voting fail")
	}

	withdrawPledgeParam := api.WithdrawPledgeParam{
		Beneficial: b.Address(),
		Amount:     am,
		PType:      "vote",
	}

	send1 := types.StateBlock{}
	err = client.Call(&send1, "pledge_getWithdrawPledgeBlock", &withdrawPledgeParam)
	if err != nil {
		t.Fatal(err)
	}
	send1Hash := send1.GetHash()
	send1.Signature = b.Sign(send1Hash)
	var w1 types.Work
	worker1, _ := types.NewWorker(w1, send1.Root())
	send1.Work = worker1.NewWork()

	reward1 := types.StateBlock{}
	err = client.Call(&reward1, "pledge_getWithdrawRewardBlock", &send1)

	if err == nil {
		t.Fatal("should return error: pledge is not ready")
	}
	//reward1.Signature = p.Sign(reward1.GetHash())
	//var w3 types.Work
	//worker3, _ := types.NewWorker(w3, reward1.Root())
	//reward1.Work = worker3.NewWork()
	//
	//err = client.Call(nil, "ledger_process", &send1)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//err = client.Call(nil, "ledger_process", &reward1)
	//if err != nil {
	//	t.Fatal(err)
	//}

	//begin test nep5 txid,the nep5 Txid of each coinage contract should be different
	pledgeParamTest := api.PledgeParam{
		Beneficial:    b.Address(),
		PledgeAddress: p.Address(),
		Amount:        am,
		PType:         "vote",
		NEP5TxId:      NEP5tTxId,
	}

	sendTest := types.StateBlock{}
	err = client.Call(&sendTest, "pledge_getPledgeBlock", &pledgeParamTest)
	if err != nil {
		t.Fatal(err)
	}
	sendTestHash := sendTest.GetHash()
	sendTest.Signature = p.Sign(sendTestHash)
	var wTest types.Work
	workerTest, _ := types.NewWorker(wTest, sendTest.Root())
	sendTest.Work = workerTest.NewWork()

	rewardTest := types.StateBlock{}
	err = client.Call(&rewardTest, "pledge_getPledgeRewardBlock", &sendTest)

	if err == nil {
		t.Fatal("should return error:invalid nep5 tx id")
	}
}
