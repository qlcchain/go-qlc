/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/types"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
)

type ContractStore interface {
	GetStorage(prefix, key []byte) ([]byte, error)
	SetStorage(prefix, key []byte, value []byte) error
	Iterator(prefix []byte, fn func(key []byte, value []byte) error) error
	IsUserAccount(address types.Address) (bool, error)

	SaveStorage(batch ...storage.Batch) error
	RemoveStorage(prefix, key []byte, batch ...storage.Batch) error

	// TODO: remove
	GetStorageByKey(key []byte) ([]byte, error)
	RemoveStorageByKey(key []byte, batch ...storage.Batch) error
}

var (
	//ErrStorageExists   = errors.New("storage already exists")
	ErrStorageNotFound = errors.New("storage not found")
)

type VMContext struct {
	//TODO  should replace `ledger.Store` to `ledger.ContractStore`
	Ledger       ledger.Store
	logger       *zap.SugaredLogger
	contractAddr *types.Address
	accountAddr  types.Address
	trie         *trie.Trie
	Cache        *VMCache
}

func NewVMContext(l ledger.Store, contractAddr *types.Address) *VMContext {
	t := trie.NewTrie(l.DBStore(), nil, trie.NewSimpleTrieNodePool())
	return &VMContext{
		Ledger:       l,
		logger:       log.NewLogger("vm_context"),
		Cache:        NewVMCache(t),
		trie:         t,
		contractAddr: contractAddr,
	}
}

// WithUserAddress load latest storage trie from the specified address
func (v *VMContext) WithUserAddress(address types.Address) *VMContext {
	v.accountAddr = address
	if value, err := v.Ledger.GetContractValue(&types.ContractKey{
		ContractAddress: *v.contractAddr,
		AccountAddress:  address,
		Suffix:          ledger.LatestSuffix,
	}); err == nil && value.Root != nil {
		v.logger.Debugf("load trie by %s", value.Root.String())
		v.trie = trie.NewTrie(v.Ledger.DBStore(), value.Root, trie.NewSimpleTrieNodePool())
	}
	v.Cache = NewVMCache(v.trie)
	return v
}

// WithBlock Load storage trie from the specified user address and block hash
func NewVMContextWithBlock(l ledger.Store, block *types.StateBlock) *VMContext {
	v := NewVMContext(l, block.ContractAddress())

	t := trie.NewTrie(v.Ledger.DBStore(), nil, trie.NewSimpleTrieNodePool())
	h := block.GetHash()
	v.accountAddr = *block.ContractAddress()

	if value, err := v.Ledger.GetContractValue(&types.ContractKey{
		ContractAddress: *v.contractAddr,
		AccountAddress:  v.accountAddr,
		Hash:            h,
	}); err == nil && value.Root != nil {
		v.logger.Debugf("load trie by %s", value.Root.String())
		t = trie.NewTrie(v.Ledger.DBStore(), value.Root, trie.NewSimpleTrieNodePool())
	}
	v.Cache = NewVMCache(t)
	return v
}

func (v *VMContext) IsUserAccount(address types.Address) (bool, error) {
	if b, err := v.Ledger.HasAccountMetaConfirmed(address); b {
		return true, nil
	} else {
		if err != nil {
			return false, err
		} else {
			return false, fmt.Errorf("can not find user account %s", address)
		}
	}
}

func getStorageKey(prefix, key []byte) []byte {
	var storageKey []byte
	storageKey = append(storageKey, prefix...)
	storageKey = append(storageKey, key...)
	return storageKey
}

func (v *VMContext) GetStorage(prefix, key []byte) ([]byte, error) {
	storageKey := getStorageKey(prefix, key)
	if s := v.Cache.GetStorage(storageKey); s == nil {
		if val, err := v.get(storageKey); err == nil {
			return val, nil
		} else {
			return nil, err
		}
	} else {
		return s, nil
	}
}

func (v *VMContext) GetStorageByKey(key []byte) ([]byte, error) {
	if s := v.Cache.GetStorage(key); s == nil {
		if val, err := v.get(key); err == nil {
			return val, nil
		} else {
			return nil, err
		}
	} else {
		return s, nil
	}
}

func (v *VMContext) RemoveStorage(prefix, key []byte, batch ...storage.Batch) error {
	storageKey := getStorageKey(prefix, key)
	v.Cache.RemoveStorage(storageKey)
	return v.remove(storageKey, batch...)
}

func (v *VMContext) RemoveStorageByKey(key []byte, batch ...storage.Batch) error {
	v.Cache.RemoveStorage(key)
	return v.remove(key, batch...)
}

