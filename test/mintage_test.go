package test

import (
	"encoding/hex"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/rpc/api"
	"path/filepath"
	"testing"
)

var beneficial = "dd20a386c735a077206619eca312072ad19266a161b8269d2f9b49785a3afde95d56683fb3f03c259dc0a703645ae0fb4f883d492d059665e4dee58c56c4e853"

func TestMintage(t *testing.T) {
	eventBus := event.New()
	dir := filepath.Join(config.QlcTestDataDir(), "consensus", uuid.New().String())
	cfgFile, _ := config.DefaultConfig(dir)
	ls := services.NewLedgerService(cfgFile, eventBus)
	err := ls.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = ls.Start()
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal([]byte(jsonTestSend), &testSendBlock)
	_ = json.Unmarshal([]byte(jsonTestReceive), &testReceiveBlock)
	l := ls.Ledger
	verifier := process.NewLedgerVerifier(l)
	p, _ := verifier.Process(&testSendBlock)
	if p != process.Progress {
		t.Fatal("process send block error")
	}
	p, _ = verifier.Process(&testReceiveBlock)
	if p != process.Progress {
		t.Fatal("process receive block error")
	}
	rPCService, err := services.NewRPCService(cfgFile, eventBus)
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
	sqliteService, err := services.NewSqliteService(cfgFile, eventBus)
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
	defer func() {
		if client != nil {
			client.Close()
		}
		if err := ls.Stop(); err != nil {
			t.Fatal(err)
		}
		if err := sqliteService.Stop(); err != nil {
			t.Fatal(err)
		}
		//if err := os.RemoveAll(dir); err != nil {
		//	t.Fatal(err)
		//}
	}()
	selfBytes, err := hex.DecodeString(testPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	s := types.NewAccount(selfBytes)
	beneficialBytes, err := hex.DecodeString(testPrivateKey)
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
}
