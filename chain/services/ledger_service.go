/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package services

import (
	"errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type LedgerService struct {
	common.ServiceLifecycle
	Ledger *ledger.Ledger
	eb     event.EventBus
	logger *zap.SugaredLogger
}

func NewLedgerService(cfg *config.Config, eb event.EventBus) *LedgerService {
	return &LedgerService{
		Ledger: ledger.NewLedger(cfg.LedgerDir(), eb),
		logger: log.NewLogger("ledger_service"),
	}
}

func (ls *LedgerService) Init() error {
	if !ls.PreInit() {
		return errors.New("pre init fail")
	}
	defer ls.PostInit()
	l := ls.Ledger

	genesis := common.GenesisBlock()
	ctx := vmstore.NewVMContext(l)
	err := ctx.SetStorage(types.MintageAddress[:], genesis.Token[:], genesis.Data)
	if err != nil {
		ls.logger.Error(err)
	}
	verifier := process.NewLedgerVerifier(l)
	mintageHash := common.GenesisMintageHash()
	if b, err := l.HasStateBlock(mintageHash); !b && err == nil {
		mintage := common.GenesisMintageBlock()
		if err := l.AddStateBlock(&mintage); err != nil {
			ls.logger.Error(err)
		}
	} else {
		if err != nil {
			return err
		}
	}

	genesisHash := common.GenesisBlockHash()
	if b, err := l.HasStateBlock(genesisHash); !b && err == nil {
		if err := verifier.BlockProcess(&genesis); err != nil {
			ls.logger.Error(err)
		}
	} else {
		if err != nil {
			return err
		}
	}

	//gas block storage
	gas := common.GasBlock()
	err = ctx.SetStorage(types.MintageAddress[:], gas.Token[:], gas.Data)
	if err != nil {
		ls.logger.Error(err)
	}

	err = ctx.SaveStorage()
	if err != nil {
		ls.logger.Error(err)
	}

	gasMintageHash := common.GasMintageHash()
	if b, err := l.HasStateBlock(gasMintageHash); !b && err == nil {
		gasMintage := common.GasMintageBlock()
		if err := l.AddStateBlock(&gasMintage); err != nil {
			ls.logger.Error(err)
		}
	} else {
		if err != nil {
			return err
		}
	}

	gasHash := common.GasBlockHash()
	if b, err := l.HasStateBlock(gasHash); !b && err == nil {
		if err := verifier.BlockProcess(&gas); err != nil {
			ls.logger.Error(err)
		}
	} else {
		if err != nil {
			return err
		}
	}

	return nil

	//return l.BatchUpdate(func(txn db.StoreTxn) error {
	//	genesis := common.QLCGenesisBlock
	//	_ = l.SetStorage(types.MintageAddress[:], genesis.Token[:], genesis.Data)
	//	verifier := process.NewLedgerVerifier(l)
	//	if b, err := l.HasStateBlock(common.GenesisMintageHash, txn); !b && err == nil {
	//		if err := l.AddStateBlock(&common.GenesisMintageBlock, txn); err != nil {
	//			ls.logger.Error(err)
	//		}
	//	}
	//
	//	if b, err := l.HasStateBlock(common.QLCGenesisBlockHash, txn); !b && err == nil {
	//		if err := verifier.BlockProcess(&common.QLCGenesisBlock); err != nil {
	//			ls.logger.Error(err)
	//		}
	//	}
	//	return nil
	//})
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
	ledger.CloseLedger()

	return nil
}

func (ls *LedgerService) Status() int32 {
	return ls.State()
}
