/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import "github.com/qlcchain/go-qlc/ledger/db"

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

func (ws *WalletStore) Close() error {
	return ws.store.Close()
}

func (ws *WalletStore) Purge() error {
	return ws.store.Purge()
}

func (ws *WalletStore) View(fn func(txn db.StoreTxn) error) error {
	return ws.store.View(func(txn db.StoreTxn) error {
		ws.txn = txn
		return fn(txn)
	})
}

func (ws *WalletStore) Update(fn func(txn db.StoreTxn) error) error {
	return ws.store.Update(func(txn db.StoreTxn) error {
		ws.txn = txn
		return fn(txn)
	})
}

func (ws *WalletStore) Upgrade(migrations []*Migration) error {
	return nil
}
