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
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/ledger/db"
)

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

	return ids, err
}

//NewWallet create new wallet and save to db
// TODO: handle error
func (ws *WalletStore) NewWallet() (WalletId, error) {
	walletId := NewWalletId()

	ids, err := ws.WalletIds()
	if err != nil {
		return walletId, err
	}
	key, _ := walletId.MarshalBinary()
	session := ws.NewSession(key)
	err = session.Init()
	if err != nil {
		return walletId, err
	}
	ids = append(ids, walletId)
	err = ws.UpdateInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixIds}
		bytes, err := jsoniter.Marshal(&ids)
		if err != nil {
			return err
		}
		return txn.Set(key, bytes)
	})
	currentId, _ := walletId.MarshalBinary()
	err = ws.setCurrentId(currentId)

	return walletId, err
}

func (ws *WalletStore) CurrentId() (WalletId, error) {
	var id WalletId

	err := ws.ViewInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixId}
		return txn.Get(key, func(val []byte, b byte) error {
			if len(val) == 0 {
				return errors.New("can not find any wallet id")
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
	walletId, _ := id.MarshalBinary()
	session := ws.NewSession(walletId)
	err = session.Remove()

	if err != nil {
		return err
	}
	var newId []byte
	if len(ids) > 0 {
		newId, _ = ids[0].MarshalBinary()
	} else {
		newId, _ = hex.DecodeString("")
	}
	err = ws.setCurrentId(newId)
	if err != nil {
		return err
	}

	return err
}

func (ws *WalletStore) setCurrentId(walletId []byte) error {
	return ws.UpdateInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixId}
		return txn.Set(key, walletId)
	})
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
