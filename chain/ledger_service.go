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

	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
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
	var mintageBlock, genesisBlock types.StateBlock
	for _, v := range ls.cfg.Genesis.GenesisBlocks {
		_ = json.Unmarshal([]byte(v.Genesis), &genesisBlock)
		_ = json.Unmarshal([]byte(v.Mintage), &mintageBlock)
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: mintageBlock,
			GenesisBlock:        genesisBlock,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
	}
	if len(common.GenesisInfos) == 0 {
		return errors.New("no genesis info")
	} else if common.ChainToken() == types.ZeroHash || common.GasToken() == types.ZeroHash {
		return errors.New("no chain token info or gas token info")
	} else {
		if c, _ := l.CountStateBlocks(); c != 0 {
			chainHash := common.GenesisBlockHash()
			gasHash := common.GasBlockHash()
			b1, _ := l.HasStateBlockConfirmed(chainHash)
			b2, _ := l.HasStateBlockConfirmed(gasHash)
			if !b1 || !b2 {
				return errors.New("chain token info or gas token info mismatch")
			}
		}
	}
	ctx := vmstore.NewVMContext(l)
	for _, v := range common.GenesisInfos {
		err := ctx.SetStorage(types.MintageAddress[:], v.GenesisBlock.Token[:], v.GenesisBlock.Data)
		if err != nil {
			ls.logger.Error(err)
		}
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
	_ = ctx.SaveStorage()
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
