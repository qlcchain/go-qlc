// +build integrate

package test

import (
	"encoding/hex"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"
)

var beneficial = "dd20a386c735a077206619eca312072ad19266a161b8269d2f9b49785a3afde95d56683fb3f03c259dc0a703645ae0fb4f883d492d059665e4dee58c56c4e853"

func TestMintage(t *testing.T) {
	teardownTestCase, client, _, err := generateChain()
	defer func() {
		if err := teardownTestCase(); err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	selfBytes, err := hex.DecodeString(testPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	s := types.NewAccount(selfBytes)
	beneficialBytes, err := hex.DecodeString(beneficial)
	if err != nil {
		t.Fatal(err)
	}
	b := types.NewAccount(beneficialBytes)
	NEP5tTxId := "asfafjjfwejwjfkagjksgjisogwij134l09afjakjf"
	mintageParam := api.MintageParams{
		SelfAddr:    s.Address(),
		PrevHash:    testReceiveBlock.GetHash(),
		TokenName:   "QN",
		TotalSupply: "1000000",
		TokenSymbol: "QN",
		Decimals:    uint8(8),
		Beneficial:  b.Address(),
		NEP5TxId:    NEP5tTxId,
	}
	send := types.StateBlock{}
	err = client.Call(&send, "mintage_getMintageBlock", &mintageParam)
	if err != nil {
		t.Fatal(err)
	}

	sendHash := send.GetHash()
	send.Signature = s.Sign(sendHash)
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()
	err = client.Call(nil, "ledger_process", &send)
	if err != nil {
		t.Fatal(err)
	}
	reward := types.StateBlock{}
	err = client.Call(&reward, "mintage_getRewardBlock", &send)

	if err != nil {
		t.Fatal(err)
	}

	reward.Signature = b.Sign(reward.GetHash())
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()
	err = client.Call(nil, "ledger_process", &reward)
	if err != nil {
		t.Fatal(err)
	}
	var ts []*types.TokenInfo
	err = client.Call(&ts, "ledger_tokens")
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 3 {
		t.Fatal("token count error")
	}

	tokenId := reward.Token
	addr, err := types.HexToAddress(testAddress)
	if err != nil {
		t.Fatal(err)
	}
	withdrawMintageParam := api.WithdrawParams{
		SelfAddr: addr, TokenId: tokenId}

	withdrawSend := types.StateBlock{}
	err = client.Call(&withdrawSend, "mintage_getWithdrawMintageBlock", &withdrawMintageParam)
	if err != nil {
		t.Fatal(err)
	}

	withdrawSendHash := withdrawSend.GetHash()
	withdrawSend.Signature = s.Sign(withdrawSendHash)
	var w3 types.Work
	worker3, _ := types.NewWorker(w3, withdrawSend.Root())
	withdrawSend.Work = worker3.NewWork()

	err = client.Call(nil, "ledger_process", &withdrawSend)
	if err != nil {
		t.Fatal(err)
	}

	withdrawReward := types.StateBlock{}
	err = client.Call(&withdrawReward, "mintage_getWithdrawRewardBlock", &withdrawSend)

	if err == nil {
		t.Fatal("pledge time should not expire yet")
	}

	/*In the offline test environment,can modify the pledge expiration
	  time,and then remove this comment. */

	/*	withdrawReward.Signature = s.Sign(withdrawReward.GetHash())
		var w4 types.Work
		worker4, _ := types.NewWorker(w4, withdrawReward.Root())
		withdrawReward.Work = worker4.NewWork()

		err = client.Call(nil, "ledger_process", &withdrawReward)
		if err != nil {
			t.Fatal(err)
		}
		var c map[string]uint64
		err = client.Call(&c, "ledger_blocksCount")
		if err != nil {
			t.Fatal(err)
		}
		count, err := ls.Ledger.CountStateBlocks()
		if err != nil {
			t.Fatal(err)
		}
		if count != 10 {
			t.Fatal("count block is error")
		}*/

	//begin test nep5 txid,the nep5 Txid of each coinage contract should be different
	mintageParamTest := api.MintageParams{
		SelfAddr:    s.Address(),
		PrevHash:    testReceiveBlock.GetHash(),
		TokenName:   "QA",
		TotalSupply: "10000000",
		TokenSymbol: "QA",
		Decimals:    uint8(8),
		Beneficial:  b.Address(),
		NEP5TxId:    NEP5tTxId,
	}
	sendTest := types.StateBlock{}
	err = client.Call(&sendTest, "mintage_getMintageBlock", &mintageParamTest)
	if err != nil {
		t.Fatal(err)
	}

	sendTestHash := sendTest.GetHash()
	sendTest.Signature = s.Sign(sendTestHash)
	var wTest types.Work
	workerTest, _ := types.NewWorker(wTest, sendTest.Root())
	sendTest.Work = workerTest.NewWork()
	err = client.Call(nil, "ledger_process", &sendTest)
	if err != nil {
		t.Fatal(err)
	}
	rewardTest := types.StateBlock{}
	err = client.Call(&rewardTest, "mintage_getRewardBlock", &sendTest)
	if err == nil {
		t.Fatal("should return error:invalid nep5 tx id")
	}

	//test
	mintageParamTestName := api.MintageParams{
		SelfAddr:    s.Address(),
		PrevHash:    testReceiveBlock.GetHash(),
		TokenName:   "QN",
		TotalSupply: "10000000",
		TokenSymbol: "QN",
		Decimals:    uint8(8),
		Beneficial:  b.Address(),
		NEP5TxId:    "abcfajfwigaighaigaieh",
	}
	sendTestName := types.StateBlock{}
	err = client.Call(&sendTestName, "mintage_getMintageBlock", &mintageParamTestName)
	if err == nil {
		t.Fatal("should return error:invalid token name(QN) or token symbol(QN)")
	}
}
