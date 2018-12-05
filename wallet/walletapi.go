/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"hash"

	"github.com/qlcchain/go-qlc/common/types"
)

type WalletStoreApi interface {
	Import(content string, password string) error
	Export(path string) error

	GetWalletId() (types.Hash, error)
	GetVersion() (int64, error)
	SetVersion(version int64) error
	GetSalt() ([]byte, error)
	SetSalt(salt []byte) error
	GetDeterministicIndex() (int64, error)
	ResetEDeterministicIndex() error
	SetDeterministicIndex(index int64) error
	InsertDeterministic() error
	GetWork() (types.Work, error)
	SetWork(work types.Work) error
	IsAddressExist(addr types.Address) bool
	GetAddress(addr types.Address) (types.Account, error)
	AddAddress(account types.Account) error
	GetAddresses() ([]types.Address, error)
	ValidPassword() bool
	AttemptPassword(password string) error
	ChangePassword(password string) error
}

type WalletAction interface {
	HasPoW() bool
	WaitPoW()
	WaitingForPoW() bool
	GeneratePowSync() error
	GeneratePoWAsync() error
	GetBalances() []*types.AccountMeta
	GetBalance(addr types.Address) types.AccountMeta
	SearchPending()
	Open(source types.Hash, token hash.Hash, representative types.Address) (*types.Block, error)
	Send(source types.Address, token types.Hash, to types.Address, amount types.Amount) (*types.Block, error)
	Receive(account types.Address, token types.Hash) (*types.Block, error)
	Change(addr types.Address, representative types.Address) (*types.Block, error)
}
