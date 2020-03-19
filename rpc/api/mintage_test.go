package api

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/vm/contract/abi"

	"github.com/qlcchain/go-qlc/common/topic"

	"github.com/google/uuid"
	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCaseMintage(t *testing.T) (func(t *testing.T), *MintageAPI, []*types.StateBlock, *types.Account, *types.Account, *process.LedgerVerifier) {
	t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	_ = cc.Init(nil)
	l := ledger.NewLedger(cm.ConfigFile)
	cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)

	mintageAPI := NewMintageApi(cm.ConfigFile, l)
	bs, ac1, ac2, _ := mock.BlockChainWithAccount(false)
	lv := process.NewLedgerVerifier(l)
	if err := lv.BlockProcess(bs[0]); err != nil {
		t.Fatal(err)
	}
	t.Log("bs hash", bs[0].GetHash())
	for _, b := range bs[1:] {
		if p, err := lv.Process(b); err != nil || p != process.Progress {
			t.Fatal(p, err)
		}
	}
	block, td := mock.GeneratePovBlock(nil, 0)
	if err := l.AddPovBlock(block, td); err != nil {
		t.Fatal(err)
	}
	if err := l.AddPovBestHash(block.GetHeight(), block.GetHash()); err != nil {
		t.Fatal(err)
	}
	if err := l.SetPovLatestHeight(block.GetHeight()); err != nil {
		t.Fatal(err)
	}
	return func(t *testing.T) {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		_ = cc.Stop()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, mintageAPI, bs, ac1, ac2, lv
}

func TestMintageAPI(t *testing.T) {
	teardownTestCase, mintageApi, bs, ac1, ac2, lv := setupTestCaseMintage(t)
	defer teardownTestCase(t)
	NEP5tTxId := "asfafjjfwejwjfkagjksgjisogwij134l09afjakjf"
	mintageParam := &MintageParams{
		SelfAddr:    ac1.Address(),
		PrevHash:    bs[5].GetHash(),
		TokenName:   "QN",
		TotalSupply: "1000000",
		TokenSymbol: "QN",
		Decimals:    uint8(8),
		Beneficial:  ac2.Address(),
		NEP5TxId:    NEP5tTxId,
	}
	_, err := mintageApi.GetMintageData(mintageParam)
	if err != nil {
		t.Fatal(err)
	}
	send, err := mintageApi.GetMintageBlock(mintageParam)
	if err != nil {
		t.Fatal(err)
	}

	sendHash := send.GetHash()
	send.Signature = ac1.Sign(sendHash)
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()

	result, err := lv.Process(send)
	if result != process.Progress {
		t.Fatal("block check error for send")
	}

	reward, err := mintageApi.GetRewardBlock(send)
	if err != nil {
		t.Fatal(err)
	}
	reward.Signature = ac2.Sign(reward.GetHash())
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()

	result, err = lv.Process(reward)
	if result != process.Progress {
		t.Fatal("block check error for reward")
	}

	tokenId := reward.Token
	_, err = mintageApi.GetWithdrawMintageData(tokenId)
	if err != nil {
		t.Fatal(err)
	}
	withdrawMintageParam := &WithdrawParams{SelfAddr: ac1.Address(), TokenId: tokenId}

	withdrawSend, err := mintageApi.GetWithdrawMintageBlock(withdrawMintageParam)
	if err != nil {
		t.Fatal(err)
	}

	withdrawSendHash := withdrawSend.GetHash()
	withdrawSend.Signature = ac1.Sign(withdrawSendHash)
	var w3 types.Work
	worker3, _ := types.NewWorker(w3, withdrawSend.Root())
	withdrawSend.Work = worker3.NewWork()

	result, err = lv.Process(withdrawSend)
	if result != process.Progress {
		t.Fatal("block check error for send")
	}

	_, err = mintageApi.GetWithdrawRewardBlock(withdrawSend)
	if err == nil {
		t.Fatal("pledge time should not expire yet")
	}

	if data, err := abi.MintageABI.PackVariable(
		abi.VariableNameToken, tokenId, "QN", "QN", big.NewInt(100), uint8(8), ac2.Address(),
		big.NewInt(100), time.Now().AddDate(0, 0, 5).Unix(),
		ac1.Address(), mintageParam.NEP5TxId); err != nil {
		t.Fatal(err)
	} else {
		tk, err := mintageApi.ParseTokenInfo(data)
		if err != nil {
			t.Fatal(err)
		}
		if tk.NEP5TxId != mintageParam.NEP5TxId {
			t.Fatal("token info error")
		}
	}
}
