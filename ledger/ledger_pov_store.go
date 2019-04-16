package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

type povStore interface {
	//POV blocks base CRUD
	AddPovBlock(blk *types.PovBlock, txns ...db.StoreTxn) error
	DeletePovBlock(blk *types.PovBlock, txns ...db.StoreTxn) error
	AddPovHeader(header *types.PovHeader, txns ...db.StoreTxn) error
	DeletePovHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) error
	GetPovHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovHeader, error)
	HasPovHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) bool
	AddPovBody(height uint64, hash types.Hash, body *types.PovBody, txns ...db.StoreTxn) error
	DeletePovBody(height uint64, hash types.Hash, txns ...db.StoreTxn) error
	GetPovBody(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovBody, error)
	AddPovHeight(hash types.Hash, height uint64, txns ...db.StoreTxn) error
	DeletePovHeight(hash types.Hash, txns ...db.StoreTxn) error
	GetPovHeight(hash types.Hash, txns ...db.StoreTxn) (uint64, error)
	AddPovTxLookup(txHash types.Hash, txLookup *types.PovTxLookup, txns... db.StoreTxn) error
	DeletePovTxLookup(txHash types.Hash, txns... db.StoreTxn) error

	// POV best chain CURD
	AddPovBestHash(height uint64, hash types.Hash, txns ...db.StoreTxn) error
	DeletePovBestHash(height uint64, txns ...db.StoreTxn) error
	GetPovBestHash(height uint64, txns ...db.StoreTxn) (types.Hash, error)

	// POV blocks complex queries
	GetPovBlockByHeightAndHash(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error)
	GetPovBlockByHeight(height uint64, txns ...db.StoreTxn) (*types.PovBlock, error)
	GetPovBlockByHash(hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error)
	GetAllPovBlocks(fn func(*types.PovBlock) error, txns ...db.StoreTxn) error
	GetLatestPovBlock(txns ...db.StoreTxn) (*types.PovBlock, error)
}