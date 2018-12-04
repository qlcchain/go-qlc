package db

import (
	"io"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
)

// Store is an interface that all stores need to implement.
type Store interface {
	io.Closer
	Purge() error
	View(fn func(txn StoreTxn) error) error
	Update(fn func(txn StoreTxn) error) error
	BadgerDb() *badger.DB
	NewTransaction(update bool) *BadgerStoreTxn
}

type StoreTxn interface {
	Set(key, val []byte) error
	SetWithMeta(key, val []byte, meta byte) error
	Get(key []byte, fn func([]byte, byte) error) error
	Delete(key []byte) error
	Iterator(pre byte, fn func([]byte, []byte, byte) error) error
	Commit(callback func(error)) error
	Discard()
}

type LedgerStore interface {
	Empty(txn StoreTxn) (bool, error)
	Flush(txn StoreTxn) error

	// blocks CURD
	AddBlock(blk types.Block, txn StoreTxn) error
	GetBlock(hash types.Hash, txn StoreTxn) (types.Block, error)
	GetBlocks(txn StoreTxn) ([]types.Block, error)
	DeleteBlock(hash types.Hash, txn StoreTxn) error
	HasBlock(hash types.Hash, txn StoreTxn) (bool, error)
	CountBlocks(txn StoreTxn) (uint64, error)
	GetRandomBlock(txn StoreTxn) (types.Block, error)

	// unchecked CURD
	AddUncheckedBlock(parentHash types.Hash, blk types.Block, kind types.UncheckedKind, txn StoreTxn) error
	GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txn StoreTxn) (types.Block, error)
	DeleteUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txn StoreTxn) error
	HasUncheckedBlock(hash types.Hash, kind types.UncheckedKind, txn StoreTxn) (bool, error)
	WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc, txn StoreTxn) error
	CountUncheckedBlocks(txn StoreTxn) (uint64, error)

	// account meta CURD
	AddAccountMeta(meta *types.AccountMeta, txn StoreTxn) error
	UpdateAccountMeta(meta *types.AccountMeta, txn StoreTxn) error
	AddOrUpdateAccountMeta(meta *types.AccountMeta, txn StoreTxn) error
	GetAccountMeta(address types.Address, txn StoreTxn) (*types.AccountMeta, error)
	HasAccountMeta(address types.Address, txn StoreTxn) (bool, error)
	DeleteAccountMeta(address types.Address, txn StoreTxn) error

	// token meta CURD
	AddTokenMeta(address types.Address, meta *types.TokenMeta, txn StoreTxn) error
	UpdateTokenMeta(address types.Address, meta *types.TokenMeta, txn StoreTxn) error
	AddOrUpdateTokenMeta(address types.Address, meta *types.TokenMeta, txn StoreTxn) error
	GetTokenMeta(address types.Address, tokenType types.Hash, txn StoreTxn) (*types.TokenMeta, error)
	HasTokenMeta(address types.Address, tokenType types.Hash, txn StoreTxn) (bool, error)
	DeleteTokenMeta(address types.Address, tokenType types.Hash, txn StoreTxn) error

	// representation CURD
	AddRepresentationWeight(address types.Address, amount types.Balance, txn StoreTxn) error
	SubRepresentationWeight(address types.Address, amount types.Balance, txn StoreTxn) error
	GetRepresentation(address types.Address, txn StoreTxn) (types.Balance, error)

	// pending CURD
	AddPending(destination types.Address, hash types.Hash, pending *types.PendingInfo, txn StoreTxn) error
	GetPending(destination types.Address, hash types.Hash, txn StoreTxn) (*types.PendingInfo, error)
	DeletePending(destination types.Address, hash types.Hash, txn StoreTxn) error

	// frontier CURD
	AddFrontier(frontier *types.Frontier, txn StoreTxn) error
	GetFrontier(hash types.Hash, txn StoreTxn) (*types.Frontier, error)
	GetFrontiers(txn StoreTxn) ([]*types.Frontier, error)
	DeleteFrontier(hash types.Hash, txn StoreTxn) error
	CountFrontiers(txn StoreTxn) (uint64, error)
}
