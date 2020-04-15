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

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
)

type ContractStore interface {
	ledger.ContractStore
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

	ListTokens() ([]*types.TokenInfo, error)
	GetTokenById(tokenId types.Hash) (*types.TokenInfo, error)
	GetTokenByName(tokenName string) (*types.TokenInfo, error)

	EventBus() event.EventBus
}

var (
	//ErrStorageExists   = errors.New("storage already exists")
	ErrStorageNotFound = errors.New("storage not found")
)

//func SaveStorage(ctx *VMContext, batch ...storage.Batch) error {
//	for k, val := range ctx.cache.storage {
//		rawKey := ctx.getRawStorageKey([]byte(k), nil)
//		err := ctx.set(rawKey, val, batch...)
//		if err != nil {
//			ctx.logger.Error(err)
//			return err
//		}
//	}
//	return nil
//}

func TrieHash(ctx *VMContext) *types.Hash {
	return ctx.cache.Trie().Hash()
}

func ToCache(ctx *VMContext) map[string]interface{} {
	result := make(map[string]interface{})
	for k, val := range ctx.cache.storage {
		rawKey := ctx.getRawStorageKey([]byte(k), nil)
		result[string(rawKey)] = val
	}
	return result
}

//func SaveTrie(ctx *VMContext, batch ...storage.Batch) error {
//	fn, err := ctx.cache.Trie().Save(batch...)
//	if err != nil {
//		return err
//	}
//	fn()
//	return nil
//}

//func SaveBlockData(ctx *VMContext, block *types.StateBlock, c ...storage.Cache) error {
//	if err := SaveTrie(ctx); err != nil {
//		return err
//	}
//	trieRoot := TrieHash(ctx)
//	return ctx.l.UpdateContractValueByBlock(block, trieRoot, c...)
//}
//
//func DeleteBlockData(ctx *VMContext, block *types.StateBlock, c ...storage.Cache) error {
//	return ctx.l.DeleteContractValueByBlock(block, c...)
//}

type VMContext struct {
	//TODO  should replace `ledger.Store` to `ledger.ContractStore`
	l            ledger.Store
	logger       *zap.SugaredLogger
	contractAddr *types.Address
	trie         *trie.Trie
	cache        *VMCache
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

// WithBlock Load storage trie from the specified user address and block hash
func NewVMContextWithBlock(l ledger.Store, block *types.StateBlock) *VMContext {
	povHdr, err := l.GetPovHeaderByHeight(block.PoVHeight)
	if err != nil {
		return nil
	}

	return &VMContext{
		l:            l,
		logger:       log.NewLogger("vm_context"),
		cache:        NewVMCache(),
		trie:         trie.NewTrie(l.DBStore(), nil, trie.NewSimpleTrieNodePool()),
		contractAddr: block.ContractAddress(),
		poVHeight:    povHdr.GetHeight(),
	}
}

func (v *VMContext) GetBlockChild(hash types.Hash, c ...storage.Cache) (types.Hash, error) {
	return v.l.GetBlockChild(hash, c...)
}

func (v *VMContext) GetStateBlock(hash types.Hash, c ...storage.Cache) (*types.StateBlock, error) {
	return v.l.GetStateBlock(hash, c...)
}

func (v *VMContext) GetStateBlockConfirmed(hash types.Hash, c ...storage.Cache) (*types.StateBlock, error) {
	return v.l.GetStateBlockConfirmed(hash, c...)
}

func (v *VMContext) GetLatestPovBlock() (*types.PovBlock, error) {
	return v.l.GetLatestPovBlock()
}

func (v *VMContext) GetAccountMeta(address types.Address, c ...storage.Cache) (*types.AccountMeta, error) {
	return v.l.GetAccountMeta(address, c...)
}

func (v *VMContext) GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error) {
	return v.l.GetTokenMeta(address, tokenType)
}

func (v *VMContext) HasAccountMetaConfirmed(address types.Address) (bool, error) {
	return v.l.HasAccountMetaConfirmed(address)
}

func (v *VMContext) GetAccountMetaByPovHeight(address types.Address, height uint64) (*types.AccountMeta, error) {
	return v.l.GetAccountMetaByPovHeight(address, height)
}

func (v *VMContext) GetTokenMetaByPovHeight(address types.Address, token types.Hash, height uint64) (*types.TokenMeta, error) {
	return v.l.GetTokenMetaByPovHeight(address, token, height)
}

func (v *VMContext) GetTokenMetaByBlockHash(hash types.Hash) (*types.TokenMeta, error) {
	return v.l.GetTokenMetaByBlockHash(hash)
}

