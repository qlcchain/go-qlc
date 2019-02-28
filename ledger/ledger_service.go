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
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
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
		sbs := mock.GetSmartContracts()
		for i := 0; i < len(sbs); i++ {
			sb := sbs[i]
			h := sb.GetHash()
			if b, err := l.HasSmartContrantBlock(h, txn); !b && err == nil {
				err := l.AddSmartContrantBlock(sb, txn)
				if err != nil {
					ls.Ledger.logger.Error(err)
					return nil
				}
				ls.Ledger.logger.Debugf("save sb[%s] successful", h.String())
			}
		}
		// insert genesis blocks
		genesis := mock.GetGenesis()
		for i := 0; i < len(genesis); i++ {
			b := genesis[i].(*types.StateBlock)
			hash := b.GetHash()
			if exist, err := l.HasStateBlock(hash); !exist && err == nil {
				err := l.BlockProcess(b)
				if err != nil {
					ls.Ledger.logger.Error(err)
				} else {
					ls.Ledger.logger.Debugf("save block[%s]", hash.String())
				}
				if err != nil {
					ls.Ledger.logger.Error(err)
				}
			} else {
				ls.Ledger.logger.Debugf("%s, %v", hash.String(), err)
				meta, _ := l.GetAccountMeta(b.Address)
				ls.Ledger.logger.Debug(util.ToString(&meta))
			}
		}

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
