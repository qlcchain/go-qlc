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
	"time"

	"go.uber.org/zap"

	chainCtx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
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
	if err := l.SetStorage(vmstore.ToCache(ctx)); err != nil {
		return err
	}

	if ls.cfg.TrieClean.Enable {
		if _, err := ls.Ledger.GetTrieCleanHeight(); err != nil {
			if err := ls.removeUselessTrie(); err != nil {
				ls.logger.Error(err)
			} else {
				_ = ls.Ledger.AddTrieCleanHeight(ls.cfg.TrieClean.SyncWriteHeight)
			}
		}
		duration := time.Duration(ls.cfg.TrieClean.PeriodDay*24) * time.Hour
		go ls.clean(duration, ls.cfg.TrieClean.HeightInterval)
	}
	return nil
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
			lastCleanHeight, _ := ls.Ledger.GetTrieCleanHeight()
			bestPovHeight, err := ls.Ledger.GetPovLatestHeight()
			if err != nil {
				break
			}
			if bestPovHeight-lastCleanHeight >= minHeightInterval {
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
	hashes := make([]*types.Hash, 0)
	if err := l.GetAccountMetas(func(am *types.AccountMeta) error {
		for _, token := range am.Tokens {
			blockHash := token.Header
			for {
				block, err := l.GetStateBlock(blockHash)
				if err != nil {
					return fmt.Errorf("%s: %s", err, blockHash.String())
				}
				if block.IsContractBlock() && !config.IsGenesisBlock(block) &&
					block.PoVHeight >= startHeight && block.PoVHeight < endHeight {
					extra := block.GetExtra()
					if !extra.IsZero() {
						t := trie.NewTrie(ls.Ledger.DBStore(), &extra, trie.NewSimpleTrieNodePool())
						//_ = t.Remove()
						nodes := t.NewNodeIterator(func(node *trie.TrieNode) bool {
							return true
						})
						for node := range nodes {
							if node != nil {
								nodeHash := node.Hash()
								hashes = append(hashes, nodeHash)
								if node.NodeType() == trie.HashNode {
									refKey := node.Value()
									refHash, err := types.BytesToHash(refKey)
									if err == nil {
										hashes = append(hashes, &refHash)
									}
								}
							}
						}
					}
				}
				blockHash = block.GetPrevious()
				if blockHash == types.ZeroHash {
					break
				}
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("clean trie: %s", err)
	}

	batch := l.DBStore().Batch(false)
	for _, hash := range hashes {
		if err := batch.Delete(encodeTrieKey(hash.Bytes(), true)); err != nil {
			return fmt.Errorf("batch put: %s", err)
		}
	}
	if err := l.DBStore().PutBatch(batch); err != nil {
		return fmt.Errorf("put batch: %s", err)
	}
	_, _ = ls.Ledger.Action(storage.GC, 0)
	return nil
}

func (ls *LedgerService) removeUselessTrie() error {
	if count, err := ls.Ledger.BlocksCount(); err != nil || int(count) <= len(config.AllGenesisBlocks()) {
		return nil
	}
	if _, err := ls.Ledger.GetLatestPovBlock(); err != nil {
		return nil
	}
	if err := ls.backupAccountTrieData(); err != nil {
		return fmt.Errorf("account trie: %s", err)
	}
	if err := ls.backupPovTrieData(); err != nil {
		return fmt.Errorf("pov trie: %s", err)
	}
	if err := ls.resetTrie(); err != nil {
		return fmt.Errorf("reset trie: %s", err)
	}
	_, _ = ls.Ledger.Action(storage.GC, 0)
	return nil
}

func (ls *LedgerService) backupPovTrieData() error {
	l := ls.Ledger
	hashes := make([]*types.Hash, 0)

	latestBlk, err := ls.Ledger.GetLatestPovBlock()
	if err != nil {
		return err
	}
	latestTrieRoot := latestBlk.GetStateHash()

	globalHashes, err := trieHashesByRoot(l.DBStore(), latestTrieRoot)
	if err != nil {
		return fmt.Errorf("pov global trie: %s", err)
	}
	hashes = append(hashes, globalHashes...)

	statedb := statedb.NewPovGlobalStateDB(l.DBStore(), latestTrieRoot)
	for _, addr := range contractaddress.PovContractStateAddressList {
		contractState, err := statedb.LookupContractStateDB(addr)
		if err != nil {
			return err
		} else {
			curTrie := contractState.GetCurTrie()
			if curTrie.Root != nil {
				root := *curTrie.Root.Hash()
				contractHashes, err := trieHashesByRoot(l.DBStore(), root)
				if err != nil {
					return fmt.Errorf("pov global trie: %s", err)
				}
				hashes = append(hashes, contractHashes...)
			}
		}
	}

	b := l.DBStore().Batch(false)
	for _, h := range hashes {
		v, err := l.DBStore().Get(encodeTrieKey(h.Bytes(), true))
		if err == nil {
			if err := b.Put(encodeTrieKey(h.Bytes(), false), v); err != nil {
				return fmt.Errorf("put: %s", err)
			}
		} else {
			ls.logger.Error(err)
		}
	}
	return l.DBStore().PutBatch(b)
}

func (ls *LedgerService) backupAccountTrieData() error {
	l := ls.Ledger
	hashes := make([]*types.Hash, 0)

	if err := ls.Ledger.GetAccountMetas(func(am *types.AccountMeta) error {
		for _, token := range am.Tokens {
			blkHash := token.Header
			for {
				blk, err := l.GetStateBlock(blkHash)
				if err != nil {
					return fmt.Errorf("get block: %s", err)
				}
				if blk.PoVHeight < ls.cfg.TrieClean.SyncWriteHeight {
					break
				}
				if blk.IsContractBlock() && !config.IsGenesisBlock(blk) {
					rootHash := blk.GetExtra()
					if !rootHash.IsZero() {
						hs, err := trieHashesByRoot(l.DBStore(), rootHash)
						if err != nil {
							return fmt.Errorf("pov account trie: %s, address: %s", err, am.Address.String())
						}
						hashes = append(hashes, hs...)
					}
				}
				blkHash = blk.GetPrevious()
				if blkHash == types.ZeroHash {
					break
				}
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("clean trie: %s", err)
	}

	batch := l.DBStore().Batch(false)
	for _, hash := range hashes {
		value, err := l.DBStore().Get(encodeTrieKey(hash.Bytes(), true))
		if err == nil {
			if err := batch.Put(encodeTrieKey(hash.Bytes(), false), value); err != nil {
				return fmt.Errorf("batch put: %s", err)
			}
		} else {
			ls.logger.Error(err)
		}
	}
	return l.DBStore().PutBatch(batch)
}

func trieHashesByRoot(db storage.Store, rootHash types.Hash) ([]*types.Hash, error) {
	hashes := make([]*types.Hash, 0)

	t := trie.NewTrie(db, &rootHash, trie.GetGlobalTriePool())
	nodes := t.NewNodeIterator(func(node *trie.TrieNode) bool {
		return true
	})
	for node := range nodes {
		if node != nil {
			nodeHash := node.Hash()
			hashes = append(hashes, nodeHash)
			if node.NodeType() == trie.HashNode {
				refKey := node.Value()
				refHash, err := types.BytesToHash(refKey)
				if err != nil {
					return nil, err
				}
				hashes = append(hashes, &refHash)
			}
		}
	}
	return hashes, nil
}

func (ls *LedgerService) resetTrie() error {
	l := ls.Ledger
	var hashes []types.Hash
	err := l.DBStore().Iterator([]byte{byte(storage.KeyPrefixTrie)}, nil, func(k, v []byte) error {
		if _, err := l.Get(encodeTrieKey(k[1:], false)); err != nil {
			if h, err := types.BytesToHash(k[1:]); err == nil {
				hashes = append(hashes, h)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("reset key: %s", err)
	}
	if len(hashes) > 0 {
		batch := l.DBStore().Batch(false)
		for _, hash := range hashes {
			if err := batch.Delete(encodeTrieKey(hash.Bytes(), true)); err != nil {
				return fmt.Errorf("batch delete: %s", err)
			}
		}
		if err := l.DBStore().PutBatch(batch); err != nil {
			return fmt.Errorf("put batch: %s", err)
		}
	}

	_ = l.DBStore().Drop([]byte{byte(storage.KeyPrefixTrieTemp)})
	_, _ = l.Action(storage.GC, 0)
	return nil
}

func encodeTrieKey(key []byte, original bool) []byte {
	result := make([]byte, len(key)+1)
	if original {
		result[0] = byte(storage.KeyPrefixTrie)
	} else {
		result[0] = byte(storage.KeyPrefixTrieTemp)
	}
	copy(result[1:], key)
	return result
}
