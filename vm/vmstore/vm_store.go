/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"errors"
	"github.com/dgraph-io/badger"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
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
	SaveStorage(txns ...db.StoreTxn) error
	RemoveStorage(prefix, key []byte, txns ...db.StoreTxn) error
}

const (
	idPrefixStorage = 100
)

var (
	ErrStorageExists   = errors.New("storage already exists")
	ErrStorageNotFound = errors.New("storage not found")
)

type VMContext struct {
	Ledger *ledger.Ledger
	logger *zap.SugaredLogger
	Cache  *VMCache
}

func NewVMContext(l *ledger.Ledger) *VMContext {
	t := trie.NewTrie(l.Store, nil, trie.NewSimpleTrieNodePool())
	return &VMContext{
		Ledger: l,
		logger: log.NewLogger("vm_context"),
		Cache:  NewVMCache(t),
	}
}

func (v *VMContext) IsUserAccount(address types.Address) (bool, error) {
	if _, err := v.Ledger.HasAccountMeta(address); err == nil {
		return true, nil
	} else {
		return false, err
	}
}

func getStorageKey(prefix, key []byte) []byte {
	var storageKey []byte
	storageKey = append(storageKey, []byte{idPrefixStorage}...)
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

func (v *VMContext) RemoveStorage(prefix, key []byte, txns ...db.StoreTxn) error {
	storageKey := getStorageKey(prefix, key)
	v.Cache.RemoveStorage(storageKey)
	return v.remove(storageKey, txns...)
}

func (v *VMContext) RemoveStorageByKey(key []byte, txns ...db.StoreTxn) error {
	v.Cache.RemoveStorage(key)
	return v.remove(key, txns...)
}

func (v *VMContext) SetStorage(prefix, key []byte, value []byte) error {
	storageKey := getStorageKey(prefix, key)

	v.Cache.SetStorage(storageKey, value)

	return nil
}

func (v *VMContext) Iterator(prefix []byte, fn func(key []byte, value []byte) error) error {
	txn := v.Ledger.Store.NewTransaction(false)
	defer func() {
		txn.Discard()
	}()

	pre := make([]byte, 0)
	pre = append(pre, idPrefixStorage)
	pre = append(pre, prefix...)
	err := txn.PrefixIterator(pre, func(key []byte, val []byte, b byte) error {
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

func (v *VMContext) SaveStorage(txns ...db.StoreTxn) error {
	storage := v.Cache.storage
	for k, val := range storage {
		err := v.set([]byte(k), val)
		if err != nil {
			v.logger.Error(err)
			return err
		}
	}
	return nil
}

func (v *VMContext) SaveTrie(txns ...db.StoreTxn) error {
	fn, err := v.Cache.Trie().Save(txns...)
	if err != nil {
		return err
	}
	fn()
	return nil
}

func (v *VMContext) get(key []byte) ([]byte, error) {
	txn := v.Ledger.Store.NewTransaction(false)
	defer func() {
		txn.Commit()
		txn.Discard()
	}()

	var storage []byte
	err := txn.Get(key, func(val []byte, b byte) (err error) {
		storage = val
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}
	return storage, nil
}

func (v *VMContext) set(key []byte, value []byte, txns ...db.StoreTxn) (err error) {
	var txn db.StoreTxn
	if len(txns) > 0 {
		txn = txns[0]
	} else {
		txn = v.Ledger.Store.NewTransaction(true)
		defer func() {
			err = txn.Commit()
			txn.Discard()
		}()
	}
	return txn.Set(key, value)
}

func (v *VMContext) remove(key []byte, txns ...db.StoreTxn) (err error) {
	var txn db.StoreTxn
	if len(txns) > 0 {
		txn = txns[0]
	} else {
		txn = v.Ledger.Store.NewTransaction(true)
		defer func() {
			err = txn.Commit()
			txn.Discard()
		}()
	}
	return txn.Delete(key)
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
