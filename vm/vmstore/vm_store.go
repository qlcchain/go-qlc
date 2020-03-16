/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"go.uber.org/zap"
)

type ContractStore interface {
	GetStorage(prefix, key []byte) ([]byte, error)
	SetStorage(prefix, key []byte, value []byte) error
	Iterator(prefix []byte, fn func(key []byte, value []byte) error) error

	//CalculateAmount(block *types.StateBlock) (types.Balance, error)
	IsUserAccount(address types.Address) (bool, error)
	//GetAccountMeta(address types.Address) (*types.AccountMeta, error)
	//GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error)
	//GetStateBlock(hash types.Hash) (*types.StateBlock, error)
	//HasTokenMeta(address types.Address, token types.Hash) (bool, error)

	SaveStorage(batch ...storage.Batch) error
	RemoveStorage(prefix, key []byte, batch ...storage.Batch) error

	// TODO: remove
	GetStorageByKey(key []byte) ([]byte, error)
	RemoveStorageByKey(key []byte, batch ...storage.Batch) error
}

var (
	//ErrStorageExists   = errors.New("storage already exists")
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
	if b, err := v.Ledger.HasAccountMetaConfirmed(address); b {
		return true, nil
	} else {
		if err != nil {
			return false, err
		} else {
			return false, fmt.Errorf("can not find user account %s", address)
		}
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
	if s := v.Cache.GetStorage(storageKey); s == nil {
		if val, err := v.get(storageKey); err == nil {
			return val, nil
		} else {
			return nil, err
		}
	} else {
		return s, nil
	}
}

func (v *VMContext) GetStorageByKey(key []byte) ([]byte, error) {
	if s := v.Cache.GetStorage(key); s == nil {
		if val, err := v.get(key); err == nil {
			return val, nil
		} else {
			return nil, err
		}
	} else {
		return s, nil
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
	err := v.Ledger.DBStore().Iterator(pre, nil, func(key []byte, val []byte) error {
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

func (v *VMContext) SaveStorage(batch ...storage.Batch) error {
	for k, val := range v.Cache.storage {
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
	bytes := make([]byte, len(value))
	copy(bytes, value)
	return b.Put(key, bytes)
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
