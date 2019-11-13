package process

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/yireyun/go-queue"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func (lv *LedgerVerifier) Rollback(hash types.Hash) error {
	if b, err := lv.l.HasBlockCache(hash); b && err == nil {
		return lv.RollbackCache(hash)
	}

	if b, err := lv.l.HasStateBlockConfirmed(hash); !b || err != nil {
		lv.logger.Warnf("rollback block not found: %s", hash.String())
		return nil
	}
	lv.logger.Errorf("process rollback  block: %s", hash.String())
	// get blocks to roll back
	rollbackMap := make(map[types.Address]*types.StateBlock)
	relatedBlocks := queue.NewQueue(512)
	relatedBlocks.Put(hash)
	lv.logger.Debugf("put block to queue %s ", hash.String())

	for {
		if v, ok, _ := relatedBlocks.Get(); ok {
			// get oldest block
			oldestHash := v.(types.Hash)
			oldestBlock, err := lv.l.GetStateBlockConfirmed(oldestHash)
			if err != nil {
				return fmt.Errorf("can not get block  %s", oldestHash.String())
			}
			lv.logger.Debugf("get block from  queue %s (%s) ,%s  ", oldestBlock.GetHash().String(), oldestBlock.GetType().String(), oldestBlock.Address.String())

			// put oldest block to rollbackMap
			if rBlock, ok := rollbackMap[oldestBlock.GetAddress()]; ok {
				lv.logger.Debugf("get %s from rollback of %s ", rBlock.GetHash().String(), oldestBlock.GetAddress().String())
				if t, err := lv.blockOrderCompare(oldestBlock, rBlock); t && err == nil {
					lv.logger.Debugf("put block to rollback %s (%s), %s ", oldestBlock.GetHash().String(), oldestBlock.GetType().String(), oldestBlock.Address.String())
					rollbackMap[oldestBlock.GetAddress()] = oldestBlock
				} else if err != nil {
					return err
				}
			} else {
				lv.logger.Debugf("put block  to rollback %s (%s), %s ", oldestBlock.GetHash().String(), oldestBlock.GetType().String(), oldestBlock.Address.String())
				rollbackMap[oldestBlock.GetAddress()] = oldestBlock
			}

			// get header block
			tm, err := lv.l.GetTokenMetaConfirmed(oldestBlock.GetAddress(), oldestBlock.GetToken())
			if err != nil {
				return fmt.Errorf("can not get account of block %s", oldestHash.String())
			}
			headerHash := tm.Header
			headerBlock, err := lv.l.GetStateBlockConfirmed(headerHash)
			if err != nil {
				return fmt.Errorf("can not get header block %s", headerHash.String())
			}

			curBlock := headerBlock
			// put link block to rollbackMap
			for {
				//if curBlock.IsOpen() {
				//	break
				//}

				if curBlock.IsSendBlock() {
					linkHash, err := lv.l.GetLinkBlock(curBlock.GetHash())
					// link not found is not error ,may be send block has created but receiver block has not created
					if err != nil && err != ledger.ErrLinkNotFound {
						return fmt.Errorf("can not get link hash %s", curBlock.GetHash().String())
					}
					if !linkHash.IsZero() {
						linkBlock, err := lv.l.GetStateBlockConfirmed(linkHash)
						if err != nil {
							return fmt.Errorf("can not get link block %s", linkHash.String())
						}
						if rBlock, ok := rollbackMap[linkBlock.GetAddress()]; ok {
							lv.logger.Debugf("get link %s from rollback of %s ", rBlock.GetHash().String(), linkBlock.GetAddress().String())
							if t, err := lv.blockOrderCompare(linkBlock, rBlock); t && err == nil {
								lv.logger.Debugf("put block to queue %s (%s) ,%s ", linkBlock.GetHash().String(), linkBlock.GetType().String(), linkBlock.Address.String())
								relatedBlocks.Put(linkBlock.GetHash())
							} else if err != nil {
								return err
							}
						} else {
							lv.logger.Debugf("put block to queue %s (%s), %s ", linkBlock.GetHash().String(), linkBlock.GetType().String(), linkBlock.Address.String())
							relatedBlocks.Put(linkBlock.GetHash())
						}
					}
				}

				if curBlock.GetHash() == oldestHash {
					break
				}

				curHash := curBlock.GetPrevious()
				curBlock, err = lv.l.GetStateBlockConfirmed(curHash)
				if err != nil {
					return fmt.Errorf("can not get previous block %s", curHash.String())
				}
			}
		} else {
			break
		}
	}
	// roll back blocks in badger
	return lv.l.BatchUpdate(func(txn db.StoreTxn) error {
		err := lv.rollbackBlocks(rollbackMap, txn)
		if err != nil {
			lv.logger.Error(err)
		}
		return err
	})
}

