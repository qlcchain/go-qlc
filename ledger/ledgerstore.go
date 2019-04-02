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
	smsStore
	Empty(txns ...db.StoreTxn) (bool, error)
	BatchUpdate(fn func(txn db.StoreTxn) error) error

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
	AddRepresentation(address types.Address, amount types.Balance, txns ...db.StoreTxn) error
	SubRepresentation(address types.Address, amount types.Balance, txns ...db.StoreTxn) error
	GetRepresentation(address types.Address, txns ...db.StoreTxn) (types.Balance, error)
	GetRepresentations(fn func(types.Address, types.Balance) error, txns ...db.StoreTxn) error
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
	AddPending(pendingKey types.PendingKey, pending *types.PendingInfo, txns ...db.StoreTxn) error
	GetPending(pendingKey types.PendingKey, txns ...db.StoreTxn) (*types.PendingInfo, error)
	DeletePending(pendingKey types.PendingKey, txns ...db.StoreTxn) error
	SearchPending(address types.Address, fn func(key *types.PendingKey, value *types.PendingInfo) error, txns ...db.StoreTxn) error
	// frontier CURD
	AddFrontier(frontier *types.Frontier, txns ...db.StoreTxn) error
	GetFrontier(hash types.Hash, txns ...db.StoreTxn) (*types.Frontier, error)
	GetFrontiers(txns ...db.StoreTxn) ([]*types.Frontier, error)
	DeleteFrontier(hash types.Hash, txns ...db.StoreTxn) error
	CountFrontiers(txns ...db.StoreTxn) (uint64, error)
	// posterior
	GetChild(hash types.Hash, address types.Address, txns ...db.StoreTxn) (types.Hash, error)
	// performance
	AddOrUpdatePerformance(p *types.PerformanceTime, txns ...db.StoreTxn) error
	PerformanceTimes(fn func(*types.PerformanceTime), txns ...db.StoreTxn) error
	GetPerformanceTime(hash types.Hash, txns ...db.StoreTxn) (*types.PerformanceTime, error)
	IsPerformanceTimeExist(hash types.Hash, txns ...db.StoreTxn) (bool, error)

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
}
