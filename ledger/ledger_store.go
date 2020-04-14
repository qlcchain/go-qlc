/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package ledger

import (
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

type Store interface {
	AccountStore
	SmartBlockStore
	BlockStore
	PendingStore
	FrontierStore
	RepresentationStore
	UncheckedBlockStore
	PeerInfoStore
	SyncStore
	DposStore
	PovStore
	VoteStore
	Relation
	CacheStore
	LedgerStore
	PrivacyStore
	contractValueStore
	vmlogsStore
	vmStore
}

type ContractStore interface {
	GetBlockChild(hash types.Hash, c ...storage.Cache) (types.Hash, error)
	GetStateBlock(hash types.Hash, c ...storage.Cache) (*types.StateBlock, error)
	GetStateBlockConfirmed(hash types.Hash, c ...storage.Cache) (*types.StateBlock, error)
	GetLatestPovBlock() (*types.PovBlock, error)

	GetAccountMeta(address types.Address, c ...storage.Cache) (*types.AccountMeta, error)
	GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error)
	HasAccountMetaConfirmed(address types.Address) (bool, error)

	GetAccountMetaByPovHeight(address types.Address, height uint64) (*types.AccountMeta, error)
	GetTokenMetaByPovHeight(address types.Address, token types.Hash, height uint64) (*types.TokenMeta, error)
	GetTokenMetaByBlockHash(hash types.Hash) (*types.TokenMeta, error)

	GetContractValue(key *types.ContractKey, c ...storage.Cache) (*types.ContractValue, error)
	//IteratorContractStorage(prefix []byte, callback func(key *types.ContractKey, value *types.ContractValue) error) error

	GetLatestPovHeader() (*types.PovHeader, error)
	GetPovMinerStat(dayIndex uint32, batch ...storage.Batch) (*types.PovMinerDayStat, error)

	IsUserAccount(address types.Address) (bool, error)
	CalculateAmount(block *types.StateBlock) (types.Balance, error)
	GetRelation(dest interface{}, query string) error
	SelectRelation(dest interface{}, query string) error
}