func (v *VMContext) IteratorContractStorage(prefix []byte, callback func(key *types.ContractKey, value *types.ContractValue) error) error {
	return v.l.IteratorContractStorage(prefix, callback)
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

//func (v *VMContext) RemoveStorage(prefix, key []byte, batch ...storage.Batch) error {
//	storageKey := v.getStorageKey(prefix, key)
//	v.cache.RemoveStorage(storageKey)
//	return v.remove(storageKey, batch...)
//}

func (v *VMContext) SetObjectStorage(prefix, key []byte, value interface{}) error {
	storageKey := v.getStorageKey(prefix, key)

	v.cache.SetStorage(storageKey, value)

	return nil
}

func (v *VMContext) GetStorageByRaw(i []byte) ([]byte, error) {
	_, val, err := v.l.Get(i)
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

func (v *VMContext) GetPovMinerStat(dayIndex uint32, batch ...storage.Batch) (*types.PovMinerDayStat, error) {
	panic("implement me")
}

func (v *VMContext) Iterator(prefix []byte, fn func(key []byte, value []byte) error) error {
	iterator := v.l.NewVMIterator(v.contractAddr)
	return iterator.Next(prefix, fn)
}

//func (v *VMContext) GetStorageByKey(key []byte) ([]byte, error) {
//	if s := v.cache.GetStorage(key); s == nil {
//		if val, err := v.get(key); err == nil {
//			return val, nil
//		} else {
//			return nil, err
//		}
//	} else {
//		return s, nil
//	}
//}

func (v *VMContext) SetStorage(prefix, key []byte, value []byte) error {
	storageKey := v.getStorageKey(prefix, key)

	v.cache.SetStorage(storageKey, value)

	return nil
}

//func (v *VMContext) IteratorAll(prefix []byte, fn func(key []byte, value []byte) error) error {
//	pre := getStorageKey(prefix, nil)
//	err := v.Ledger.Iterator(pre, nil, func(key []byte, val []byte) error {
//		err := fn(key, val)
//		if err != nil {
//			v.logger.Error(err)
//		}
//		return nil
//	})
//
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

//func (v *VMContext) Iterator(prefix []byte, fn func(key []byte, value []byte) error) error {
//	pre := getStorageKey(prefix, nil)
//	err := v.Ledger.DBStore().Iterator(pre, nil, func(key []byte, val []byte) error {
//		err := fn(key, val)
//		if err != nil {
//			v.logger.Error(err)
//		}
//		return nil
//	})
//
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func (v *VMContext) get(key []byte) ([]byte, error) {
	rawKey := v.getRawStorageKey(key, nil)
	i, val, err := v.l.Get(rawKey)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}
	if i != nil {
		switch o := i.(type) {
		case types.Serializer:
			return o.Serialize()
		}
	}
	return val, nil
}

func (v *VMContext) set(key []byte, value interface{}, batch ...storage.Batch) (err error) {
	var b storage.Batch
	if len(batch) > 0 {
		b = batch[0]
	} else {
		b = v.l.DBStore().Batch(true)
		defer func() {
			if err := v.l.DBStore().PutBatch(b); err != nil {
				v.logger.Error(err)
			}
		}()
	}
	return b.Put(key, value)
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
	return v.PovGlobalState().LookupContractStateDB(*v.contractAddr)
}

func (v *VMContext) PovGlobalStateByHeight(h uint64) *statedb.PovGlobalStateDB {
	povHdr, err := v.l.GetPovHeaderByHeight(h)
	if err != nil {
		return nil
	}

	return statedb.NewPovGlobalStateDB(v.l.DBStore(), povHdr.GetStateHash())
}

func (v *VMContext) PoVContractStateByHeight(h uint64) (*statedb.PovContractStateDB, error) {
	return v.PovGlobalStateByHeight(h).LookupContractStateDB(*v.contractAddr)
}

//
//func (v *VMContext) remove(key []byte, batch ...storage.Batch) (err error) {
//	var b storage.Batch
//	if len(batch) > 0 {
//		b = batch[0]
//	} else {
//		b = v.l.DBStore().Batch(true)
//		defer func() {
//			if err := v.l.DBStore().PutBatch(b); err != nil {
//				v.logger.Error(err)
//			}
//		}()
//	}
//	return b.Delete(key)
//}

//func (v *VMContext) iteratorContractValue(predicate func(key *types.ContractKey) bool,
//	callback func(key *types.ContractKey, value *types.ContractValue) error) error {
//	return v.l.IteratorContractStorage(v.contractAddr[:], func(key *types.ContractKey, value *types.ContractValue) error {
//		if predicate(key) {
//			return callback(key, value)
//		}
//		return nil
//	})
//}

//func (v *VMContext) Iterator(prefix []byte, fn func(key []byte, value []byte) error) error {
//	return v.iteratorContractValue(func(key *types.ContractKey) bool {
//		// get all account latest records
//		return len(key.Suffix) > 0 && bytes.EqualFold(key.Suffix, ledger.LatestSuffix)
//	}, func(key *types.ContractKey, value *types.ContractValue) error {
//		if !value.BlockHash.IsZero() {
//			// get latest trie root by account
//			if val, err := v.l.GetContractValue(&types.ContractKey{
//				ContractAddress: key.ContractAddress,
//				AccountAddress:  key.AccountAddress,
//				Hash:            value.BlockHash,
//			}); err == nil {
//				if val.Root != nil && !val.Root.IsZero() {
//					t := trie.NewTrie(v.l.DBStore(), value.Root, trie.NewSimpleTrieNodePool())
//					iterator := t.NewIterator(prefix)
//					for {
//						if key, value, ok := iterator.Next(); !ok {
//							break
//						} else {
//							if err := fn(key, value); err != nil {
//								return err
//							}
//						}
//					}
//				} else {
//					v.logger.Debugf("%: %s latest trie root is empty", key.ContractAddress, key.AccountAddress)
//				}
//			} else {
//				v.logger.Error(err)
//			}
//		}
//		return nil
//	})
//}

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
