package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) AddStateBlock(value *types.StateBlock, txns ...db.StoreTxn) error {
	if err := l.addStateBlock(value, txns...); err != nil {
		return err
	}
	l.logger.Debug("publish addRelation,", value.GetHash())
	l.EB.Publish(common.EventAddRelation, value)
	return nil
}

func (l *Ledger) AddSyncStateBlock(value *types.StateBlock, txns ...db.StoreTxn) error {
	if err := l.addStateBlock(value, txns...); err != nil {
		return err
	}
	l.logger.Debug("publish sync addRelation,", value.GetHash())
	l.EB.Publish(common.EventAddSyncBlocks, value, false)
	return nil
}

func (l *Ledger) addStateBlock(value *types.StateBlock, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)

	k, err := getKeyOfParts(idPrefixBlock, value.GetHash())
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	if err := txn.Set(k, v); err != nil {
		return err
	}

	if b, err := l.HasBlockCache(value.GetHash()); b && err == nil {
		if err := l.DeleteBlockCache(value.GetHash(), txn); err != nil {
			return fmt.Errorf("delete block cache error: %s", err)
		}
	}
	if err := addChild(value, txn); err != nil {
		return fmt.Errorf("add block child error: %s", err)
	}
	if err := addLink(value, txn); err != nil {
		return fmt.Errorf("add block link error: %s", err)
	}
	l.releaseTxn(txn, flag)
	return nil
}

func (l *Ledger) GetStateBlock(key types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	if blkCache, err := l.GetBlockCache(key); err == nil {
		return blkCache, nil
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlock, key)
	if err != nil {
		return nil, err
	}

	value := new(types.StateBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetStateBlockConfirmed(key types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlock, key)
	if err != nil {
		return nil, err
	}

	value := new(types.StateBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetStateBlocks(fn func(*types.StateBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			l.logger.Errorf("deserialize block error: %s", err)
			errStr = append(errStr, err.Error())
			return nil
		}
		if err := fn(blk); err != nil {
			l.logger.Errorf("process block error: %s", err)
			errStr = append(errStr, err.Error())
		}
		return nil
	})

	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}

func (l *Ledger) DeleteStateBlock(key types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)

	k, err := getKeyOfParts(idPrefixBlock, key)
	if err != nil {
		return err
	}

	blk := new(types.StateBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := blk.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := txn.Delete(k); err != nil {
		return err
	}

	if err := l.deleteChild(blk, txn); err != nil {
		return fmt.Errorf("delete child error: %s", err)
	}
	if err := l.deleteLink(blk, txn); err != nil {
		return fmt.Errorf("delete link error: %s", err)
	}

	l.releaseTxn(txn, flag)
	l.logger.Info("publish deleteRelation,", key.String())
	l.EB.Publish(common.EventDeleteRelation, key)
	return nil
}

func (l *Ledger) HasStateBlock(key types.Hash, txns ...db.StoreTxn) (bool, error) {
	if exit, err := l.HasBlockCache(key); err == nil && exit {
		return exit, nil
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlock, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) HasStateBlockConfirmed(key types.Hash, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlock, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) CountStateBlocks(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Count([]byte{idPrefixBlock})
}

func (l *Ledger) GetRandomStateBlock(txns ...db.StoreTxn) (*types.StateBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	c, err := l.CountStateBlocks()
	if err != nil {
		return nil, err
	}
	if c == 0 {
		return nil, ErrStoreEmpty
	}
	blk := new(types.StateBlock)
	errFound := errors.New("state block found")

	for i := 0; i < 3; i++ {
		index := rand.Int63n(int64(c))
		var temp int64
		err = txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
			if temp == index {
				var b = new(types.StateBlock)
				if err = b.Deserialize(val); err != nil {
					return err
				}
				if !common.IsGenesisBlock(b) {
					blk = b
					return errFound
				}
			}
			temp++
			return nil
		})
		if err != nil && err != errFound {
			return nil, err
		}
		if !blk.Token.IsZero() {
			break
		}
	}
	if blk.Token.IsZero() {
		return nil, errors.New("state block not found")
	}
	return blk, nil
}

