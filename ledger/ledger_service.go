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
	"github.com/qlcchain/go-qlc/test/mock"
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
	// TODO: remove
	l := ls.Ledger
	return l.BatchUpdate(func(txn db.StoreTxn) error {
		//insert smart contract block
		for _, sb := range mock.GetSmartContracts() {
			h := sb.GetHash()
			if b, err := l.HasBlock(h, txn); !b && err != nil {
				err := l.AddBlock(sb, txn)
				if err != nil {
					logger.Error(err)
					return nil
				}
			}
		}
		// insert genesis blocks
		for _, b := range mock.GetGenesis() {
			if l.BlockCheck(b) == Progress {
				err := l.BlockProcess(b)
				if err != nil {
					logger.Error(err)
					return err
				}
			}
		}
		//l.BlockProcess()
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
