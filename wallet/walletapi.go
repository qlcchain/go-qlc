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

type walletStoreApi interface {
	GetWalletId() ([]byte, error)
	GetVersion() (int64, error)
	SetVersion(version int64) error
	GetSeed() ([]byte, error)
	SetSeed(seed []byte) error
	GetDeterministicIndex() (int64, error)
	ResetDeterministicIndex() error
	SetDeterministicIndex(index int64) error
	GetWork() (types.Work, error)
	SetWork(work types.Work) error
	GetRepresentative() (types.Address, error)
	SetRepresentative(address types.Address) error
	IsAccountExist(addr types.Address) bool
	AttemptPassword(password string) error
	ChangePassword(password string) error
	EnterPassword(password string) error

	Init() error
	Remove() error
}

type walletAction interface {
	//HasPoW() bool
	//WaitPoW()
	//WaitingForPoW() bool
	//GeneratePowSync() error
	//GeneratePoWAsync() error
	Import(content string, password string) error
	Export(path string) error

	GetBalances() ([]*types.AccountMeta, error)
	GetBalance(addr types.Address) (*types.AccountMeta, error)
	SearchPending()
	Open(source types.Hash, token hash.Hash, representative types.Address) (*types.Block, error)
	Send(source types.Address, token types.Hash, to types.Address, amount types.Amount) (*types.Block, error)
	Receive(account types.Address, token types.Hash) (*types.Block, error)
	Change(addr types.Address, representative types.Address) (*types.Block, error)
}

type Walleter interface {
	walletAction
	walletStoreApi
}
