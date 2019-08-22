/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/ledger/db"
)

type Store interface {
	Empty(txns ...db.StoreTxn) (bool, error)
	BatchUpdate(fn func(txn db.StoreTxn) error) error
	BatchView(fn func(txn db.StoreTxn) error) error
	DBStore() db.Store

	// account meta CURD
	AddAccountMeta(meta *types.AccountMeta, txns ...db.StoreTxn) error
	GetAccountMeta(address types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error)
	GetAccountMetas(fn func(am *types.AccountMeta) error, txns ...db.StoreTxn) error
	CountAccountMetas(txns ...db.StoreTxn) (uint64, error)
	UpdateAccountMeta(meta *types.AccountMeta, txns ...db.StoreTxn) error
	DeleteAccountMeta(address types.Address, txns ...db.StoreTxn) error
	HasAccountMeta(address types.Address, txns ...db.StoreTxn) (bool, error)
	// token meta CURD
	AddTokenMeta(address types.Address, meta *types.TokenMeta, txns ...db.StoreTxn) error
	GetTokenMeta(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) (*types.TokenMeta, error)
	UpdateTokenMeta(address types.Address, meta *types.TokenMeta, txns ...db.StoreTxn) error
	DeleteTokenMeta(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) error
	// state block CURD
	AddStateBlock(blk *types.StateBlock, txns ...db.StoreTxn) error
	GetStateBlock(hash types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error)
	GetStateBlocks(fn func(*types.StateBlock) error, txns ...db.StoreTxn) error
	DeleteStateBlock(hash types.Hash, txns ...db.StoreTxn) error
	HasStateBlock(hash types.Hash, txns ...db.StoreTxn) (bool, error)
	CountStateBlocks(txns ...db.StoreTxn) (uint64, error)
	GetRandomStateBlock(txns ...db.StoreTxn) (*types.StateBlock, error)
	// smart contract block CURD
	AddSmartContractBlock(blk *types.SmartContractBlock, txns ...db.StoreTxn) error
	GetSmartContractBlock(hash types.Hash, txns ...db.StoreTxn) (*types.SmartContractBlock, error)
	HasSmartContractBlock(hash types.Hash, txns ...db.StoreTxn) (bool, error)
	GetSmartContractBlocks(fn func(block *types.SmartContractBlock) error, txns ...db.StoreTxn) error
	CountSmartContractBlocks(txns ...db.StoreTxn) (uint64, error)
	// representation CURD
	AddRepresentation(address types.Address, benefit *types.Benefit, txns ...db.StoreTxn) error
	SubRepresentation(address types.Address, benefit *types.Benefit, txns ...db.StoreTxn) error
	GetRepresentation(address types.Address, txns ...db.StoreTxn) (*types.Benefit, error)
	GetRepresentations(fn func(types.Address, *types.Benefit) error, txns ...db.StoreTxn) error
	GetOnlineRepresentations(txns ...db.StoreTxn) ([]types.Address, error)
	SetOnlineRepresentations(addresses []*types.Address, txns ...db.StoreTxn) error
	// unchecked CURD
	AddUncheckedBlock(parentHash types.Hash, blk *types.StateBlock, kind types.UncheckedKind, sync types.SynchronizedKind, txns ...db.StoreTxn) error
	GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (*types.StateBlock, types.SynchronizedKind, error)
	DeleteUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) error
	HasUncheckedBlock(hash types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (bool, error)
	WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc, txns ...db.StoreTxn) error
	CountUncheckedBlocks(txns ...db.StoreTxn) (uint64, error)
	// pending CURD
	AddPending(pendingKey *types.PendingKey, pending *types.PendingInfo, txns ...db.StoreTxn) error
	GetPending(pendingKey *types.PendingKey, txns ...db.StoreTxn) (*types.PendingInfo, error)
	GetPendings(fn func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error, txns ...db.StoreTxn) error
	DeletePending(pendingKey *types.PendingKey, txns ...db.StoreTxn) error
	SearchPending(address types.Address, fn func(key *types.PendingKey, value *types.PendingInfo) error, txns ...db.StoreTxn) error
	// frontier CURD
	AddFrontier(frontier *types.Frontier, txns ...db.StoreTxn) error
	GetFrontier(hash types.Hash, txns ...db.StoreTxn) (*types.Frontier, error)
	GetFrontiers(txns ...db.StoreTxn) ([]*types.Frontier, error)
	DeleteFrontier(hash types.Hash, txns ...db.StoreTxn) error
	CountFrontiers(txns ...db.StoreTxn) (uint64, error)
	// posterior
	GetChild(hash types.Hash, txns ...db.StoreTxn) (types.Hash, error)
	// performance
	AddOrUpdatePerformance(p *types.PerformanceTime, txns ...db.StoreTxn) error
	PerformanceTimes(fn func(*types.PerformanceTime), txns ...db.StoreTxn) error
	GetPerformanceTime(hash types.Hash, txns ...db.StoreTxn) (*types.PerformanceTime, error)
	IsPerformanceTimeExist(hash types.Hash, txns ...db.StoreTxn) (bool, error)
	GetLinkBlock(hash types.Hash, txns ...db.StoreTxn) (types.Hash, error)

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

	//GenerateBlock
	GenerateSendBlock(block *types.StateBlock, amount types.Balance, prk ed25519.PrivateKey) (*types.StateBlock, error)
	GenerateReceiveBlock(sendBlock *types.StateBlock, prk ed25519.PrivateKey) (*types.StateBlock, error)
	GenerateChangeBlock(account types.Address, representative types.Address, prk ed25519.PrivateKey) (*types.StateBlock, error)

	//Token
	//ListTokens(txns ...db.StoreTxn) ([]*types.TokenInfo, error)
	//GetTokenById(tokenId types.Hash, txns ...db.StoreTxn) (*types.TokenInfo, error)
	//GetTokenByName(tokenName string, txns ...db.StoreTxn) (*types.TokenInfo, error)
	GetGenesis(txns ...db.StoreTxn) ([]*types.StateBlock, error)

	//CalculateAmount calculate block amount by balance and check block type
	CalculateAmount(block *types.StateBlock, txns ...db.StoreTxn) (types.Balance, error)
	AddMessageInfo(mHash types.Hash, message []byte, txns ...db.StoreTxn) error
	GetMessageInfo(mHash types.Hash, txns ...db.StoreTxn) ([]byte, error)

	//POV blocks base CRUD
	AddPovBlock(blk *types.PovBlock, td *types.PovTD, txns ...db.StoreTxn) error
	DeletePovBlock(blk *types.PovBlock, txns ...db.StoreTxn) error
	AddPovHeader(header *types.PovHeader, txns ...db.StoreTxn) error
	DeletePovHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) error
	GetPovHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovHeader, error)
	HasPovHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) bool
	AddPovBody(height uint64, hash types.Hash, body *types.PovBody, txns ...db.StoreTxn) error
	DeletePovBody(height uint64, hash types.Hash, txns ...db.StoreTxn) error
	GetPovBody(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovBody, error)
	HasPovBody(height uint64, hash types.Hash, txns ...db.StoreTxn) bool
	AddPovHeight(hash types.Hash, height uint64, txns ...db.StoreTxn) error
	DeletePovHeight(hash types.Hash, txns ...db.StoreTxn) error
	GetPovHeight(hash types.Hash, txns ...db.StoreTxn) (uint64, error)
	HasPovHeight(hash types.Hash, txns ...db.StoreTxn) bool
	AddPovTD(hash types.Hash, height uint64, td *types.PovTD, txns ...db.StoreTxn) error
	DeletePovTD(hash types.Hash, height uint64, txns ...db.StoreTxn) error
	GetPovTD(hash types.Hash, height uint64, txns ...db.StoreTxn) (*types.PovTD, error)
	AddPovTxLookup(txHash types.Hash, txLookup *types.PovTxLookup, txns ...db.StoreTxn) error
	DeletePovTxLookup(txHash types.Hash, txns ...db.StoreTxn) error
	GetPovTxLookup(txHash types.Hash, txns ...db.StoreTxn) (*types.PovTxLookup, error)
	HasPovTxLookup(txHash types.Hash, txns ...db.StoreTxn) bool

	// POV best chain CURD
	AddPovBestHash(height uint64, hash types.Hash, txns ...db.StoreTxn) error
	DeletePovBestHash(height uint64, txns ...db.StoreTxn) error
	GetPovBestHash(height uint64, txns ...db.StoreTxn) (types.Hash, error)

	AddPovMinerStat(dayStat *types.PovMinerDayStat, txns ...db.StoreTxn) error
	DeletePovMinerStat(dayIndex uint32, txns ...db.StoreTxn) error
	GetPovMinerStat(dayIndex uint32, txns ...db.StoreTxn) (*types.PovMinerDayStat, error)
	GetLatestPovMinerStat(txns ...db.StoreTxn) (*types.PovMinerDayStat, error)

	// POV blocks complex queries
	GetPovBlockByHeightAndHash(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error)
	GetPovBlockByHeight(height uint64, txns ...db.StoreTxn) (*types.PovBlock, error)
	GetPovBlockByHash(hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error)
	BatchGetPovHeadersByHeightAsc(height uint64, count uint64, txns ...db.StoreTxn) ([]*types.PovHeader, error)
	BatchGetPovHeadersByHeightDesc(height uint64, count uint64, txns ...db.StoreTxn) ([]*types.PovHeader, error)
	GetPovHeaderByHash(hash types.Hash, txns ...db.StoreTxn) (*types.PovHeader, error)
	GetPovHeaderByHeight(height uint64, txns ...db.StoreTxn) (*types.PovHeader, error)
	GetAllPovHeaders(fn func(header *types.PovHeader) error, txns ...db.StoreTxn) error
	GetAllPovBlocks(fn func(*types.PovBlock) error, txns ...db.StoreTxn) error
	GetAllPovBestHeaders(fn func(header *types.PovHeader) error, txns ...db.StoreTxn) error
	GetAllPovBestBlocks(fn func(*types.PovBlock) error, txns ...db.StoreTxn) error
	GetLatestPovHeader(txns ...db.StoreTxn) (*types.PovHeader, error)
	GetLatestPovBlock(txns ...db.StoreTxn) (*types.PovBlock, error)
	HasPovBlock(height uint64, hash types.Hash, txns ...db.StoreTxn) bool

	DropAllPovBlocks() error
	CountPovBlocks(txns ...db.StoreTxn) (uint64, error)
	CountPovTxs(txns ...db.StoreTxn) (uint64, error)
	CountPovBestHashs(txns ...db.StoreTxn) (uint64, error)
}
