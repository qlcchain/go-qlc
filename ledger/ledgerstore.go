/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

type Store interface {
	Empty(txns ...db.StoreTxn) (bool, error)
	BatchUpdate(fn func(txn db.StoreTxn) error) error

	// account meta CURD
	AddAccountMeta(meta *types.AccountMeta, txns ...db.StoreTxn) error
	GetAccountMeta(address types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error)
	UpdateAccountMeta(meta *types.AccountMeta, txns ...db.StoreTxn) error
	DeleteAccountMeta(address types.Address, txns ...db.StoreTxn) error
	HasAccountMeta(address types.Address, txns ...db.StoreTxn) (bool, error)
	// token meta CURD
	AddTokenMeta(address types.Address, meta *types.TokenMeta, txns ...db.StoreTxn) error
	GetTokenMeta(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) (*types.TokenMeta, error)
	UpdateTokenMeta(address types.Address, meta *types.TokenMeta, txns ...db.StoreTxn) error
	DeleteTokenMeta(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) error
	// blocks CURD
	AddBlock(blk types.Block, txns ...db.StoreTxn) error
	GetBlock(hash types.Hash, txns ...db.StoreTxn) (types.Block, error)
	GetBlocks(txns ...db.StoreTxn) ([]*types.Block, error)
	DeleteBlock(hash types.Hash, txns ...db.StoreTxn) error
	HasBlock(hash types.Hash, txns ...db.StoreTxn) (bool, error)
	CountBlocks(txns ...db.StoreTxn) (uint64, error)
	GetRandomBlock(txns ...db.StoreTxn) (types.Block, error)
	// representation CURD
	AddRepresentation(address types.Address, amount types.Balance, txns ...db.StoreTxn) error
	SubRepresentation(address types.Address, amount types.Balance, txns ...db.StoreTxn) error
	GetRepresentation(address types.Address, txns ...db.StoreTxn) (types.Balance, error)
	// unchecked CURD
	AddUncheckedBlock(parentHash types.Hash, blk types.Block, kind types.UncheckedKind, txns ...db.StoreTxn) error
	GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (types.Block, error)
	DeleteUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) error
	HasUncheckedBlock(hash types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (bool, error)
	WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc, txns ...db.StoreTxn) error
	CountUncheckedBlocks(txns ...db.StoreTxn) (uint64, error)
	// pending CURD
	AddPending(pendingKey types.PendingKey, pending *types.PendingInfo, txns ...db.StoreTxn) error
	GetPending(pendingKey types.PendingKey, txns ...db.StoreTxn) (*types.PendingInfo, error)
	DeletePending(pendingKey types.PendingKey, txns ...db.StoreTxn) error
	// frontier CURD
	AddFrontier(frontier *types.Frontier, txns ...db.StoreTxn) error
	GetFrontier(hash types.Hash, txns ...db.StoreTxn) (*types.Frontier, error)
	GetFrontiers(txns ...db.StoreTxn) ([]*types.Frontier, error)
	DeleteFrontier(hash types.Hash, txns ...db.StoreTxn) error
	CountFrontiers(txns ...db.StoreTxn) (uint64, error)

	//Latest block hash by account and token type, if not exist, return zero hash
	Latest(account types.Address, token types.Hash, txns ...db.StoreTxn) types.Hash
	//Account get account meta by block hash
	Account(hash types.Hash, txns ...db.StoreTxn) (*types.AccountMeta, error)
	//Token get token meta by block hash
	Token(hash types.Hash, txns ...db.StoreTxn) (*types.TokenMeta, error)
	//Pending get account pending (token_hash->pending)
	Pending(account types.Address, txns ...db.StoreTxn) ([]*types.PendingKey, error)
	//Balance get account balance (token_hash->pending)
	Balance(account types.Address, txns ...db.StoreTxn) (map[types.Hash]types.Balance, error)
	//TokenPending get account token pending
	TokenPending(account types.Address, token types.Hash, txns ...db.StoreTxn) ([]*types.PendingKey, error)
	//TokenBalance get account token balance
	TokenBalance(account types.Address, token types.Hash, txns ...db.StoreTxn) (types.Balance, error)
	//Weight get vote weight for PoS
	Weight(account types.Address, txns ...db.StoreTxn) types.Balance
	//Rollback blocks until `hash' doesn't exist
	Rollback(hash types.Hash) error
}
