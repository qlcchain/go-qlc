/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
)

const defaultPassword = ""

var (
	cache             = make(map[string]*WalletStore)
	lock              = sync.RWMutex{}
	ErrEmptyCurrentId = errors.New("can not find any wallet id")
)

func NewWalletStore(cfg *config.Config) *WalletStore {
	lock.Lock()
	defer lock.Unlock()
	dir := cfg.WalletDir()
	logger := log.NewLogger("wallet store")
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

func (ws *WalletStore) WalletIds() ([]types.Address, error) {
	var ids []types.Address
	err := ws.ViewInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixIds}
		return txn.Get(key, func(val []byte, b byte) error {
			if len(val) != 0 {
				return json.Unmarshal(val, &ids)
			}
			return nil
		})
	})

	if err != nil && err == badger.ErrKeyNotFound {
		err = nil
	}

	return ids, err
}

// NewWalletBySeed create wallet from hex seed string
func (ws *WalletStore) NewWalletBySeed(seed, password string) (types.Address, error) {
	var walletId types.Address
	bytes, err := hex.DecodeString(seed)
	if err != nil {
		return walletId, err
	}

	s, err := types.BytesToSeed(bytes)
	if err != nil {
		return walletId, err
	}
	walletId = s.MasterAddress()
	if b, err := ws.IsWalletExist(walletId); b && err == nil {
		return walletId, fmt.Errorf("seed[%s] already exist", seed)
	}

	session := ws.NewSession(walletId)
	ids, err := ws.WalletIds()
	if err != nil {
		return types.ZeroAddress, err
	}

	ids = append(ids, walletId)

	err = ws.UpdateInTx(func(txn db.StoreTxn) error {
		//add new walletId to ids
		key := []byte{idPrefixIds}
		bytes, err := json.Marshal(&ids)
		if err != nil {
			return err
		}

		err = txn.Set(key, bytes)
		if err != nil {
			return err
		}

		// update current wallet id
		err = ws.setCurrentId(txn, walletId.Bytes())
		if err != nil {
			return err
		}

		err = session.setDeterministicIndex(txn, 1)
		if err != nil {
			return err
		}
		_ = session.setVersion(txn, Version)

		err = session.EnterPassword(password)
		if err != nil {
			return err
		}

		err = session.setSeedByTxn(txn, s[:])

		if err != nil {
			return err
		}

		return nil
	})

	return walletId, err
}

// IsWalletExist check is the wallet exist by master address
func (ws *WalletStore) IsWalletExist(address types.Address) (bool, error) {
	addresses, err := ws.WalletIds()
	if err != nil {
		return false, err
	}
	for _, addr := range addresses {
		if addr == address {
			return true, nil
		}
	}
	return false, nil
}

//NewWallet create new wallet and save to db
func (ws *WalletStore) NewWallet() (types.Address, error) {
	seed, err := types.NewSeed()
	if err != nil {
		return types.ZeroAddress, err
	}

	return ws.NewWalletBySeed(seed.String(), defaultPassword)
}

func (ws *WalletStore) CurrentId() (types.Address, error) {
	var id types.Address

	err := ws.ViewInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixId}
		return txn.Get(key, func(val []byte, b byte) error {
			if len(val) == 0 {
				return ErrEmptyCurrentId
			}
			addr, err := types.BytesToAddress(val)
			id = addr
			return err
		})
	})

	return id, err
}

func (ws *WalletStore) RemoveWallet(id types.Address) error {
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
		newId = ids[0].Bytes()
	} else {
		newId, _ = hex.DecodeString("")
	}

	return ws.UpdateInTx(func(txn db.StoreTxn) error {
		//update ids
		key := []byte{idPrefixIds}
		bytes, err := json.Marshal(&ids)
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
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[ws.dir]; ok {
		err := ws.Store.Close()
		delete(cache, ws.dir)
		ws.logger.Info("wallet closed , ", ws.dir)
		return err
	}
	return nil
}

func (ws *WalletStore) setCurrentId(txn db.StoreTxn, walletId []byte) error {
	key := []byte{idPrefixId}
	return txn.Set(key, walletId)
}

func indexOf(ids []types.Address, id types.Address) (int, error) {
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

func remove(ids []types.Address, id types.Address) ([]types.Address, error) {
	if i, err := indexOf(ids, id); err == nil {
		ids = append(ids[:i], ids[i+1:]...)
		return ids, nil
	} else {
		return ids, err
	}

}
