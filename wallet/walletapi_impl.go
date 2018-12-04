/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"github.com/qlcchain/go-qlc/ledger/db"
	"golang.org/x/crypto/argon2"
)

const (
	idPrefixId byte = iota
	idPrefixVersion
	idPrefixSalt
	idPrefixWork
	idPrefixAccounts
	idPrefixIndex
	idPrefixSeed
	idPrefixRepresentation
	idPrefixCheck
)

const KdfWork = 64 * 1024

type WalletStore struct {
	store   db.BadgerStore
	version int64 // current version
}

func (ws *WalletStore) encrypt(password string, salt []byte) []byte {
	key := argon2.IDKey([]byte(password), salt, 1, KdfWork, 4, 32)
	return key
}

func (ws *WalletStore) Close() error {
	return ws.store.Close()
}

func (ws *WalletStore) Purge() error {
	return ws.store.Purge()
}

func (ws *WalletStore) View(fn func(txn db.StoreTxn) error) error {
	return ws.store.View(func(txn db.StoreTxn) error {
		return fn(txn)
	})
}

func (ws *WalletStore) Update(fn func(txn db.StoreTxn) error) error {
	return ws.store.Update(func(txn db.StoreTxn) error {
		return fn(txn)
	})
}

func (ws *WalletStore) Upgrade(migrations []*Migration) error {
	return nil
}
