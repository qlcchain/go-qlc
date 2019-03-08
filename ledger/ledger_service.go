/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package ledger

import (
	"errors"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/db"
)

type LedgerService struct {
	common.ServiceLifecycle
	Ledger *Ledger
}

func NewLedgerService(cfg *config.Config) *LedgerService {
	return &LedgerService{
		Ledger: NewLedger(cfg.LedgerDir()),
	}
}

func (ls *LedgerService) Init() error {
	if !ls.PreInit() {
		return errors.New("pre init fail")
	}
	defer ls.PostInit()
	l := ls.Ledger
	return l.BatchUpdate(func(txn db.StoreTxn) error {
		//TODO: init genesis block
		genesis := common.QLCGenesisBlock
		_ = l.SetStorage(genesis.Token[:], genesis.Data)
		return nil
	})
}

func (ls *LedgerService) Start() error {
	if !ls.PreStart() {
		return errors.New("pre start fail")
	}
	defer ls.PostStart()

	return nil
}

func (ls *LedgerService) Stop() error {
	if !ls.PreStop() {
		return errors.New("pre stop fail")
	}
	defer ls.PostStop()

	ls.Ledger.Close()
	// close all ledger
	CloseLedger()

	return nil
}

func (ls *LedgerService) Status() int32 {
	return ls.State()
}
