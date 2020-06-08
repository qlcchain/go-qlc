/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"go.uber.org/zap"
)

type ContractStore interface {
	GetStorage(prefix, key []byte) ([]byte, error)
	SetStorage(prefix, key []byte, value []byte) error
	//RemoveStorage(prefix, key []byte) error

	SetObjectStorage(prefix, key []byte, value interface{}) error
	GetStorageByRaw([]byte) ([]byte, error)

	Iterator(prefix []byte, fn func(key []byte, value []byte) error) error

	PovGlobalState() *statedb.PovGlobalStateDB
	PoVContractState() (*statedb.PovContractStateDB, error)
	PovGlobalStateByHeight(h uint64) *statedb.PovGlobalStateDB
	PoVContractStateByHeight(h uint64) (*statedb.PovContractStateDB, error)
	GetLatestPovBlock() (*types.PovBlock, error)
	GetLatestPovHeader() (*types.PovHeader, error)
	GetPovMinerStat(dayIndex uint32) (*types.PovMinerDayStat, error)

	ListTokens() ([]*types.TokenInfo, error)
	GetTokenById(tokenId types.Hash) (*types.TokenInfo, error)
	GetTokenByName(tokenName string) (*types.TokenInfo, error)

	EventBus() event.EventBus

	GetBlockChild(hash types.Hash) (types.Hash, error)
	GetStateBlock(hash types.Hash) (*types.StateBlock, error)
	GetStateBlockConfirmed(hash types.Hash) (*types.StateBlock, error)
	HasStateBlockConfirmed(hash types.Hash) (bool, error)

	GetAccountMeta(address types.Address) (*types.AccountMeta, error)
	GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error)
	HasAccountMetaConfirmed(address types.Address) (bool, error)

	GetAccountMetaByPovHeight(address types.Address) (*types.AccountMeta, error)
	GetTokenMetaByPovHeight(address types.Address, token types.Hash) (*types.TokenMeta, error)
	GetTokenMetaByBlockHash(hash types.Hash) (*types.TokenMeta, error)

	IsUserAccount(address types.Address) (bool, error)
	CalculateAmount(block *types.StateBlock) (types.Balance, error)
	GetRelation(dest interface{}, query string) error
	SelectRelation(dest interface{}, query string) error
}

var (
	//ErrStorageExists   = errors.New("storage already exists")
	ErrStorageNotFound = errors.New("storage not found")
)

func TrieHash(ctx *VMContext) *types.Hash {
	return ctx.cache.Trie(func(bytes []byte) []byte {
		return ctx.getRawStorageKey(bytes, nil)
	}).Hash()
}

func Trie(ctx *VMContext) *trie.Trie {
	return ctx.cache.Trie(func(bytes []byte) []byte {
		return ctx.getRawStorageKey(bytes, nil)
	})
}

func ToCache(ctx *VMContext) map[string]interface{} {
	result := make(map[string]interface{})
	for k, val := range ctx.cache.storage {
		rawKey := ctx.getRawStorageKey([]byte(k), nil)
		result[string(rawKey)] = val
	}
	return result
}

type VMContext struct {
	//TODO  should replace `ledger.Store` to `ledger.ContractStore`
	l            ledger.Store
	logger       *zap.SugaredLogger
	contractAddr *types.Address
	trie         *trie.Trie
	cache        *VMCache
	storageCache storage.Cache
	poVHeight    uint64
}

func NewVMContext(l ledger.Store, contractAddr *types.Address) *VMContext {
	h := uint64(0)
	if povHdr, err := l.GetLatestPovHeader(); err == nil {
		h = povHdr.GetHeight()
	}

	return &VMContext{
		l:            l,
		logger:       log.NewLogger("vm_context"),
		cache:        NewVMCache(),
		trie:         trie.NewTrie(l.DBStore(), nil, trie.NewSimpleTrieNodePool()),
		contractAddr: contractAddr,
		poVHeight:    h,
	}
}

func (v *VMContext) WithCache(c storage.Cache) *VMContext {
	v.storageCache = c
	return v
}

