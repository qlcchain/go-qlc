/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/rpc/api"
)

var (
	logger     = log.NewLogger("main")
	l          *chain.LedgerService
	rs         *chain.RPCService
	qlcAccount *types.Account
	gasAccount *types.Account
	verifier   *process.LedgerVerifier
)

func main() {
	qlcPri := flag.String("qlcAccount", "", "")
	gasPri := flag.String("gasAccount", "", "")
	testnet := flag.Bool("testnet", false, "")

	flag.Parse()
	prk, err := hex.DecodeString(*qlcPri)
	qlcAccount = types.NewAccount(prk)
	prk2, err := hex.DecodeString(*gasPri)
	gasAccount = types.NewAccount(prk2)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	dir := filepath.Join(config.QlcTestDataDir(), "cmd", uuid.New().String())
	if *testnet {
		dir = filepath.Join(config.DefaultDataDir())
	}
	cm := config.NewCfgManager(dir)
	cfg, err := cm.Load()
	cfg.RPC.Enable = true
	if cfg.RPC.Enable == false {
		return
	}
	_ = cm.Save()

	l = chain.NewLedgerService(cm.ConfigFile)
	if err := l.Init(); err != nil {
		return
	}
	if err := l.Start(); err != nil {
		logger.Fatal(err)
	}
	logger.Info("ledger started")
	verifier = process.NewLedgerVerifier(l.Ledger)
	//initData(l.Ledger)

	w := chain.NewWalletService(cm.ConfigFile)
	if err := w.Init(); err != nil {
		return
	}
	if err := w.Start(); err != nil {
		logger.Fatal(err)
	}
	logger.Info("wallet started")

	rs, err = chain.NewRPCService(cm.ConfigFile)
	if err != nil {
		logger.Fatal(err)
		return
	}
	if err = rs.Init(); err != nil {
		logger.Fatal(err)
	}
	if err = rs.Start(); err != nil {
		logger.Fatal(err)
	}
	logger.Info("rpc started")
	rollback()
	defer func() {
		l.Stop()
		w.Stop()
		rs.Stop()
		if *testnet {
			fmt.Println()
		} else {
			os.RemoveAll(dir)
			fmt.Println("delete ", dir)
		}
	}()
	s := <-c
	fmt.Println("Got signal: ", s)
}

