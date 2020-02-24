/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"errors"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"go.uber.org/zap"
)

type ContractStore interface {
	GetStorage(prefix, key []byte) ([]byte, error)
	SetStorage(prefix, key []byte, value []byte) error
	DelStorage(prefix, key []byte)
	Iterator(prefix []byte, fn func(key []byte, value []byte) error) error

	CalculateAmount(block *types.StateBlock) (types.Balance, error)
	IsUserAccount(address types.Address) (bool, error)
	GetAccountMeta(address types.Address) (*types.AccountMeta, error)
	GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error)
	GetStateBlock(hash types.Hash) (*types.StateBlock, error)
	HasTokenMeta(address types.Address, token types.Hash) (bool, error)
	SaveStorage(batch ...storage.Batch) error
	RemoveStorage(prefix, key []byte, batch ...storage.Batch) error
}

//const (
//	idPrefixStorage = 100
//)

var (
	ErrStorageExists   = errors.New("storage already exists")
	ErrStorageNotFound = errors.New("storage not found")
)

type VMContext struct {
	Ledger ledger.Store
	logger *zap.SugaredLogger
	Cache  *VMCache
}

func NewVMContext(l ledger.Store) *VMContext {
	t := trie.NewTrie(l.DBStore(), nil, trie.NewSimpleTrieNodePool())
	return &VMContext{
		Ledger: l,
		logger: log.NewLogger("vm_context"),
		Cache:  NewVMCache(t),
	}
}

func (v *VMContext) IsUserAccount(address types.Address) (bool, error) {
	if _, err := v.Ledger.HasAccountMetaConfirmed(address); err == nil {
		return true, nil
	} else {
		v.logger.Error("account not found", address)
		return false, err
	}
}

func getStorageKey(prefix, key []byte) []byte {
	var storageKey []byte
	storageKey = append(storageKey, []byte{byte(storage.KeyPrefixTrieVMStorage)}...)
	storageKey = append(storageKey, prefix...)
	storageKey = append(storageKey, key...)
	return storageKey
}

func (v *VMContext) GetStorage(prefix, key []byte) ([]byte, error) {
	storageKey := getStorageKey(prefix, key)
	if storage := v.Cache.GetStorage(storageKey); storage == nil {
		if val, err := v.get(storageKey); err == nil {
			return val, nil
		} else {
			return nil, err
		}
	} else {
		return storage, nil
	}
}

func (v *VMContext) GetStorageByKey(key []byte) ([]byte, error) {
	if storage := v.Cache.GetStorage(key); storage == nil {
		if val, err := v.get(key); err == nil {
			return val, nil
		} else {
			return nil, err
		}
	} else {
		return storage, nil
	}
}

func (v *VMContext) RemoveStorage(prefix, key []byte, batch ...storage.Batch) error {
	storageKey := getStorageKey(prefix, key)
	v.Cache.RemoveStorage(storageKey)
	return v.remove(storageKey, batch...)
}

func (v *VMContext) RemoveStorageByKey(key []byte, batch ...storage.Batch) error {
	v.Cache.RemoveStorage(key)
	return v.remove(key, batch...)
}

func (v *VMContext) SetStorage(prefix, key []byte, value []byte) error {
	storageKey := getStorageKey(prefix, key)

	v.Cache.SetStorage(storageKey, value)

	return nil
}

