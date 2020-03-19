package api

import (
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/vm/contract/abi"

	"github.com/qlcchain/go-qlc/crypto/random"

	"github.com/google/uuid"
	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCasePledge(t *testing.T) (func(t *testing.T), *ledger.Ledger, *NEP5PledgeApi, *types.Account, *types.Account, *process.LedgerVerifier) {
	t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	_ = cc.Init(nil)
	l := ledger.NewLedger(cm.ConfigFile)
	cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)

	pledgeAPI := NewNEP5PledgeAPI(cm.ConfigFile, l)
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
	}, l, pledgeAPI, ac1, ac2, lv
}

func TestNEP5PledgeApi(t *testing.T) {
	teardownTestCase, l, pledgeAPI, ac1, ac2, lv := setupTestCasePledge(t)
	defer teardownTestCase(t)
	am := types.StringToBalance("1000000000")
	NEP5tTxId := random.RandomHexString(32)
	pledgeParam := &PledgeParam{
		Beneficial:    ac2.Address(),
		PledgeAddress: ac1.Address(),
		Amount:        am,
		PType:         "vote",
		NEP5TxId:      NEP5tTxId,
	}

	_, err := pledgeAPI.GetPledgeData(pledgeParam)
	if err != nil {
		t.Fatal(err)
	}

	send, err := pledgeAPI.GetPledgeBlock(pledgeParam)
	if err != nil {
		t.Fatal(err)
	}
	sendHash := send.GetHash()
	send.Signature = ac1.Sign(sendHash)
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()

	reward, err := pledgeAPI.GetPledgeRewardBlock(send)
	if err != nil {
		t.Fatal(err)
	}
	reward.Signature = ac2.Sign(reward.GetHash())
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()

	result, err := lv.Process(send)
	if result != process.Progress {
		t.Fatal("block check error for send")
	}
	_, err = pledgeAPI.GetPledgeRewardBlockBySendHash(sendHash)
	if err != nil {
		t.Fatal(err)
	}
	result, err = lv.Process(reward)
	if result != process.Progress {
		t.Fatal("block check error for send")
	}

	tm, err := l.GetAccountMeta(ac2.Address())
	if err != nil {
		t.Fatal(err)
	}
	if !tm.CoinVote.Equal(am) {
		t.Fatal("get voting fail")
	}
	withdrawPledgeParam := &WithdrawPledgeParam{
		Beneficial: ac2.Address(),
		Amount:     am,
		PType:      "vote",
		NEP5TxId:   NEP5tTxId,
	}

	_, err = pledgeAPI.GetWithdrawPledgeData(withdrawPledgeParam)
	if err != nil {
		t.Fatal(err)
	}
	send1, err := pledgeAPI.GetWithdrawPledgeBlock(withdrawPledgeParam)
	if err != nil {
		t.Fatal(err)
	}
	send1Hash := send1.GetHash()
	send1.Signature = ac2.Sign(send1Hash)
	var w1 types.Work
	worker1, _ := types.NewWorker(w1, send1.Root())
	send1.Work = worker1.NewWork()

	_, err = pledgeAPI.GetWithdrawRewardBlock(send1)
	if err == nil {
		t.Fatal("should return error: pledge is not ready")
	}
	_, err = pledgeAPI.GetWithdrawRewardBlockBySendHash(send1Hash)
	if err == nil {
		t.Fatal("should return error: pledge is not ready")
	}
	info := &abi.NEP5PledgeInfo{
		PType:         uint8(abi.Network),
		Amount:        big.NewInt(100),
		WithdrawTime:  time.Now().Unix(),
		Beneficial:    mock.Address(),
		PledgeAddress: mock.Address(),
		NEP5TxId:      mock.Hash().String(),
	}
	if data, err := info.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		if info2, err := pledgeAPI.ParsePledgeInfo(data); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(info, info2) {
			t.Fatalf("exp: %v, act: %v", info, info2)
		}
	}
	pi := pledgeAPI.GetPledgeInfosByPledgeAddress(ac1.Address())
	if len(pi.PledgeInfo) != 1 {
		t.Fatal("pledger info error")
	}
	_, err = pledgeAPI.GetPledgeBeneficialTotalAmount(ac1.Address())
	if err != nil {
		t.Fatal(err)
	}
	pb := pledgeAPI.GetBeneficialPledgeInfosByAddress(ac2.Address())
	if len(pb.PledgeInfo) != 1 {
		t.Fatal("beneficial pledger info error")
	}
	pi1, err := pledgeAPI.GetBeneficialPledgeInfos(ac2.Address(), "vote")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pi1)
	_, err = pledgeAPI.GetPledgeBeneficialAmount(ac2.Address(), "vote")
	if err != nil {
		t.Fatal(err)
	}
	_, err = pledgeAPI.GetPledgeInfo(withdrawPledgeParam)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pledgeAPI.GetPledgeInfoWithNEP5TxId(withdrawPledgeParam)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pledgeAPI.GetPledgeInfoWithTimeExpired(withdrawPledgeParam)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pledgeAPI.GetAllPledgeInfo()
	if err != nil {
		t.Fatal(err)
	}
	_, err = pledgeAPI.GetTotalPledgeAmount()
	if err != nil {
		t.Fatal(err)
	}
}