func rollback() {
	accInfo, _ := l.Ledger.GetAccountMeta(qlcAccount.Address())
	fmt.Println(accInfo)
	tokenInfo, _ := l.Ledger.GetTokenMeta(qlcAccount.Address(), config.ChainToken())
	fmt.Println(tokenInfo)

	client, err := rs.RPC().Attach()
	if err != nil {
		fmt.Println(err)
		return
	}
	acc := mock.Account()
	nep5Id := "5594c690c3618a170a77d2696688f908efec4da2b94363fcb96749516307031d"

	fmt.Println("----pledge----")
	para := api.PledgeParam{
		Beneficial:    acc.Address(),
		PledgeAddress: qlcAccount.Address(),
		Amount:        types.Balance{Int: big.NewInt(240000000)},
		PType:         "vote",
		NEP5TxId:      nep5Id,
	}
	sendSb := new(types.StateBlock)
	err = client.Call(sendSb, "pledge_getPledgeBlock", para)
	if err != nil {
		fmt.Println(err)
		return
	}
	signAndWork(sendSb, qlcAccount)
	var hash types.Hash
	err = client.Call(&hash, "ledger_process", sendSb)
	if err != nil {
		fmt.Println(err)
		return
	}
	receSb := new(types.StateBlock)
	err = client.Call(receSb, "pledge_getPledgeRewardBlock", sendSb)
	if err != nil {
		fmt.Println(err)
		return
	}
	signAndWork(receSb, acc)
	err = client.Call(&hash, "ledger_process", receSb)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("----send----")
	sPara := api.APISendBlockPara{
		From:      gasAccount.Address(),
		TokenName: "QGAS",
		To:        acc.Address(),
		Amount:    types.Balance{Int: big.NewInt(50)},
	}
	sendBlk := new(types.StateBlock)
	err = client.Call(&sendBlk, "ledger_generateSendBlock", sPara, hex.EncodeToString(gasAccount.PrivateKey()))
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.Call(&hash, "ledger_process", sendBlk)
	if err != nil {
		fmt.Println(err)
		return
	}
	rBlk := new(types.StateBlock)
	err = client.Call(&rBlk, "ledger_generateReceiveBlock", sendBlk, hex.EncodeToString(acc.PrivateKey()))
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.Call(&hash, "ledger_process", rBlk)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("----reward----")
	param := &api.RewardsParam{
		Id:     nep5Id,
		Amount: types.Balance{Int: big.NewInt(100)},
		Self:   gasAccount.Address(),
		To:     acc.Address(),
	}

	err = client.Call(&hash, "rewards_getUnsignedRewardData", param)
	if err != nil {
		fmt.Println(err)
		return
	}
	sign := gasAccount.Sign(hash)
	rewardSendBlk := new(types.StateBlock)
	err = client.Call(&rewardSendBlk, "rewards_getSendRewardBlock", param, sign)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.Call(&hash, "ledger_process", rewardSendBlk)
	if err != nil {
		fmt.Println(err)
		return
	}
	rewardReceBlk1 := new(types.StateBlock)
	err = client.Call(&rewardReceBlk1, "rewards_getReceiveRewardBlock", rewardSendBlk.GetHash())
	if err != nil {
		fmt.Println(err)
		return
	}
	time.Sleep(1 * time.Second)
	rewardReceBlk2 := new(types.StateBlock)
	err = client.Call(&rewardReceBlk2, "rewards_getReceiveRewardBlock", rewardSendBlk.GetHash())
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.Call(&hash, "ledger_process", rewardReceBlk1)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.Call(&hash, "ledger_process", rewardReceBlk2)
	if err != nil {
		fmt.Println(err)
		//return
	}

	time.Sleep(2 * time.Second)
	err = verifier.Rollback(rewardReceBlk1.GetHash())
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.Call(&hash, "ledger_process", rewardReceBlk2)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("----done----")
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

func initData(ledger *ledger.Ledger) {
	blocks, _ := mock.BlockChain(false)
	if err := verifier.BlockProcess(blocks[0]); err != nil {
		fmt.Println(err)
		return
	}
	for i, b := range blocks[1:5] {
		fmt.Println(i + 1)
		fmt.Println(b.String())
		if r, err := verifier.Process(b); r != process.Progress || err != nil {
			fmt.Println(r.String(), err)
			return
		}
	}
	fmt.Println("account1, ", blocks[0].GetAddress().String())
	fmt.Println("account2, ", blocks[2].GetAddress().String())
	fmt.Println("account1 resp,", blocks[0].GetRepresentative())
	ac, err := ledger.GetAccountMeta(blocks[0].GetAddress())
	if err != nil {
		return
	}
	ac.CoinVote = types.Balance{Int: big.NewInt(123)}
	err = ledger.UpdateAccountMeta(ac, ledger.Cache().GetCache())
	if err != nil {
		return
	}
	//
	//fmt.Println("roll back hash: ", blocks[4].GetHash())
	//if err := ledger.Rollback(blocks[4].GetHash()); err != nil {
	//	fmt.Println("rollback err, ", err)
	//	return
	//}

	// unchecked
	if err := ledger.AddUncheckedBlock(mock.Hash(), mock.StateBlockWithoutWork(), types.UncheckedKindLink, types.UnSynchronized); err != nil {
		fmt.Println(err)
		return
	}
	if err := ledger.AddUncheckedBlock(mock.Hash(), mock.StateBlockWithoutWork(), types.UncheckedKindPrevious, types.UnSynchronized); err != nil {
		fmt.Println(err)
		return
	}
	if err := ledger.AddSmartContractBlock(mock.SmartContractBlock()); err != nil {
		fmt.Println(err)
		return
	}
	ph1 := []byte("15800001111")
	ph2 := []byte("15811110000")
	message := "hello"
	m, _ := json.Marshal(message)
	mHash, _ := types.HashBytes(m)
	blk1 := mock.StateBlockWithoutWork()
	blk1.Sender = ph1
	blk2 := mock.StateBlockWithoutWork()
	blk2.Sender = ph2
	blk2.Receiver = ph1
	blk2.Message = &mHash
	if err := ledger.AddStateBlock(blk1); err != nil {
		fmt.Println(err)
		return
	}
	if err := ledger.AddStateBlock(blk2); err != nil {
		fmt.Println(err)
		return
	}

	pk := types.PendingKey{
		Hash:    blocks[3].GetHash(),
		Address: mock.Address(),
	}
	pi := types.PendingInfo{
		Type:   blocks[3].GetToken(),
		Source: mock.Address(),
	}
	if err := ledger.AddPending(&pk, &pi, ledger.Cache().GetCache()); err != nil {
		fmt.Println(err)
		return
	}
}
