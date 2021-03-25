package api

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
)

var (
	mAccount1      = mock.Account3
	mAddress1      = mAccount1.Address()
	mAccount2      = mock.Account2
	mAddress2      = mAccount2.Address()
	pledgeAmount   = types.Balance{Int: big.NewInt(1000000)}
	withdrawAmount = types.Balance{Int: big.NewInt(500000)}
)

func setupQGasSwapTestCase(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *QGasSwapAPI) {
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := chainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewQGasSwapAPI(l, cm.ConfigFile)

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(mock.MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}

	for i := range blocks {
		block := blocks[i]
		if err := verifier.BlockProcess(block); err != nil {
			t.Fatal(err)
		}
	}

	if err := verifier.BlockProcess(&mock.TestQGasOpenBlock); err != nil {
		t.Fatal(err)
	}

	return func(t *testing.T) {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
		_ = cc.Stop()
	}, verifier, api
}

func TestQGasSwapAPI_GetPledgeSendBlock(t *testing.T) {
	teardownTestCase, verifier, api := setupQGasSwapTestCase(t)
	defer teardownTestCase(t)

	// pledge
	pledgeParams := &QGasPledgeParams{
		FromAddress: mAddress1,
		Amount:      pledgeAmount,
		ToAddress:   mAddress2,
	}
	pledgeSendBlk, err := api.GetPledgeSendBlock(pledgeParams)
	if err != nil {
		t.Fatal(err)
	}
	signAndWork(pledgeSendBlk, mAccount1, t)
	if _, err = verifier.Process(pledgeSendBlk); err != nil {
		t.Fatal(err)
	}
	pledgeRewardBlk, err := api.GetPledgeRewardBlock(pledgeSendBlk.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	signAndWork(pledgeRewardBlk, mAccount2, t)
	if _, err = verifier.Process(pledgeRewardBlk); err != nil {
		t.Fatal(err)
	}
	if r, err := api.GetAllSwapInfos(10, 0, nil); err != nil || len(r) != 1 {
		t.Fatal(err)
	}

	// withdraw
	withdrawParams := &QGasWithdrawParams{
		ToAddress:   mAddress1,
		Amount:      withdrawAmount,
		FromAddress: mAddress2,
	}
	withdrawSendBlk, err := api.GetWithdrawSendBlock(withdrawParams)
	if err != nil {
		t.Fatal(err)
	}
	signAndWork(withdrawSendBlk, mAccount2, t)
	if _, err = verifier.Process(withdrawSendBlk); err != nil {
		t.Fatal(err)
	}
	withdrawRewardBlk, err := api.GetWithdrawRewardBlock(withdrawSendBlk.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	signAndWork(withdrawRewardBlk, mAccount1, t)
	if _, err = verifier.Process(withdrawRewardBlk); err != nil {
		t.Fatal(err)
	}
	trueV := true
	falseV := false
	if r, err := api.GetAllSwapInfos(10, 0, nil); err != nil || len(r) != 2 {
		t.Fatal(err)
	}
	if r, err := api.GetAllSwapInfos(10, 0, &trueV); err != nil || len(r) != 1 {
		t.Fatal(err)
	}
	if r, err := api.GetAllSwapInfos(10, 0, &falseV); err != nil || len(r) != 1 {
		t.Fatal(err)
	}
	if r, err := api.GetSwapInfosByAddress(mAddress1, 10, 0, nil); err != nil || len(r) != 2 {
		t.Fatal(err)
	}
	if r, err := api.GetSwapInfosByAddress(mAddress1, 10, 0, &trueV); err != nil || len(r) != 1 {
		t.Fatal(err)
	}
	if r, err := api.GetSwapInfosByAddress(mAddress1, 10, 0, &falseV); err != nil || len(r) != 1 {
		t.Fatal(err)
	}
	if r, err := api.GetSwapAmountByAddress(mAddress1); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
	if r, err := api.GetSwapAmount(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
}

func TestQGasSwapAPI_Check(t *testing.T) {
	teardownTestCase, _, api := setupQGasSwapTestCase(t)
	defer teardownTestCase(t)

	if _, err := api.GetPledgeSendBlock(nil); err != ErrParameterNil {
		t.Fatal(err)
	}

	pledgeParams := &QGasPledgeParams{
		FromAddress: mock.Address(),
		Amount:      pledgeAmount,
		ToAddress:   mAddress2,
	}
	if _, err := api.GetPledgeSendBlock(pledgeParams); err == nil {
		t.Fatal(err)
	}

}

func signAndWork(block *types.StateBlock, account *types.Account, t *testing.T) {
	var w types.Work
	worker, err := types.NewWorker(w, block.Root())
	if err != nil {
		t.Fatal(err)
	}
	block.Work = worker.NewWork()
	block.Signature = account.Sign(block.GetHash())
}
