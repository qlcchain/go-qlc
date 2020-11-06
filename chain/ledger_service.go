/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type LedgerService struct {
	common.ServiceLifecycle
	Ledger *ledger.Ledger
	logger *zap.SugaredLogger
	cfg    *config.Config
}

func NewLedgerService(cfgFile string) *LedgerService {
	cc := context.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	l := ledger.NewLedger(cfgFile)
	_ = l.SetCacheCapacity()
	return &LedgerService{
		Ledger: l,
		logger: log.NewLogger("ledger_service"),
		cfg:    cfg,
	}
}

func (ls *LedgerService) Init() error {
	if !ls.PreInit() {
		return errors.New("pre init fail")
	}
	defer ls.PostInit()
	l := ls.Ledger

	genesisInfos := config.GenesisInfos()

	if len(genesisInfos) == 0 {
		return errors.New("no genesis info")
	} else if config.ChainToken() == types.ZeroHash || config.GasToken() == types.ZeroHash {
		return errors.New("no chain token info or gas token info")
	} else {
		if c, _ := l.CountStateBlocks(); c != 0 {
			chainHash := config.GenesisBlockHash()
			gasHash := config.GasBlockHash()
			b1, _ := l.HasStateBlockConfirmed(chainHash)
			b2, _ := l.HasStateBlockConfirmed(gasHash)
			if !b1 || !b2 {
				return errors.New("chain token info or gas token info mismatch")
			}
		}
	}
	ctx := vmstore.NewVMContext(l, &contractaddress.MintageAddress)
	for _, v := range genesisInfos {
		mb := v.Mintage
		gb := v.Genesis
		err := ctx.SetStorage(contractaddress.MintageAddress[:], gb.Token[:], gb.Data)
		if err != nil {
			ls.logger.Error(err)
		}
		verifier := process.NewLedgerVerifier(l)
		if b, err := l.HasStateBlock(mb.GetHash()); !b && err == nil {
			if err := l.AddStateBlock(&mb); err != nil {
				ls.logger.Error(err)
			}
		} else {
			if err != nil {
				return err
			}
		}

		if b, err := l.HasStateBlock(gb.GetHash()); !b && err == nil {
			if err := verifier.BlockProcess(&gb); err != nil {
				ls.logger.Error(err)
			}
		} else {
			if err != nil {
				return err
			}
		}
	}
	return l.SetStorage(vmstore.ToCache(ctx))
}

func (ls *LedgerService) Start() error {
	if !ls.PreStart() {
		return errors.New("pre start fail")
	}
	defer ls.PostStart()
	if err := ls.registerRelation(); err != nil {
		return fmt.Errorf("ledger start: %s", err)
	}
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

func (ls *LedgerService) registerRelation() error {
	//if err := ls.Ledger.RegisterInterface(new(StructA),new(StructB));err !=nil{
	//	return err
	//}
	return nil
}
