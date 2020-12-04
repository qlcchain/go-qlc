/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"context"
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/trie"
	"time"

	"go.uber.org/zap"

	chainCtx "github.com/qlcchain/go-qlc/chain/context"
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
	ctx    context.Context
	cancel context.CancelFunc
}

func NewLedgerService(cfgFile string) *LedgerService {
	cc := chainCtx.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	l := ledger.NewLedger(cfgFile)
	_ = l.SetCacheCapacity()
	ctx, cancel := context.WithCancel(context.Background())
	return &LedgerService{
		Ledger: l,
		logger: log.NewLogger("ledger_service"),
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
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

	if ls.cfg.TrieClean.Enable {
		duration := time.Duration(ls.cfg.TrieClean.PeriodDay*24) * time.Hour
		go ls.clean(duration, ls.cfg.TrieClean.HeightInterval)
	}
	return nil
}

func (ls *LedgerService) Stop() error {
	if !ls.PreStop() {
		return errors.New("pre stop fail")
	}
	defer ls.PostStop()
	ls.cancel()
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

func (ls *LedgerService) clean(duration time.Duration, minHeightInterval uint64) {
	cTicker := time.NewTicker(duration)
	for {
		select {
		case <-cTicker.C:
			lastCleanHeight, err := ls.Ledger.GetTrieCleanHeight()
			if err != nil {
				break
			}
			bestPovHeight, err := ls.Ledger.GetPovLatestHeight()
			if err != nil {
				break
			}
			if bestPovHeight-lastCleanHeight > minHeightInterval {
				if err := ls.cleanTrie(lastCleanHeight, bestPovHeight-minHeightInterval); err != nil {
					ls.logger.Error(err)
					break
				}
			}
			if err := ls.Ledger.AddTrieCleanHeight(bestPovHeight); err != nil {
				break
			}
		case <-ls.ctx.Done():
			return
		}
	}
}

func (ls *LedgerService) cleanTrie(startHeight, endHeight uint64) error {
	l := ls.Ledger
	if err := l.GetAccountMetas(func(am *types.AccountMeta) error {
		for _, token := range am.Tokens {
			startHash := token.Header
			endLoop := false
			for endLoop {
				blk, err := l.GetStateBlockConfirmed(startHash)
				if err != nil {
					ls.logger.Errorf("%s: %s", err, startHash.String())
					endLoop = true
				}
				if blk.IsContractBlock() && !config.IsGenesisBlock(blk) &&
					blk.PoVHeight >= startHeight && blk.PoVHeight < endHeight {
					extra := blk.GetExtra()
					if !extra.IsZero() {
						t := trie.NewTrie(ls.Ledger.DBStore(), &extra, trie.NewSimpleTrieNodePool())
						_ = t.Remove()
					}
				}
				startHash = blk.GetPrevious()
				if startHash == types.ZeroHash {
					endLoop = true
				}
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("clean trie: %s", err)
	}
	_, _ = ls.Ledger.Action(storage.GC, 0)
	return nil
}