func (lv *LedgerVerifier) RollbackCache(hash types.Hash) error {
	if b, err := lv.l.HasBlockCache(hash); b && err == nil {
		lv.logger.Errorf("process rollback cache block: %s", hash.String())
		return lv.l.BatchUpdate(func(txn db.StoreTxn) error {
			err = lv.rollbackCache(hash, txn)
			if err != nil {
				lv.logger.Error(err)
				return err
			}
			return nil
		})
	}
	return nil
}

// rollback cache blocks
func (lv *LedgerVerifier) rollbackCache(hash types.Hash, txn db.StoreTxn) error {
	block, err := lv.l.GetBlockCache(hash, txn)
	if err != nil {
		return fmt.Errorf("get cache block (%s) err: %s", hash.String(), err)
	}

	// get all blocks of the address
	blocks := make([]*types.StateBlock, 0)
	err = lv.l.GetBlockCaches(func(b *types.StateBlock) error {
		if block.GetAddress() == b.GetAddress() && block.GetToken() == b.GetToken() {
			blocks = append(blocks, b)
		}
		return nil
	})
	if err != nil {
		lv.logger.Error("get block cache error")
	}

	// if receive repeated , rollback later block
	rollBlock := block
	if rollBlock.IsReceiveBlock() {
		for _, b := range blocks {
			if rollBlock.GetLink() == b.GetLink() && rollBlock.GetHash() != b.GetHash() {
				if b.Timestamp > rollBlock.Timestamp {
					rollBlock = b
				}
				break
			}
		}
	}

	// get rollback blocks
	tempBlocks := make([]*types.StateBlock, 0)
	for _, b := range blocks {
		if b.Timestamp >= rollBlock.Timestamp {
			tempBlocks = append(tempBlocks, b)
		}
	}

	// sort
	sort.Slice(tempBlocks, func(i, j int) bool {
		return tempBlocks[i].Timestamp < tempBlocks[j].Timestamp
	})

	rollBlocks := make([]*types.StateBlock, 0)
	rollBlocks = append(rollBlocks, rollBlock)
	for {
		findPre := false
		for _, b := range tempBlocks {
			if b.GetPrevious() == rollBlock.GetHash() {
				rollBlocks = append(rollBlocks, b)
				rollBlock = b
				findPre = true
				break
			}
		}
		if !findPre {
			break
		}
	}

	// delete blocks
	if err := lv.rollbackCacheBlocks(rollBlocks, true, txn); err != nil {
		lv.logger.Error(err)
		return err
	}
	return nil
}