func (v *VMContext) SetStorage(prefix, key []byte, value []byte) error {
	storageKey := getStorageKey(prefix, key)

	v.Cache.SetStorage(storageKey, value)

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

func (v *VMContext) SaveStorage(batch ...storage.Batch) error {
	for k, val := range v.Cache.storage {
		err := v.set([]byte(k), val, batch...)
		if err != nil {
			v.logger.Error(err)
			return err
		}
	}
	return nil
}

func (v *VMContext) SaveTrie(batch ...storage.Batch) error {
	fn, err := v.Cache.Trie().Save(batch...)
	if err != nil {
		return err
	}
	fn()
	return nil
}

func (v *VMContext) get(key []byte) ([]byte, error) {
	i, val, err := v.Ledger.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}
	if i != nil {
		return i.([]byte), nil
	}
	return val, nil

	//
	//if v, err := v.Ledger.Cache().Get(key); err == nil {
	//	return v.([]byte), nil
	//}
	//val, err := v.Ledger.Store().Get(key)
	//if err != nil {
	//	if err == storage.KeyNotFound {
	//		return nil, ErrStorageNotFound
	//	}
	//	return nil, err
	//}
	//
	//s = val
	//return s, nil
}

func (v *VMContext) set(key []byte, value []byte, batch ...storage.Batch) (err error) {
	var b storage.Batch
	if len(batch) > 0 {
		b = batch[0]
	} else {
		b = v.Ledger.DBStore().Batch(true)
		defer func() {
			if err := v.Ledger.DBStore().PutBatch(b); err != nil {
				v.logger.Error(err)
			}
		}()
	}
	return b.Put(key, value)
}

func (v *VMContext) remove(key []byte, batch ...storage.Batch) (err error) {
	var b storage.Batch
	if len(batch) > 0 {
		b = batch[0]
	} else {
		b = v.Ledger.DBStore().Batch(true)
		defer func() {
			if err := v.Ledger.DBStore().PutBatch(b); err != nil {
				v.logger.Error(err)
			}
		}()
	}
	return b.Delete(key)
}

func (v *VMContext) iteratorContractValue(predicate func(key *types.ContractKey) bool,
	callback func(key *types.ContractKey, value *types.ContractValue) error) error {
	return v.Ledger.IteratorContractStorage(v.contractAddr[:], func(key *types.ContractKey, value *types.ContractValue) error {
		if predicate(key) {
			return callback(key, value)
		}
		return nil
	})
}

func (v *VMContext) Iterator(prefix []byte, fn func(key []byte, value []byte) error) error {
	return v.iteratorContractValue(func(key *types.ContractKey) bool {
		// get all account latest records
		return len(key.Suffix) > 0 && bytes.EqualFold(key.Suffix, ledger.LatestSuffix)
	}, func(key *types.ContractKey, value *types.ContractValue) error {
		if !value.BlockHash.IsZero() {
			// get latest trie root by account
			if val, err := v.Ledger.GetContractValue(&types.ContractKey{
				ContractAddress: key.ContractAddress,
				AccountAddress:  key.AccountAddress,
				Hash:            value.BlockHash,
			}); err == nil {
				if val.Root != nil && !val.Root.IsZero() {
					t := trie.NewTrie(v.Ledger.DBStore(), value.Root, trie.NewSimpleTrieNodePool())
					iterator := t.NewIterator(prefix)
					for {
						if key, value, ok := iterator.Next(); !ok {
							break
						} else {
							if err := fn(key, value); err != nil {
								return err
							}
						}
					}
				} else {
					v.logger.Debugf("%: %s latest trie root is empty", key.ContractAddress, key.AccountAddress)
				}
			} else {
				v.logger.Error(err)
			}
		}
		return nil
	})
}

func (v *VMContext) SaveBlockData(block *types.StateBlock, c ...storage.Cache) error {
	if err := v.SaveTrie(); err != nil {
		return err
	}
	trieRoot := v.Cache.Trie().Root.Hash()
	return v.Ledger.UpdateContractValueByBlock(block, trieRoot, c...)
}

func (v *VMContext) DeleteBlockData(block *types.StateBlock, c ...storage.Cache) error {
	if err := v.RemoveTrie(); err != nil {
		return err
	}
	return v.Ledger.DeleteContractValueByBlock(block, c...)
}

func (v *VMContext) RemoveTrie(batch ...storage.Batch) error {
	if v.trie != nil {
		return v.trie.Remove(batch...)
	}
	return nil
}