func (v *VMContext) IteratorAll(prefix []byte, fn func(key []byte, value []byte) error) error {
	pre := make([]byte, 0)
	pre = append(pre, byte(storage.KeyPrefixTrieVMStorage))
	pre = append(pre, prefix...)

	err := v.Ledger.Iterator(pre, nil, func(key []byte, val []byte) error {
		err := fn(key, val)
		if err != nil {
			v.logger.Error(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (v *VMContext) Iterator(prefix []byte, fn func(key []byte, value []byte) error) error {
	pre := make([]byte, 0)
	pre = append(pre, byte(storage.KeyPrefixTrieVMStorage))
	pre = append(pre, prefix...)
	err := v.Ledger.Iterator(pre, nil, func(key []byte, val []byte) error {
		err := fn(key, val)
		if err != nil {
			v.logger.Error(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (v *VMContext) CalculateAmount(block *types.StateBlock) (types.Balance, error) {
	b, err := v.Ledger.CalculateAmount(block)
	if err != nil {
		v.logger.Error("calculate amount error: ", err)
		return types.ZeroBalance, err
	}
	return b, nil
}

func (v *VMContext) GetAccountMeta(address types.Address) (*types.AccountMeta, error) {
	return v.Ledger.GetAccountMeta(address)
}

func (v *VMContext) SaveStorage(batch ...storage.Batch) error {
	storage := v.Cache.storage
	for k, val := range storage {
		err := v.set([]byte(k), val, batch...)
		if err != nil {
			v.logger.Error(err)
			return err
		}
	}
	return nil
}

func (v *VMContext) SaveTrie(batch ...storage.Batch) error {
	fn, err := v.Cache.Trie().Save(batch...)
	if err != nil {
		return err
	}
	fn()
	return nil
}

func (v *VMContext) get(key []byte) ([]byte, error) {
	i, val, err := v.Ledger.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}
	if i != nil {
		return i.([]byte), nil
	}
	return val, nil

	//
	//if v, err := v.Ledger.Cache().Get(key); err == nil {
	//	return v.([]byte), nil
	//}
	//val, err := v.Ledger.Store().Get(key)
	//if err != nil {
	//	if err == storage.KeyNotFound {
	//		return nil, ErrStorageNotFound
	//	}
	//	return nil, err
	//}
	//
	//s = val
	//return s, nil
}

func (v *VMContext) set(key []byte, value []byte, batch ...storage.Batch) (err error) {
	var b storage.Batch
	if len(batch) > 0 {
		b = batch[0]
	} else {
		b = v.Ledger.DBStore().Batch(true)
		defer func() {
			if err := v.Ledger.DBStore().PutBatch(b); err != nil {
				v.logger.Error(err)
			}
		}()
	}
	return b.Put(key, value)
}

func (v *VMContext) remove(key []byte, batch ...storage.Batch) (err error) {
	var b storage.Batch
	if len(batch) > 0 {
		b = batch[0]
	} else {
		b = v.Ledger.DBStore().Batch(true)
		defer func() {
			if err := v.Ledger.DBStore().PutBatch(b); err != nil {
				v.logger.Error(err)
			}
		}()
	}
	return b.Delete(key)
}

func (v *VMContext) HasTokenMeta(address types.Address, token types.Hash) (bool, error) {
	return v.Ledger.HasTokenMeta(address, token)
}

func (v *VMContext) GetTokenMeta(address types.Address, token types.Hash) (*types.TokenMeta, error) {
	return v.Ledger.GetTokenMeta(address, token)
}

func (v *VMContext) GetRepresentation(address types.Address) (*types.Benefit, error) {
	return v.Ledger.GetRepresentation(address)
}

func (v *VMContext) GetStateBlock(hash types.Hash) (*types.StateBlock, error) {
	return v.Ledger.GetStateBlock(hash)
}

func (v *VMContext) GetPovHeaderByHeight(height uint64) (*types.PovHeader, error) {
	return v.Ledger.GetPovHeaderByHeight(height)
}

func (v *VMContext) GetPovBlockByHeight(height uint64) (*types.PovBlock, error) {
	return v.Ledger.GetPovBlockByHeight(height)
}

func (v *VMContext) GetLatestPovBlock() (*types.PovBlock, error) {
	return v.Ledger.GetLatestPovBlock()
}

func (v *VMContext) GetPovMinerStat(dayIndex uint32) (*types.PovMinerDayStat, error) {
	return v.Ledger.GetPovMinerStat(dayIndex)
}

func (v *VMContext) GetPovTxLookup(txHash types.Hash) (*types.PovTxLookup, error) {
	return v.Ledger.GetPovTxLookup(txHash)
}
