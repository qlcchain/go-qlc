/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/trie"
)

type VMCache struct {
	logList types.VmLogs
	storage map[string][]byte
	delete  map[string]struct{}

	trie      *trie.Trie
	trieDirty bool
}

func NewVMCache(trie *trie.Trie) *VMCache {
	return &VMCache{
		storage:   make(map[string][]byte),
		delete:    make(map[string]struct{}),
		trie:      trie.Clone(),
		trieDirty: false,
	}
}

func (cache *VMCache) Trie() *trie.Trie {
	if cache.trieDirty {
		for key, value := range cache.storage {
			cache.trie.SetValue([]byte(key), value)
		}

		cache.storage = make(map[string][]byte)
		cache.trieDirty = false
	}
	return cache.trie
}

func (cache *VMCache) SetStorage(key []byte, value []byte) {
	if value == nil {
		value = make([]byte, 0)
	}

	cache.storage[string(key)] = value
	cache.trieDirty = true
}

func (cache *VMCache) GetStorage(key []byte) []byte {
	if value, ok := cache.storage[string(key)]; ok && value != nil {
		return value
	}

	return cache.trie.GetValue(key)
}

func (cache *VMCache) RemoveStorage(key []byte) {
	if _, ok := cache.storage[string(key)]; ok {
		delete(cache.storage, string(key))
	}
}

func (cache *VMCache) AppendLog(log *types.VmLog) {
	if cache.logList.Logs == nil {
		cache.logList.Logs = make([]*types.VmLog, 0)
	}
	cache.logList.Logs = append(cache.logList.Logs, log)
}

func (cache *VMCache) DelStorage(key []byte) {
	if _, ok := cache.delete[string(key)]; ok {
		return
	} else {
		cache.delete[string(key)] = struct{}{}
	}
}

func (cache *VMCache) LogList() types.VmLogs {
	return cache.logList
}

func (cache *VMCache) Storage() map[string][]byte {
	return cache.storage
}

func (cache *VMCache) Clear() {
	//TODO: reset trie
	cache.logList.Logs = cache.logList.Logs[:0]
	cache.trieDirty = false
	cache.storage = make(map[string][]byte)
}