func (l *Ledger) AddSmartContractBlock(value *types.SmartContractBlock, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixSmartContractBlock, value.GetHash())
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetSmartContractBlock(key types.Hash, txns ...db.StoreTxn) (*types.SmartContractBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixSmartContractBlock, key)
	if err != nil {
		return nil, err
	}

	value := new(types.SmartContractBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetSmartContractBlocks(fn func(block *types.SmartContractBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	err := txn.Iterator(idPrefixSmartContractBlock, func(key []byte, val []byte, b byte) error {
		blk := new(types.SmartContractBlock)
		if err := blk.Deserialize(val); err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		if err := fn(blk); err != nil {
			errStr = append(errStr, err.Error())
		}
		return nil
	})

	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}

func (l *Ledger) HasSmartContractBlock(key types.Hash, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixSmartContractBlock, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) CountSmartContractBlocks(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	return txn.Count([]byte{idPrefixSmartContractBlock})
}

func addChild(cBlock *types.StateBlock, txn db.StoreTxn) error {
	pHash := cBlock.Parent()
	cHash := cBlock.GetHash()
	if !common.IsGenesisBlock(cBlock) && pHash != types.ZeroHash && !cBlock.IsOpen() {
		// is parent block existed
		kCache, err := getKeyOfParts(idPrefixBlockCache, pHash)
		if err != nil {
			return err
		}
		err = txn.Get(kCache, func(v []byte, b byte) error {
			return nil
		})
		if err != nil {
			kConfirmed, err := getKeyOfParts(idPrefixBlock, pHash)
			if err != nil {
				return err
			}
			err = txn.Get(kConfirmed, func(v []byte, b byte) error {
				return nil
			})
			if err != nil {
				return fmt.Errorf("%s can not find parent %s", cHash.String(), pHash.String())
			}
		}

		// is parent have used
		k, err := getKeyOfParts(idPrefixChild, pHash)
		if err != nil {
			return err
		}

		err = txn.Get(k, func(val []byte, b byte) error {
			return nil
		})
		if err == nil {
			return fmt.Errorf("%s already have child ", pHash.String())
		}

		// add new relationship
		v, err := cHash.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := txn.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func addLink(block *types.StateBlock, txn db.StoreTxn) error {
	if block.GetType() == types.Open || block.GetType() == types.Receive || block.GetType() == types.ContractReward {
		k, err := getKeyOfParts(idPrefixLink, block.GetLink())
		if err != nil {
			return err
		}
		h := block.GetHash()
		v, err := h.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := txn.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) deleteChild(blk *types.StateBlock, txn db.StoreTxn) error {
	pHash := blk.Parent()
	if !pHash.IsZero() {

		k, err := getKeyOfParts(idPrefixChild, pHash)
		if err != nil {
			return err
		}
		if err := txn.Delete(k); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) deleteLink(blk *types.StateBlock, txn db.StoreTxn) error {
	k, err := getKeyOfParts(idPrefixLink, blk.GetLink())
	if err != nil {
		return err
	}
	if err := txn.Delete(k); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetChild(key types.Hash, txns ...db.StoreTxn) (types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixChild, key)
	if err != nil {
		return types.ZeroHash, err
	}

	value := new(types.Hash)
	err = txn.Get(k, func(val []byte, b byte) error {
		if _, err := value.UnmarshalMsg(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return types.ZeroHash, err
	}
	return *value, nil
}

func (l *Ledger) GetLinkBlock(key types.Hash, txns ...db.StoreTxn) (types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixLink, key)
	if err != nil {
		return types.ZeroHash, err
	}
	value := new(types.Hash)
	err = txn.Get(k, func(val []byte, b byte) (err error) {
		if _, err := value.UnmarshalMsg(val); err != nil {
			return errors.New("unmarshal hash error")
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return types.ZeroHash, ErrLinkNotFound
		}
		return types.ZeroHash, fmt.Errorf("get link error: %s", err)
	}
	return *value, nil
}

func (l *Ledger) GetGenesis(txns ...db.StoreTxn) ([]*types.StateBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var blocks []*types.StateBlock
	err := txn.Iterator(idPrefixToken, func(key []byte, val []byte, b byte) error {
		t := new(types.TokenInfo)
		if err := json.Unmarshal(val, t); err != nil {
			return err
		}
		tm, err := l.GetTokenMeta(t.Owner, t.TokenId, txn)
		if err != nil {
			return err
		}
		block, err := l.GetStateBlock(tm.OpenBlock, txn)
		if err != nil {
			return err
		}
		blocks = append(blocks, block)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (l *Ledger) AddBlockCache(value *types.StateBlock, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCache, value.GetHash())
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetBlockCache(key types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCache, key)
	if err != nil {
		return nil, err
	}

	value := new(types.StateBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) HasBlockCache(key types.Hash, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCache, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) DeleteBlockCache(key types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCache, key)
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err != nil {
		return err
	}

	if err := txn.Delete(k); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) CountBlockCache(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Count([]byte{idPrefixBlockCache})
}

func (l *Ledger) GetBlockCaches(fn func(*types.StateBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixBlockCache, func(key []byte, val []byte, b byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			l.logger.Errorf("deserialize block error: %s", err)
			return nil
		}
		if err := fn(blk); err != nil {
			l.logger.Errorf("process block error: %s", err)
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) AddMessageInfo(key types.Hash, value []byte, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixMessageInfo, key)
	if err != nil {
		return err
	}
	if err := txn.Set(k, value); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetMessageInfo(key types.Hash, txns ...db.StoreTxn) ([]byte, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixMessageInfo, key)
	if err != nil {
		return nil, err
	}
	var value []byte
	err = txn.Get(k, func(v []byte, b byte) error {
		value = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (l *Ledger) AddSyncCacheBlock(value *types.StateBlock, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixSyncCacheBlock, value.GetHash())
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetSyncCacheBlock(hash types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixSyncCacheBlock, hash)
	if err != nil {
		return nil, err
	}

	value := new(types.StateBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetSyncCacheBlocks(fn func(*types.StateBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixSyncCacheBlock, func(key []byte, val []byte, b byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			return nil
		}
		if err := fn(blk); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) DeleteSyncCacheBlock(key types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixSyncCacheBlock, key)
	if err != nil {
		return err
	}
	return txn.Delete(k)
}
