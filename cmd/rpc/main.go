/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/mock"
)

var logger = log.NewLogger("main")

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	dir := filepath.Join(config.QlcTestDataDir(), "cmd")
	cm := config.NewCfgManager(dir)
	cfg, err := cm.Load()
	cfg.RPC.Enable = true
	if cfg.RPC.Enable == false {
		return
	}

	ss, err := services.NewSqliteService(cfg)
	if err != nil {
		logger.Fatal(err)
		return
	}
	if err = ss.Init(); err != nil {
		logger.Fatal(err)
	}
	if err = ss.Start(); err != nil {
		logger.Fatal(err)
	}
	logger.Info("sqlite started")

	l := services.NewLedgerService(cfg)
	if err := l.Init(); err != nil {
		return
	}
	if err := l.Start(); err != nil {
		logger.Fatal(err)
	}
	logger.Info("ledger started")
	initData(l.Ledger)

	w := services.NewWalletService(cfg)
	if err := w.Init(); err != nil {
		return
	}
	if err := w.Start(); err != nil {
		logger.Fatal(err)
	}
	logger.Info("wallet started")

	rs, err := services.NewRPCService(cfg)
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

	defer func() {
		l.Stop()
		w.Stop()
		rs.Stop()
		ss.Stop()
		os.RemoveAll(dir)
	}()
	s := <-c
	fmt.Println("Got signal: ", s)
}

func initData(ledger *ledger.Ledger) {
	verifier := process.NewLedgerVerifier(ledger)

	blocks, _ := mock.BlockChain()

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
	ac.CoinVote = types.Balance{big.NewInt(123)}
	err = ledger.UpdateAccountMeta(ac)
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
	blk2.Message = mHash
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
	if err := ledger.AddPending(&pk, &pi); err != nil {
		fmt.Println(err)
		return
	}
}