func (lv *LedgerVerifier) rollbackCacheBlocks(blocks []*types.StateBlock, cache bool, txn db.StoreTxn) error {
	if len(blocks) == 0 {
		return nil
	}
	if cache {
		for i := len(blocks) - 1; i >= 0; i-- {
			block := blocks[i]
			if block.IsReceiveBlock() {
				pendingKey := types.PendingKey{
					Address: block.GetAddress(),
					Hash:    block.GetLink(),
				}
				if err := lv.l.UpdatePending(&pendingKey, types.PendingNotUsed, txn); err != nil {
					return err
				}
			}

			if err := lv.l.DeleteBlockCache(block.GetHash(), txn); err != nil {
				return fmt.Errorf("delete BlockCache fail(%s), hash(%s)", err, block.GetHash().String())
			}
			lv.l.EB.Publish(common.EventRollback, block.GetHash())
			lv.logger.Errorf("rollback delete cache block %s (previous: %s, type: %s,  address: %s)", block.GetHash().String(), block.GetPrevious().String(), block.GetType(), block.GetAddress().String())

			if b, _ := lv.l.HasBlockCache(block.GetPrevious()); b {
				if err := lv.rollbackCacheAccount(block, txn); err != nil {
					lv.logger.Errorf("roll back cache account error : %s", err)
					return err
				}
			} else {
				if err := lv.rollbackCacheAccountDel(block.GetAddress(), block.GetToken(), txn); err != nil {
					lv.logger.Errorf("roll back cache account del error : %s", err)
					return err
				}
			}
		}
		return nil
	}
	for _, block := range blocks {
		if block.IsReceiveBlock() {
			pendingKey := types.PendingKey{
				Address: block.GetAddress(),
				Hash:    block.GetLink(),
			}
			if err := lv.l.UpdatePending(&pendingKey, types.PendingNotUsed, txn); err != nil {
				return err
			}
		}

		if err := lv.l.DeleteBlockCache(block.GetHash(), txn); err != nil {
			return fmt.Errorf("delete BlockCache fail(%s), hash(%s)", err, block.GetHash().String())
		}
		lv.l.EB.Publish(common.EventRollback, block.GetHash())
		lv.logger.Errorf("rollback delete cache block %s (previous: %s, type: %s,  address: %s)", block.GetHash().String(), block.GetPrevious().String(), block.GetType(), block.GetAddress().String())
	}

	blk := blocks[0]
	if err := lv.rollbackCacheAccountDel(blk.GetAddress(), blk.GetToken(), txn); err != nil {
		lv.logger.Errorf("roll back cache account error : %s", err)
		return err
	}
	return nil
}

func (lv *LedgerVerifier) rollbackCacheAccount(block *types.StateBlock, txn db.StoreTxn) error {
	ac, err := lv.l.GetAccountMetaCache(block.GetAddress(), txn)
	if err == nil {
		preBlk, err := lv.l.GetBlockCache(block.GetPrevious())
		if err == nil {
			if preBlk.GetToken() == common.ChainToken() {
				ac.CoinVote = preBlk.GetVote()
				ac.CoinOracle = preBlk.GetOracle()
				ac.CoinNetwork = preBlk.GetNetwork()
				ac.CoinStorage = preBlk.GetStorage()
				ac.CoinBalance = preBlk.GetBalance()
			}
			tm := ac.Token(block.GetToken())
			if tm != nil {
				tm.Balance = preBlk.GetBalance()
				tm.Header = preBlk.GetHash()
				tm.Representative = preBlk.GetRepresentative()
				tm.BlockCount = tm.BlockCount - 1
				tm.Modified = common.TimeNow().Unix()
				lv.logger.Debug("update token, ", tm)
				for index, t := range ac.Tokens {
					if t.Type == tm.Type {
						ac.Tokens[index] = tm
						if err := lv.l.UpdateAccountMetaCache(ac, txn); err != nil {
							return err
						}
						lv.logger.Warnf("rollback update account cache, %s", ac.String())
						return nil
					}
				}
			}
		}
	}
	return nil
}

func (lv *LedgerVerifier) rollbackCacheAccountDel(address types.Address, token types.Hash, txn db.StoreTxn) error {
	ac, err := lv.l.GetAccountMetaCache(address, txn)
	if err != nil {
		if err == ledger.ErrAccountNotFound {
			return nil
		}
		return err
	}

	if tm := ac.Token(token); tm == nil {
		return nil
	} else {
		if len(ac.Tokens) == 1 {
			if err := lv.l.DeleteAccountMetaCache(address, txn); err != nil {
				return err
			}
			lv.logger.Errorf("rollback delete account cache, %s", address.String())
			return nil
		} else {
			if err := lv.l.DeleteTokenMetaCache(address, token, txn); err != nil {
				return err
			}
			lv.logger.Errorf("rollback delete token cache, %s, %s", address, token)
			return nil
		}
	}
}

