/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"sync"
)

var (
	cache             = make(map[string]*WalletStore)
	lock              = sync.RWMutex{}
	logger            = log.NewLogger("wallet store")
	ErrEmptyCurrentId = errors.New("can not find any wallet id")
)

func NewWalletStore(cfg *config.Config) *WalletStore {
	lock.RLock()
	defer lock.RUnlock()
	dir := cfg.WalletDir()
	if _, ok := cache[dir]; !ok {
		store, err := db.NewBadgerStore(dir)
		if err != nil {
			logger.Fatal(err.Error())
		}

		cache[dir] = &WalletStore{
			ledger: ledger.NewLedger(cfg.LedgerDir()),
			logger: logger,
			Store:  store,
			dir:    dir,
		}
	}
	return cache[dir]
}

func (ws *WalletStore) WalletIds() ([]WalletId, error) {
	var ids []WalletId
	err := ws.ViewInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixIds}
		return txn.Get(key, func(val []byte, b byte) error {
			if len(val) != 0 {
				return jsoniter.Unmarshal(val, &ids)
			}
			return nil
		})
	})

	if err != nil && err == badger.ErrKeyNotFound {
		err = nil
	}

	return ids, err
}

//NewWallet create new wallet and save to db
func (ws *WalletStore) NewWallet() (WalletId, error) {
	walletId := NewWalletId()

	ids, err := ws.WalletIds()
	if err != nil {
		return walletId, err
	}

	session := ws.NewSession(walletId)

	ids = append(ids, walletId)

	err = ws.UpdateInTx(func(txn db.StoreTxn) error {
		//add new walletId to ids
		key := []byte{idPrefixIds}
		bytes, err := jsoniter.Marshal(&ids)
		if err != nil {
			return err
		}

		err = txn.Set(key, bytes)
		if err != nil {
			return err
		}

		// update current wallet id
		currentId, _ := walletId.MarshalBinary()
		err = ws.setCurrentId(txn, currentId)
		if err != nil {
			return err
		}

		err = session.setDeterministicIndex(txn, 1)
		if err != nil {
			return err
		}
		_ = session.setVersion(txn, Version)
		//default password is empty
		err = session.EnterPassword("")
		if err != nil {
			return err
		}

		seed, err := types.NewSeed()
		if err != nil {
			return err
		}
		err = session.setSeed(txn, seed[:])

		if err != nil {
			return err
		}

		return nil
	})

	return walletId, err
}

func (ws *WalletStore) CurrentId() (WalletId, error) {
	var id WalletId

	err := ws.ViewInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixId}
		return txn.Get(key, func(val []byte, b byte) error {
			if len(val) == 0 {
				return ErrEmptyCurrentId
			}
			return id.UnmarshalBinary(val)
		})
	})

	return id, err
}

func (ws *WalletStore) RemoveWallet(id WalletId) error {
	ids, err := ws.WalletIds()
	if err != nil {
		return err
	}

	ids, err = remove(ids, id)
	if err != nil {
		return err
	}

	var newId []byte
	if len(ids) > 0 {
		newId, _ = ids[0].MarshalBinary()
	} else {
		newId, _ = hex.DecodeString("")
	}

	return ws.UpdateInTx(func(txn db.StoreTxn) error {
		//update ids
		key := []byte{idPrefixIds}
		bytes, err := jsoniter.Marshal(&ids)
		if err != nil {
			return err
		}
		err = txn.Set(key, bytes)
		if err != nil {
			return err
		}
		//update current id
		err = ws.setCurrentId(txn, newId)
		if err != nil {
			return err
		}

		// remove wallet data by walletId
		session := ws.NewSession(id)
		err = session.removeWallet(txn)
		if err != nil {
			return err
		}
		return nil
	})
}

func (ws *WalletStore) Close() error {
	lock.RLock()
	defer lock.RUnlock()
	err := ws.Store.Close()
	delete(cache, ws.dir)
	return err
}

func (ws *WalletStore) setCurrentId(txn db.StoreTxn, walletId []byte) error {
	key := []byte{idPrefixId}
	return txn.Set(key, walletId)
}

func indexOf(ids []WalletId, id WalletId) (int, error) {
	index := -1

	for i, _id := range ids {
		if id == _id {
			index = i
			break
		}
	}

	if index < 0 {
		return -1, fmt.Errorf("can not find id(%s)", id.String())
	}

	return index, nil
}

func remove(ids []WalletId, id WalletId) ([]WalletId, error) {
	if i, err := indexOf(ids, id); err == nil {
		ids = append(ids[:i], ids[i+1:]...)
		return ids, nil
	} else {
		return ids, err
	}

}
