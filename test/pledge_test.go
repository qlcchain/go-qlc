// +build integrate

package test

import (
	"encoding/hex"
	"testing"

	"github.com/qlcchain/go-qlc/crypto/random"

	"github.com/qlcchain/go-qlc/ledger/process"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"
)

var beneficialPledge = "dd20a386c735a077206619eca312072ad19266a161b8269d2f9b49785a3afde95d56683fb3f03c259dc0a703645ae0fb4f883d492d059665e4dee58c56c4e853"

func TestPledge(t *testing.T) {
	teardownTestCase, client, ls, err := generateChain()
	defer func() {
		if err := teardownTestCase(); err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
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
	NEP5tTxId := random.RandomHexString(32)
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

	lv := process.NewLedgerVerifier(ls.Ledger)
	r, _ := lv.Process(&send)
	if r != process.Progress {
		t.Fatal("process send block error")
	}
	r, _ = lv.Process(&reward)
	if r != process.Progress {
		t.Fatal("process reward block error")
	}

	tm, err := ls.Ledger.GetAccountMeta(b.Address())
	if !tm.CoinVote.Equal(am) {
		t.Fatal("get voting fail")
	}

	withdrawPledgeParam := api.WithdrawPledgeParam{
		Beneficial: b.Address(),
		Amount:     am,
		PType:      "vote",
		NEP5TxId:   NEP5tTxId,
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
	//err = client.Call(nil, "ledger_process", &reward1)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Log(reward1.String())

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
