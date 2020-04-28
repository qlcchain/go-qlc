/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package ledger

import (
	"github.com/qlcchain/go-qlc/common/event"
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
	VmlogsStore
	VmStore
}

type ContractStore interface {
	GetBlockChild(hash types.Hash, c ...storage.Cache) (types.Hash, error)
	GetStateBlock(hash types.Hash, c ...storage.Cache) (*types.StateBlock, error)
	GetStateBlockConfirmed(hash types.Hash, c ...storage.Cache) (*types.StateBlock, error)
	ContractAddress(b *types.StateBlock) (*types.Address, error)

	GetAccountMeta(address types.Address, c ...storage.Cache) (*types.AccountMeta, error)
	GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error)
	HasAccountMetaConfirmed(address types.Address) (bool, error)

	GetAccountMetaByPovHeight(address types.Address, height uint64) (*types.AccountMeta, error)
	GetTokenMetaByPovHeight(address types.Address, token types.Hash, height uint64) (*types.TokenMeta, error)
	GetTokenMetaByBlockHash(hash types.Hash) (*types.TokenMeta, error)

	GetLatestPovHeader() (*types.PovHeader, error)
	GetLatestPovBlock() (*types.PovBlock, error)
	GetPovHeaderByHeight(height uint64) (*types.PovHeader, error)
	GetPovMinerStat(dayIndex uint32, batch ...storage.Batch) (*types.PovMinerDayStat, error)

	IsUserAccount(address types.Address) (bool, error)
	CalculateAmount(block *types.StateBlock) (types.Balance, error)
	GetRelation(dest interface{}, query string) error
	SelectRelation(dest interface{}, query string) error

	NewVMIterator(address *types.Address) *Iterator
	ListTokens() ([]*types.TokenInfo, error)
	GetTokenById(tokenId types.Hash) (*types.TokenInfo, error)
	GetTokenByName(tokenName string) (*types.TokenInfo, error)

	DBStore() storage.Store
	EventBus() event.EventBus
	Get(k []byte, c ...storage.Cache) ([]byte, error)
	GetObject(k []byte, c ...storage.Cache) (interface{}, []byte, error)
	Iterator([]byte, []byte, func([]byte, []byte) error) error
	IteratorObject(prefix []byte, end []byte, fn func([]byte, interface{}) error) error
}