// WithBlock Load storage trie from the specified user address and block hash
func NewVMContextWithBlock(l ledger.Store, block *types.StateBlock) *VMContext {
	vmlog := log.NewLogger("vm_context")
	latestPovHdr, err := l.GetLatestPovHeader()
	if err == nil {
		vmlog.Warnf("latest height %d, block height %d ", latestPovHdr.GetHeight(), block.PoVHeight)
	}
	povHeight := block.PoVHeight
	if block.PoVHeight > latestPovHdr.GetHeight() {
		povHeight = latestPovHdr.GetHeight()
	}

	povHdr, err := l.GetPovHeaderByHeight(povHeight)
	if err != nil {
		vmlog.Error(err)
		return nil
	}

	vm := &VMContext{
		l:         l,
		logger:    vmlog,
		cache:     NewVMCache(),
		trie:      trie.NewTrie(l.DBStore(), nil, trie.NewSimpleTrieNodePool()),
		poVHeight: povHdr.GetHeight(),
	}
	if contractAddr, err := l.ContractAddress(block); err != nil {
		vm.logger.Error(err)
		return nil
	} else {
		vm.contractAddr = contractAddr
		return vm
	}
}

func (v *VMContext) GetBlockChild(hash types.Hash) (types.Hash, error) {
	return v.l.GetBlockChild(hash, v.storageCache)
}

func (v *VMContext) GetStateBlock(hash types.Hash) (*types.StateBlock, error) {
	return v.l.GetStateBlock(hash, v.storageCache)
}

func (v *VMContext) GetStateBlockConfirmed(hash types.Hash) (*types.StateBlock, error) {
	return v.l.GetStateBlockConfirmed(hash, v.storageCache)
}

func (v *VMContext) HasStateBlockConfirmed(hash types.Hash) (bool, error) {
	return v.l.HasStateBlockConfirmed(hash)
}

func (v *VMContext) GetLatestPovBlock() (*types.PovBlock, error) {
	return v.l.GetLatestPovBlock()
}

func (v *VMContext) GetAccountMeta(address types.Address) (*types.AccountMeta, error) {
	return v.l.GetAccountMeta(address, v.storageCache)
}

func (v *VMContext) GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error) {
	return v.l.GetTokenMeta(address, tokenType)
}

func (v *VMContext) HasAccountMetaConfirmed(address types.Address) (bool, error) {
	return v.l.HasAccountMetaConfirmed(address)
}

func (v *VMContext) GetAccountMetaByPovHeight(address types.Address) (*types.AccountMeta, error) {
	return v.l.GetAccountMetaByPovHeight(address, v.poVHeight)
}

func (v *VMContext) GetTokenMetaByPovHeight(address types.Address, token types.Hash) (*types.TokenMeta, error) {
	return v.l.GetTokenMetaByPovHeight(address, token, v.poVHeight)
}

func (v *VMContext) GetTokenMetaByBlockHash(hash types.Hash) (*types.TokenMeta, error) {
	return v.l.GetTokenMetaByBlockHash(hash)
}

func (v *VMContext) CalculateAmount(block *types.StateBlock) (types.Balance, error) {
	return v.l.CalculateAmount(block)
}

func (v *VMContext) GetRelation(dest interface{}, query string) error {
	return v.l.GetRelation(dest, query)
}

func (v *VMContext) SelectRelation(dest interface{}, query string) error {
	return v.l.SelectRelation(dest, query)
}

func (v *VMContext) SetObjectStorage(prefix, key []byte, value interface{}) error {
	storageKey := v.getStorageKey(prefix, key)

	v.cache.SetStorage(storageKey, value)

	return nil
}

func (v *VMContext) GetStorageByRaw(i []byte) ([]byte, error) {
	_, val, err := v.l.GetObject(i)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}
	return val, nil
}

func (v *VMContext) GetLatestPovHeader() (*types.PovHeader, error) {
	return v.l.GetLatestPovHeader()
}