// rollback confirmed blocks
func (lv *LedgerVerifier) rollbackBlocks(rollbackMap map[types.Address]*types.StateBlock, txn db.StoreTxn) error {
	sendBlocks, err := lv.sendBlocksInRollback(rollbackMap, txn)
	if err != nil {
		return err
	}

	for address, oldestBlock := range rollbackMap {
		tm, err := lv.l.GetTokenMetaConfirmed(address, oldestBlock.GetToken())
		if err != nil {
			return fmt.Errorf("get tokenmeta error: %s (%s)", err, oldestBlock.GetHash().String())
		}
		headerHash := tm.Header

		hashCur := headerHash
		blockCur, err := lv.l.GetStateBlockConfirmed(hashCur)
		if err != nil {
			return fmt.Errorf("get block error: %s (%s)", err, hashCur.String())
		}

		for {
			blockType := blockCur.GetType()
			blockPre := new(types.StateBlock)
			if !blockCur.IsOpen() {
				blockPre, err = lv.l.GetStateBlockConfirmed(blockCur.Previous, txn)
				if err != nil {
					return fmt.Errorf("get previous block %s : %s", blockCur.Previous.String(), err)
				}
			}

			switch blockType {
			case types.Open:
				if err := lv.rollBackTokenDel(tm, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), open(%s)", err, hashCur)
				}
				if err := lv.rollBackRep(blockCur.GetRepresentative(), blockCur, nil, false, blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), open(%s)", err, hashCur)
				}
				if err := lv.rollBackFrontier(types.Hash{}, blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), open(%s)", err, hashCur)
				}
				if _, ok := sendBlocks[blockCur.GetLink()]; !ok {
					if err := lv.rollBackPendingAdd(blockCur, tm.Balance, blockCur.GetToken(), txn); err != nil {
						return fmt.Errorf("rollback pending fail(%s), open(%s)", err, hashCur)
					}
				}
			case types.Send:
				if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), send(%s)", err, hashCur)
				}
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), send(%s)", err, hashCur)
				}
				if err := lv.rollBackRep(blockCur.GetRepresentative(), blockCur, blockPre, true, blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), send(%s)", err, hashCur)
				}
				if err := lv.rollBackPendingDel(blockCur, txn); err != nil {
					return fmt.Errorf("rollback pending fail(%s), send(%s)", err, hashCur)
				}
			case types.Receive:
				if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), receive(%s)", err, hashCur)
				}
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), receive(%s)", err, hashCur)
				}
				if err := lv.rollBackRep(blockCur.GetRepresentative(), blockCur, blockPre, false, blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), receive(%s)", err, hashCur)
				}
				if _, ok := sendBlocks[blockCur.GetLink()]; !ok {
					if err := lv.rollBackPendingAdd(blockCur, blockCur.GetBalance().Sub(blockPre.GetBalance()), blockCur.GetToken(), txn); err != nil {
						return fmt.Errorf("rollback pending fail(%s), receive(%s)", err, hashCur)
					}
				}
			case types.Change, types.Online:
				if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), change(%s)", err, hashCur)
				}
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), change(%s)", err, hashCur)
				}
				if err := lv.rollBackRepChange(blockPre.GetRepresentative(), blockCur.GetRepresentative(), blockCur, txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), change(%s)", err, hashCur)
				}
			case types.ContractReward:
				previousHash := blockCur.GetPrevious()
				if previousHash.IsZero() {
					if err := lv.rollBackTokenDel(tm, txn); err != nil {
						return fmt.Errorf("rollback token fail(%s), ContractReward(%s)", err, hashCur)
					}
				} else {
					if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
						return fmt.Errorf("rollback token fail(%s), ContractReward(%s)", err, hashCur)
					}
				}
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), ContractReward(%s)", err, hashCur)
				}
				if _, ok := sendBlocks[blockCur.GetLink()]; !ok {
					if err := lv.rollBackPendingAdd(blockCur, types.ZeroBalance, types.ZeroHash, txn); err != nil {
						return fmt.Errorf("rollback pending fail(%s), ContractReward(%s)", err, hashCur)
					}
				}
				if err := lv.rollBackContractData(blockCur, txn); err != nil {
					return fmt.Errorf("rollback contract data fail(%s), ContractReward(%s)", err, blockCur.String())
				}
			case types.ContractSend:
				if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), ContractSend(%s)", err, hashCur)
				}
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), ContractSend(%s)", err, hashCur)
				}
				if err := lv.rollBackPendingDel(blockCur, txn); err != nil {
					return fmt.Errorf("rollback pending fail(%s), ContractSend(%s)", err, hashCur)
				}
				if err := lv.rollBackContractData(blockCur, txn); err != nil {
					return fmt.Errorf("rollback contract data fail(%s), ContractSend(%s)", err, blockCur.String())
				}
			}

			// rollback Block
			if err := lv.l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block error: %s, %s", err, hashCur)
			}
			lv.l.EB.Publish(common.EventRollback, hashCur)
			lv.logger.Errorf("rollback delete block %s (previous: %s, type: %s,  address: %s) ", hashCur.String(), blockCur.GetPrevious().String(), blockCur.GetType(), blockCur.GetAddress().String())

			if err := lv.checkBlockCache(blockCur, txn); err != nil {
				lv.logger.Errorf("roll back block cache error : %s", err)
				return err
			}
			if err := lv.rollbackCacheAccountDel(blockCur.GetAddress(), blockCur.GetToken(), txn); err != nil {
				lv.logger.Errorf("roll back account cache error : %s", err)
				return err
			}

			if hashCur == oldestBlock.GetHash() {
				break
			}

			hashCur = blockCur.GetPrevious()
			blockCur, err = lv.l.GetStateBlockConfirmed(hashCur, txn)
			if err != nil {
				return fmt.Errorf("get previous block error %s, %s ", err, hashCur.String())
			}
		}
	}
	return nil
}

