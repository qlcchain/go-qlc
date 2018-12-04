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
}

type StoreTxn interface {
	Set(key, val []byte) error
	SetWithMeta(key, val []byte, meta byte) error
	Get(key []byte, fn func([]byte, byte) error) error
	Delete(key []byte) error
	Iterator(pre byte, fn func([]byte, []byte, byte) error) error
}

type LedgerStore interface {
	Empty() (bool, error)
	Flush() error

	// blocks CURD
	AddBlock(blk types.Block) error
	GetBlock(hash types.Hash) (types.Block, error)
	GetBlocks() ([]types.Block, error)
	DeleteBlock(hash types.Hash) error
	HasBlock(hash types.Hash) (bool, error)
	CountBlocks() (uint64, error)
	GetRandomBlock() (types.Block, error)

	// unchecked CURD
	AddUncheckedBlock(parentHash types.Hash, blk types.Block, kind types.UncheckedKind) error
	GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind) (types.Block, error)
	DeleteUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind) error
	HasUncheckedBlock(hash types.Hash, kind types.UncheckedKind) (bool, error)
	WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc) error
	CountUncheckedBlocks() (uint64, error)

	// account meta CURD
	AddAccountMeta(meta *types.AccountMeta) error
	UpdateAccountMeta(meta *types.AccountMeta) error
	AddOrUpdateAccountMeta(meta *types.AccountMeta) error
	GetAccountMeta(address types.Address) (*types.AccountMeta, error)
	HasAccountMeta(address types.Address) (bool, error)
	DeleteAccountMeta(address types.Address) error

	// token meta CURD
	AddTokenMeta(address types.Address, meta *types.TokenMeta) error
	UpdateTokenMeta(address types.Address, meta *types.TokenMeta) error
	AddOrUpdateTokenMeta(address types.Address, meta *types.TokenMeta) error
	GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error)
	HasTokenMeta(address types.Address, tokenType types.Hash) (bool, error)
	DeleteTokenMeta(address types.Address, tokenType types.Hash) error

	// representation CURD
	AddRepresentationWeight(address types.Address, amount types.Balance) error
	SubRepresentationWeight(address types.Address, amount types.Balance) error
	GetRepresentation(address types.Address) (types.Balance, error)

	// pending CURD
	AddPending(destination types.Address, hash types.Hash, pending *types.PendingInfo) error
	GetPending(destination types.Address, hash types.Hash) (*types.PendingInfo, error)
	DeletePending(destination types.Address, hash types.Hash) error

	// frontier CURD
	AddFrontier(frontier *types.Frontier) error
	GetFrontier(hash types.Hash) (*types.Frontier, error)
	GetFrontiers() ([]*types.Frontier, error)
	DeleteFrontier(hash types.Hash) error
	CountFrontiers() (uint64, error)
}
