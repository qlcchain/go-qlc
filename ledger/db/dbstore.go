package db

import (
	"io"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/qlcchain/go-qlc/ledger"
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
	AddAccountMeta(account types.Address, meta *ledger.AccountMeta) error
	GetAccountMeta(address types.Address) (*ledger.AccountMeta, error)
	UpdateAccountMeta(address types.Address, meta *ledger.AccountMeta) error
	DeleteAccountMeta(address types.Address) error
	HasAccountMeta(address types.Address) (bool, error)
	// token meta CURD
	AddTokenMeta(account types.Address, meta *ledger.TokenMeta) error
	GetTokenMeta(account types.Address, meta *ledger.TokenMeta) error
	DelTokenMeta(account types.Address, meta *ledger.TokenMeta) error
	// blocks CURD
	AddBlock(blk ledger.Block) error
	GetBlock(hash types.Hash) (ledger.Block, error)
	DeleteBlock(hash types.Hash) error
	HasBlock(hash types.Hash) (bool, error)
	CountBlocks() (uint64, error)
	GetRandomBlock() (*ledger.Block, error)

	AddRepresentation(address types.Address, amount types.Balance) error
	SubRepresentation(address types.Address, amount types.Balance) error
	GetRepresentation(address types.Address) (types.Balance, error)

	// unchecked CURD
	AddUncheckedBlock(parentHash types.Hash, blk ledger.Block, kind ledger.UncheckedKind) error
	GetUncheckedBlock(parentHash types.Hash, kind ledger.UncheckedKind) (ledger.Block, error)
	DeleteUncheckedBlock(parentHash types.Hash, kind ledger.UncheckedKind) error
	HasUncheckedBlock(hash types.Hash, kind ledger.UncheckedKind) (bool, error)
	WalkUncheckedBlocks(visit ledger.UncheckedBlockWalkFunc) error
	CountUncheckedBlocks() (uint64, error)
	//
	//
	AddPending(destination types.Address, hash types.Hash, pending *ledger.PendingInfo) error
	GetPending(destination types.Address, hash types.Hash) (*ledger.PendingInfo, error)
	DeletePending(destination types.Address, hash types.Hash) error
}
