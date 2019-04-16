/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"bytes"
	"errors"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type ContractStore interface {
	GetStorage(prefix, key []byte) ([]byte, error)
	SetStorage(prefix, key []byte, value []byte) error
	Iterator(prefix []byte, fn func(key []byte, value []byte) error) error

	CalculateAmount(block *types.StateBlock) (types.Balance, error)
	IsUserAccount(address types.Address) (bool, error)
	GetAccountMeta(address types.Address) (*types.AccountMeta, error)
}

const (
	idPrefixStorage = 100
)

var (
	ErrStorageExists   = errors.New("storage already exists")
	ErrStorageNotFound = errors.New("storage not found")
)

type VMContext struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

func (v *VMContext) IsUserAccount(address types.Address) (bool, error) {
	if _, err := v.ledger.HasAccountMeta(address); err == nil {
		return true, nil
	} else {
		return false, err
	}
}

func NewVMContext(l *ledger.Ledger) *VMContext {
	context := &VMContext{ledger: l}
	context.logger = log.NewLogger("vm_context")
	return context
}

func getStorageKey(prefix, key []byte) []byte {
	var storageKey []byte
	storageKey = append(storageKey, []byte{idPrefixStorage}...)
	storageKey = append(storageKey, prefix...)
	storageKey = append(storageKey, key...)
	return storageKey
}

func (v *VMContext) GetStorage(prefix, key []byte) ([]byte, error) {
	txn := v.ledger.Store.NewTransaction(false)
	defer func() {
		txn.Commit(nil)
		txn.Discard()
	}()
	storageKey := getStorageKey(prefix, key)
	var storage []byte
	err := txn.Get(storageKey, func(val []byte, b byte) (err error) {
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

func (v *VMContext) SetStorage(prefix, key []byte, value []byte) error {
	txn := v.ledger.Store.NewTransaction(true)
	defer func() {
		txn.Commit(nil)
		txn.Discard()
	}()
	storageKey := getStorageKey(prefix, key)
	err := txn.Get(storageKey, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrStorageExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(storageKey, value)
}

func (v *VMContext) Iterator(prefix []byte, fn func(key []byte, value []byte) error) error {
	txn := v.ledger.Store.NewTransaction(false)
	defer func() {
		txn.Discard()
	}()

	err := txn.Iterator(idPrefixStorage, func(key []byte, val []byte, b byte) error {
		if bytes.HasPrefix(key[1:], prefix) {
			err := fn(key, val)
			if err != nil {
				v.logger.Error(err)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (v *VMContext) CalculateAmount(block *types.StateBlock) (types.Balance, error) {
	return v.ledger.CalculateAmount(block)

}

func (v *VMContext) GetAccountMeta(address types.Address) (*types.AccountMeta, error) {
	return v.ledger.GetAccountMeta(address)
}
