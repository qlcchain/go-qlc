package db

import (
	"io"

	"github.com/qlcchain/go-qlc/common/types"
)

// Store is an interface that all stores need to implement.
type Store interface {
	io.Closer
	Purge() error
	View(fn func(txn StoreTxn) error) error
	Update(fn func(txn StoreTxn) error) error
}

type StoreTxn interface {
	Empty() (bool, error)
	Flush() error

	// account meta CURD
	AddAccountMeta(account types.Address, meta *types.AccountMeta) error
	GetAccountMeta(address types.Address) (*types.AccountMeta, error)
	UpdateAccountMeta(address types.Address, meta *types.AccountMeta) error
	DeleteAccountMeta(address types.Address) error
	HasAccountMeta(address types.Address) (bool, error)
	// token meta CURD
	AddTokenMeta(account types.Address, meta *types.TokenMeta) error
	GetTokenMeta(account types.Address, meta *types.TokenMeta) error
	DelTokenMeta(account types.Address, meta *types.TokenMeta) error
	// blocks CURD
	AddBlock(blk types.Block) error
	GetBlock(hash types.Hash) (types.Block, error)
	DeleteBlock(hash types.Hash) error
	HasBlock(hash types.Hash) (bool, error)
	CountBlocks() (uint64, error)
	GetRandomBlock() (*types.Block, error)

	AddRepresentation(address types.Address, amount types.Balance) error
	SubRepresentation(address types.Address, amount types.Balance) error
	GetRepresentation(address types.Address) (types.Balance, error)

	// unchecked CURD
	AddUncheckedBlock(parentHash types.Hash, blk types.Block, kind types.UncheckedKind) error
	GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind) (types.Block, error)
	DeleteUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind) error
	HasUncheckedBlock(hash types.Hash, kind types.UncheckedKind) (bool, error)
	WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc) error
	CountUncheckedBlocks() (uint64, error)
	//
	//
	AddPending(destination types.Address, hash types.Hash, pending *types.PendingInfo) error
	GetPending(destination types.Address, hash types.Hash) (*types.PendingInfo, error)
	DeletePending(destination types.Address, hash types.Hash) error
}