func (v *VMContext) IsUserAccount(address types.Address) (bool, error) {
	if b, err := v.l.HasAccountMetaConfirmed(address); b {
		return true, nil
	} else {
		if err != nil {
			return false, err
		} else {
			return false, fmt.Errorf("can not find user account %s", address)
		}
	}
}

func (v *VMContext) GetStorage(prefix, key []byte) ([]byte, error) {
	storageKey := v.getStorageKey(prefix, key)
	if s, ok := v.cache.GetStorage(storageKey); !ok {
		if val, err := v.get(storageKey); err == nil {
			return val, nil
		} else {
			return nil, err
		}
	} else {
		if s != nil {
			switch o := s.(type) {
			case []byte:
				return o, nil
			default:
				return nil, fmt.Errorf("invalid type, %s", reflect.TypeOf(o))
			}
		}
	}

	return nil, nil
}

//FIXME: pov height
func (v *VMContext) GetPovMinerStat(dayIndex uint32) (*types.PovMinerDayStat, error) {
	return v.l.GetPovMinerStat(dayIndex)
}

func (v *VMContext) Iterator(prefix []byte, fn func(key []byte, value []byte) error) error {
	iterator := v.l.NewVMIterator(v.contractAddr)
	return iterator.Next(prefix, fn)
}

func (v *VMContext) SetStorage(prefix, key []byte, value []byte) error {
	storageKey := v.getStorageKey(prefix, key)

	v.cache.SetStorage(storageKey, value)

	return nil
}

func (v *VMContext) get(key []byte) ([]byte, error) {
	rawKey := v.getRawStorageKey(key, nil)
	if val, err := v.l.Get(rawKey); err == nil {
		return val, err
	} else if err == storage.KeyNotFound {
		return nil, ErrStorageNotFound
	} else {
		return nil, err
	}
}

func (v *VMContext) EventBus() event.EventBus {
	return v.l.EventBus()
}

func (v *VMContext) PovGlobalState() *statedb.PovGlobalStateDB {
	povHdr, err := v.l.GetLatestPovHeader()
	if err != nil {
		return nil
	}

	return statedb.NewPovGlobalStateDB(v.l.DBStore(), povHdr.GetStateHash())
}

func (v *VMContext) PoVContractState() (*statedb.PovContractStateDB, error) {
	state := v.PovGlobalState()
	if state != nil {
		return state.LookupContractStateDB(*v.contractAddr)
	} else {
		return nil, errors.New("can not get pov global state")
	}
}

func (v *VMContext) PovGlobalStateByHeight(h uint64) *statedb.PovGlobalStateDB {
	povHdr, err := v.l.GetPovHeaderByHeight(h)
	if err != nil {
		return nil
	}

	return statedb.NewPovGlobalStateDB(v.l.DBStore(), povHdr.GetStateHash())
}

func (v *VMContext) PoVContractStateByHeight(h uint64) (*statedb.PovContractStateDB, error) {
	status := v.PovGlobalStateByHeight(h)
	if status == nil {
		return nil, errors.New("can not get pov header")
	}
	return status.LookupContractStateDB(*v.contractAddr)
}

func (v *VMContext) ListTokens() ([]*types.TokenInfo, error) {
	return v.l.ListTokens()
}

func (v *VMContext) GetTokenById(tokenId types.Hash) (*types.TokenInfo, error) {
	return v.l.GetTokenById(tokenId)
}

func (v *VMContext) GetTokenByName(tokenName string) (*types.TokenInfo, error) {
	return v.l.GetTokenByName(tokenName)
}

func (v *VMContext) getStorageKey(prefix, key []byte) []byte {
	var storageKey []byte
	storageKey = append(storageKey, prefix...)
	storageKey = append(storageKey, key...)
	return storageKey
}

func (v *VMContext) getRawStorageKey(prefix, key []byte) []byte {
	var storageKey []byte
	storageKey = append(storageKey, []byte{byte(storage.KeyPrefixVMStorage)}...)
	if v.contractAddr != nil {
		storageKey = append(storageKey, v.contractAddr[:]...)
	}
	storageKey = append(storageKey, prefix...)
	storageKey = append(storageKey, key...)
	return storageKey
}
