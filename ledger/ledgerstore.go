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
	"github.com/qlcchain/go-qlc/crypto/ed25519"
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
	Relation

	Close() error
	DBStore() storage.Store
	Cache() *MemoryCache
	EventBus() event.EventBus
	Get(k []byte, c ...storage.Cache) (interface{}, []byte, error)
	Iterator([]byte, []byte, func([]byte, []byte) error) error
	GenerateSendBlock(block *types.StateBlock, amount types.Balance, prk ed25519.PrivateKey) (*types.StateBlock, error)
	GenerateReceiveBlock(sendBlock *types.StateBlock, prk ed25519.PrivateKey) (*types.StateBlock, error)
	GenerateChangeBlock(account types.Address, representative types.Address, prk ed25519.PrivateKey) (*types.StateBlock, error)
	GenerateOnlineBlock(account types.Address, prk ed25519.PrivateKey, povHeight uint64) (*types.StateBlock, error)
	Action(at storage.ActionType, t int) (interface{}, error)
}
