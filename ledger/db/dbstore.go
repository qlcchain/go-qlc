package db

import (
	"errors"
	"io"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	ErrStoreEmpty    = errors.New("the store is empty")
	ErrBlockExists   = errors.New("block already exists")
	ErrBlockNotFound = errors.New("block not found")
	ErrAccountExists = errors.New("account already exists")
	ErrTokenExists   = errors.New("token already exists")
	ErrTokenNotFound = errors.New("token not found")
	ErrPendingExists = errors.New("pending transaction already exists")
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
	AddAccountMeta(meta *types.AccountMeta) error
	GetAccountMeta(address types.Address) (*types.AccountMeta, error)
	UpdateAccountMeta(meta *types.AccountMeta) error
	DeleteAccountMeta(address types.Address) error
	HasAccountMeta(address types.Address) (bool, error)
	// token meta CURD
	AddTokenMeta(address types.Address, meta *types.TokenMeta) error
	GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error)
	DelTokenMeta(address types.Address, meta *types.TokenMeta) error
	// blocks CURD
	AddBlock(blk types.Block) error
	GetBlock(hash types.Hash) (types.Block, error)
	DeleteBlock(hash types.Hash) error
	HasBlock(hash types.Hash) (bool, error)
	CountBlocks() (uint64, error)
	GetRandomBlock() (types.Block, error)

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

	// pending CURD
	AddPending(destination types.Address, hash types.Hash, pending *types.PendingInfo) error
	GetPending(destination types.Address, hash types.Hash) (*types.PendingInfo, error)
	DeletePending(destination types.Address, hash types.Hash) error
}
