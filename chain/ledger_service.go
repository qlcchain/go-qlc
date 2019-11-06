/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"encoding/json"
	"errors"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
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
	return &LedgerService{
		Ledger: ledger.NewLedger(cfgFile),
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
	ctx := vmstore.NewVMContext(l)
	var mintageBlock, genesisBlock types.StateBlock
	for _, v := range ls.cfg.Genesis.GenesisBlocks {
		_ = json.Unmarshal([]byte(v.Genesis), &genesisBlock)
		_ = json.Unmarshal([]byte(v.Mintage), &mintageBlock)
		err := ctx.SetStorage(types.MintageAddress[:], genesisBlock.Token[:], genesisBlock.Data)
		if err != nil {
			ls.logger.Error(err)
		}
		_ = ctx.SaveStorage()
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: mintageBlock,
			GenesisBlock:        genesisBlock,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
		verifier := process.NewLedgerVerifier(l)
		if b, err := l.HasStateBlock(mintageBlock.GetHash()); !b && err == nil {
			if err := l.AddStateBlock(&mintageBlock); err != nil {
				ls.logger.Error(err)
			}
		} else {
			if err != nil {
				return err
			}
		}

		if b, err := l.HasStateBlock(genesisBlock.GetHash()); !b && err == nil {
			if err := verifier.BlockProcess(&genesisBlock); err != nil {
				ls.logger.Error(err)
			}
		} else {
			if err != nil {
				return err
			}
		}
	}
	return nil
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
