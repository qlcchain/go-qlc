/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"encoding/binary"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"golang.org/x/crypto/argon2"
	"hash"
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
	idPrefixWallet
)

const KdfWork = 64 * 1024

var log = common.NewLogger("wallet")

type WalletStore struct {
	store   db.BadgerStore
	version int64 // current version
}

func (ws *WalletStore) HasPoW() bool {
	panic("implement me")
}

func (ws *WalletStore) WaitPoW() {
	panic("implement me")
}

func (ws *WalletStore) WaitingForPoW() bool {
	panic("implement me")
}

func (ws *WalletStore) GeneratePowSync() error {
	panic("implement me")
}

func (ws *WalletStore) GeneratePoWAsync() error {
	panic("implement me")
}

func (ws *WalletStore) GetBalances() []*types.AccountMeta {
	panic("implement me")
}

func (ws *WalletStore) GetBalance(addr types.Address) types.AccountMeta {
	panic("implement me")
}

func (ws *WalletStore) SearchPending() {
	panic("implement me")
}

func (ws *WalletStore) Open(source types.Hash, token hash.Hash, representative types.Address) (*types.Block, error) {
	panic("implement me")
}

func (ws *WalletStore) Send(source types.Address, token types.Hash, to types.Address, amount types.Amount) (*types.Block, error) {
	panic("implement me")
}

func (ws *WalletStore) Receive(account types.Address, token types.Hash) (*types.Block, error) {
	panic("implement me")
}

func (ws *WalletStore) Change(addr types.Address, representative types.Address) (*types.Block, error) {
	panic("implement me")
}

func (ws *WalletStore) Erase() error {
	return ws.store.BadgerDb().DropAll()
}

func (ws *WalletStore) Import(content string, password string) error {
	panic("implement me")
}

func (ws *WalletStore) Export(path string) error {
	panic("implement me")
}

func (ws *WalletStore) GetWalletId() (types.Hash, error) {
	panic("implement me")
}

func (ws *WalletStore) GetVersion() (int64, error) {
	var i int64
	err := ws.store.View(func(txn db.StoreTxn) error {

		key := []byte{idPrefixIndex}
		return txn.Get(key, func(val []byte, b byte) error {
			i, _ = binary.Varint(val)
			return nil
		})
	})

	return i, err
}

func (ws *WalletStore) SetVersion(version int64) error {
	return ws.store.Update(func(txn db.StoreTxn) error {
		key := []byte{idPrefixIndex}
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buf, version)
		return txn.Set(key, buf[:n])
	})
}

func (ws *WalletStore) GetSalt() ([]byte, error) {
	panic("implement me")
}

func (ws *WalletStore) SetSalt(salt []byte) error {
	panic("implement me")
}

func (ws *WalletStore) GetDeterministicIndex() (int64, error) {
	panic("implement me")
}

func (ws *WalletStore) ResetEDeterministicIndex() error {
	panic("implement me")
}

func (ws *WalletStore) SetDeterministicIndex(index int64) error {
	panic("implement me")
}

func (ws *WalletStore) InsertDeterministic() error {
	panic("implement me")
}

func (ws *WalletStore) GetWork() (types.Work, error) {
	panic("implement me")
}

func (ws *WalletStore) SetWork(work types.Work) error {
	panic("implement me")
}

func (ws *WalletStore) IsAddressExist(addr types.Address) bool {
	panic("implement me")
}

func (ws *WalletStore) GetAddress(addr types.Address) (types.Account, error) {
	panic("implement me")
}

func (ws *WalletStore) AddAddress(account types.Account) error {
	panic("implement me")
}

func (ws *WalletStore) GetAddresses() ([]types.Address, error) {
	panic("implement me")
}

func (ws *WalletStore) ValidPassword() bool {
	panic("implement me")
}

func (ws *WalletStore) AttemptPassword(password string) error {
	panic("implement me")
}

func (ws *WalletStore) ChangePassword(password string) error {
	panic("implement me")
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

func (ws *WalletStore) Upgrade(migrations []*db.Migration) error {
	return nil
}