func (lv *LedgerVerifier) checkBlockCache(block *types.StateBlock, txn db.StoreTxn) error {
	rollbacks := make([]*types.StateBlock, 0)
	err := lv.l.GetBlockCaches(func(b *types.StateBlock) error {
		if block.GetAddress() == b.GetAddress() && block.GetToken() == b.GetToken() {
			rollbacks = append(rollbacks, b)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err = lv.rollbackCacheBlocks(rollbacks, false, txn); err != nil {
		return err
	}

	if block.IsSendBlock() {
		err := lv.l.GetBlockCaches(func(b *types.StateBlock) error {
			if block.GetHash() == b.GetLink() {
				err = lv.rollbackCache(b.GetHash(), txn)
				if err != nil {
					lv.logger.Error(err)
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) rollBackPendingKind(block *types.StateBlock, txn db.StoreTxn) error {
	if block.IsReceiveBlock() {
		pendingKey := &types.PendingKey{
			Address: block.GetAddress(),
			Hash:    block.GetLink(),
		}
		if err := lv.l.UpdatePending(pendingKey, types.PendingNotUsed, txn); err != nil {
			return err
		}
	}
	return nil
}

// all Send block to rollback
func (lv *LedgerVerifier) sendBlocksInRollback(blocks map[types.Address]*types.StateBlock, txn db.StoreTxn) (map[types.Hash]types.Address, error) {
	sendBlocks := make(map[types.Hash]types.Address)
	for address, oldestBlock := range blocks {
		tm, err := lv.l.GetTokenMetaConfirmed(address, oldestBlock.GetToken(), txn)
		if err != nil {
			return nil, fmt.Errorf("get tokenmeta error: %s (%s)", err, oldestBlock.GetHash().String())
		}

		curHash := tm.Header
		curBlock, err := lv.l.GetStateBlockConfirmed(curHash, txn)
		if err != nil {
			return nil, fmt.Errorf("get block error: %s (%s)", err, curHash.String())
		}
		for {
			if curBlock.IsSendBlock() {
				sendBlocks[curBlock.GetHash()] = curBlock.GetAddress()
			}

			if curBlock.GetHash() == oldestBlock.GetHash() {
				break
			}

			curHash := curBlock.GetPrevious()
			curBlock, err = lv.l.GetStateBlockConfirmed(curHash, txn)
			if err != nil {
				return nil, fmt.Errorf("can not get previous block %s", curHash.String())
			}
		}
	}
	return sendBlocks, nil
}

// if aBlock is created early than bBlock , return true, else return false
func (lv *LedgerVerifier) blockOrderCompare(aBlock, bBlock *types.StateBlock) (bool, error) {
	lv.logger.Debugf("block compare, %s %s %s ", aBlock.GetHash().String(), bBlock.GetHash().String(), aBlock.Address.String())
	if aBlock.GetAddress() != bBlock.GetAddress() || aBlock.GetToken() != bBlock.GetToken() {
		return false, fmt.Errorf("can not compare block, %s %s", aBlock.GetHash().String(), bBlock.GetHash().String())
	}
	tm, _ := lv.l.GetTokenMetaConfirmed(aBlock.GetAddress(), aBlock.GetToken())
	curHash := tm.Header
	for {
		curBlock, err := lv.l.GetStateBlockConfirmed(curHash)
		if err != nil {
			return false, fmt.Errorf("get block error %s", curBlock.String())
		}
		if curBlock.GetHash() == aBlock.GetHash() {
			return false, nil
		}
		if curBlock.GetHash() == bBlock.GetHash() {
			return true, nil
		}
		curHash = curBlock.GetPrevious()
		if curHash.IsZero() {
			return false, fmt.Errorf("can not find blocks when block compare")
		}
	}
}

func (lv *LedgerVerifier) rollBackFrontier(pre types.Hash, cur types.Hash, txn db.StoreTxn) error {
	frontier, err := lv.l.GetFrontier(cur, txn)
	if err != nil {
		return err
	}
	lv.logger.Debug("delete frontier, ", frontier)
	if err := lv.l.DeleteFrontier(cur, txn); err != nil {
		return err
	}
	if !pre.IsZero() {
		frontier.HeaderBlock = pre
		lv.logger.Debug("add frontier, ", frontier)
		if err := lv.l.AddFrontier(frontier, txn); err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) rollBackToken(token *types.TokenMeta, pre *types.StateBlock, txn db.StoreTxn) error {
	ac, err := lv.l.GetAccountMetaConfirmed(token.BelongTo, txn)
	if err != nil {
		return err
	}
	if pre.GetToken() == common.ChainToken() {
		ac.CoinVote = pre.GetVote()
		ac.CoinOracle = pre.GetOracle()
		ac.CoinNetwork = pre.GetNetwork()
		ac.CoinStorage = pre.GetStorage()
		ac.CoinBalance = pre.GetBalance()
	}
	tm := ac.Token(pre.GetToken())
	if tm == nil {
		return fmt.Errorf("can not get token %s from account %s", pre.GetToken().String(), ac.Address.String())
	}
	tm.Balance = pre.GetBalance()
	tm.Header = pre.GetHash()
	tm.Representative = pre.GetRepresentative()
	tm.BlockCount = tm.BlockCount - 1
	tm.Modified = common.TimeNow().Unix()
	lv.logger.Debug("update token, ", tm)
	for index, t := range ac.Tokens {
		if t.Type == tm.Type {
			ac.Tokens[index] = tm
			if err := lv.l.UpdateAccountMeta(ac, txn); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (lv *LedgerVerifier) rollBackTokenDel(tm *types.TokenMeta, txn db.StoreTxn) error {
	address := tm.BelongTo
	lv.logger.Debug("delete token, ", address, tm.Type)
	if err := lv.l.DeleteTokenMeta(address, tm.Type, txn); err != nil {
		return err
	}
	ac, err := lv.l.GetAccountMetaConfirmed(address, txn)
	if err != nil {
		return err
	}
	if len(ac.Tokens) == 0 {
		if err := lv.l.DeleteAccountMeta(address, txn); err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) rollBackRep(representative types.Address, blockCur, blockPre *types.StateBlock, isSend bool, token types.Hash, txn db.StoreTxn) error {
	if token == common.ChainToken() {
		if isSend {
			diff := &types.Benefit{
				Vote:    blockPre.GetVote().Sub(blockCur.GetVote()),
				Network: blockPre.GetNetwork().Sub(blockCur.GetNetwork()),
				Oracle:  blockPre.GetOracle().Sub(blockCur.GetOracle()),
				Storage: blockPre.GetStorage().Sub(blockCur.GetStorage()),
				Balance: blockPre.GetBalance().Sub(blockCur.GetBalance()),
				Total:   blockPre.TotalBalance().Sub(blockCur.TotalBalance()),
			}
			lv.logger.Debugf("add rep(%s) to %s", diff, representative)
			if err := lv.l.AddRepresentation(representative, diff, txn); err != nil {
				return err
			}
		} else {
			diff := new(types.Benefit)
			if blockPre == nil {
				diff = &types.Benefit{
					Vote:    blockCur.GetVote(),
					Network: blockCur.GetNetwork(),
					Oracle:  blockCur.GetOracle(),
					Storage: blockCur.GetStorage(),
					Balance: blockCur.GetBalance(),
					Total:   blockCur.TotalBalance(),
				}
			} else {
				diff = &types.Benefit{
					Vote:    blockCur.GetVote().Sub(blockPre.GetVote()),
					Network: blockCur.GetNetwork().Sub(blockPre.GetNetwork()),
					Oracle:  blockCur.GetOracle().Sub(blockPre.GetOracle()),
					Storage: blockCur.GetStorage().Sub(blockPre.GetStorage()),
					Balance: blockCur.GetBalance().Sub(blockPre.GetBalance()),
					Total:   blockCur.TotalBalance().Sub(blockPre.TotalBalance()),
				}
			}
			lv.logger.Debugf("sub rep %s from %s", diff, representative)
			if err := lv.l.SubRepresentation(representative, diff, txn); err != nil {
				return err
			}
		}
	}
	return nil
}

func (lv *LedgerVerifier) rollBackRepChange(preRepresentation types.Address, curRepresentation types.Address, blockCur *types.StateBlock, txn db.StoreTxn) error {
	diff := &types.Benefit{
		Vote:    blockCur.GetVote(),
		Network: blockCur.GetNetwork(),
		Oracle:  blockCur.GetOracle(),
		Storage: blockCur.GetStorage(),
		Balance: blockCur.GetBalance(),
		Total:   blockCur.TotalBalance(),
	}
	lv.logger.Debugf("add rep(%s) to %s", diff, preRepresentation)
	if err := lv.l.AddRepresentation(preRepresentation, diff, txn); err != nil {
		return err
	}
	lv.logger.Debugf("sub rep(%s) from %s", diff, curRepresentation)
	if err := lv.l.SubRepresentation(curRepresentation, diff, txn); err != nil {
		return err
	}
	return nil
}

func (lv *LedgerVerifier) rollBackPendingAdd(blockCur *types.StateBlock, amount types.Balance, token types.Hash, txn db.StoreTxn) error {
	blockLink, err := lv.l.GetStateBlockConfirmed(blockCur.GetLink(), txn)
	if err != nil {
		return fmt.Errorf("%s %s", err, blockCur.GetLink())
	}

	if blockCur.GetType() == types.ContractReward {
		if c, ok, err := contract.GetChainContract(types.Address(blockLink.Link), blockLink.Data); ok && err == nil {
			switch v := c.(type) {
			case contract.ChainContractV1:
				if pendingKey, pendingInfo, err := v.DoPending(blockLink); err == nil && pendingKey != nil {
					lv.logger.Debug("add contract reward pending , ", pendingKey)
					if err := lv.l.AddPending(pendingKey, pendingInfo, txn); err != nil {
						return err
					}
				}
			case contract.ChainContractV2:
				vmCtx := vmstore.NewVMContext(lv.l)
				if pendingKey, pendingInfo, err := v.ProcessSend(vmCtx, blockLink); err == nil && pendingKey != nil {
					lv.logger.Debug("contractSend add pending , ", pendingKey)
					if err := lv.l.AddPending(pendingKey, pendingInfo, txn); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("process send error, %s", err)
				}
			default:
				return fmt.Errorf("unsupported chain contract %s", reflect.TypeOf(v))
			}
		}
		return nil
	} else {
		pendingkey := types.PendingKey{
			Address: blockCur.GetAddress(),
			Hash:    blockLink.GetHash(),
		}
		pendinginfo := types.PendingInfo{
			Source: blockLink.GetAddress(),
			Amount: amount,
			Type:   token,
		}
		lv.logger.Debug("add pending, ", pendingkey, pendinginfo)
		if err := lv.l.AddPending(&pendingkey, &pendinginfo, txn); err != nil {
			return err
		}
		return nil
	}
}

func (lv *LedgerVerifier) rollBackPendingDel(blockCur *types.StateBlock, txn db.StoreTxn) error {
	if blockCur.GetType() == types.ContractSend {
		if c, ok, err := contract.GetChainContract(types.Address(blockCur.Link), blockCur.Data); ok && err == nil {
			switch v := c.(type) {
			case contract.ChainContractV1:
				if pendingKey, _, err := v.DoPending(blockCur); err == nil && pendingKey != nil {
					lv.logger.Debug("delete contract send pending , ", pendingKey)
					if err := lv.l.DeletePending(pendingKey, txn); err != nil {
						return err
					}
				}
			case contract.ChainContractV2:
				vmCtx := vmstore.NewVMContext(lv.l)
				if pendingKey, _, err := v.ProcessSend(vmCtx, blockCur); err == nil && pendingKey != nil {
					lv.logger.Debug("delete contract send pending , ", pendingKey)
					if err := lv.l.DeletePending(pendingKey, txn); err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unsupported chain contract %s", reflect.TypeOf(v))
			}
		}
		return nil
	} else {
		address := types.Address(blockCur.Link)
		hash := blockCur.GetHash()
		pendingkey := types.PendingKey{
			Address: address,
			Hash:    hash,
		}
		lv.logger.Debug("delete pending ,", pendingkey)
		if err := lv.l.DeletePending(&pendingkey, txn); err != nil {
			return err
		}
		return nil
	}
}

func (lv *LedgerVerifier) rollBackContractData(block *types.StateBlock, txn db.StoreTxn) error {
	extra := block.GetExtra()
	if !extra.IsZero() {
		//t := trie.NewTrie(lv.l.Store, &extra, nil)
		//
		//nodes := t.NewNodeIterator(func(node *trie.TrieNode) bool {
		//	return true
		//})
		//
		//for node := range nodes {
		//	trieData, err := node.Serialize()
		//	if err != nil {
		//		return err
		//	}
		//
		//	vmContext := vmstore.NewVMContext(lv.l)
		//	h := node.Hash()
		//	contractData, _ := vmContext.GetStorage(h[:], nil)
		//	if contractData != nil {
		//		if !bytes.Equal(contractData, trieData) {
		//			return errors.New("contract data invalid")
		//		}
		//		//move contract data to new table
		//		if err := lv.l.AddTemporaryData(h[:], contractData, txn); err != nil {
		//			return err
		//		}
		//		//delete trie and contract data
		//		if err := t.DeleteNode(h[:], txn); err != nil {
		//			return err
		//		}
		//		if err := vmContext.DeleteStorage(h[:], txn); err != nil {
		//			return err
		//		}
		//	}
		//}
	}
	return nil
}
