/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"github.com/qlcchain/go-qlc/common/relation"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/trie"
)

type VMCache struct {
	logList   types.VmLogs
	storage   map[string]interface{}
	trie      *trie.Trie
	trieDirty bool
}

func NewVMCache() *VMCache {
	return &VMCache{
		trie:      trie.NewTrie(nil, nil, trie.NewSimpleTrieNodePool()),
		storage:   make(map[string]interface{}),
		trieDirty: false,
		logList:   types.VmLogs{Logs: make([]*types.VmLog, 0)},
	}
}

func (cache *VMCache) Trie(fn func([]byte) []byte) *trie.Trie {
	if cache.trieDirty {
		for key, value := range cache.storage {
			if data, err := relation.ConvertToBytes(value); err == nil {
				if fn != nil {
					cache.trie.SetValue(fn([]byte(key)), data)
				} else {
					cache.trie.SetValue([]byte(key), data)
				}
			}
		}

		cache.storage = make(map[string]interface{})
		cache.trieDirty = false
	}
	return cache.trie
}

func (cache *VMCache) SetStorage(key []byte, value interface{}) {
	if value == nil {
		value = make([]byte, 0)
	}

	cache.storage[string(key)] = value
	cache.trieDirty = true
}

func (cache *VMCache) GetStorage(key []byte) (interface{}, bool) {
	val, ok := cache.storage[string(key)]

	return val, ok
}

func (cache *VMCache) AppendLog(log *types.VmLog) {
	cache.logList.Logs = append(cache.logList.Logs, log)
}

func (cache *VMCache) LogList() types.VmLogs {
	return cache.logList
}

func (cache *VMCache) Clear() {
	cache.logList.Logs = cache.logList.Logs[:0]
	cache.trie = trie.NewTrie(nil, nil, trie.NewSimpleTrieNodePool())
	cache.trieDirty = false
	cache.storage = make(map[string]interface{})
}
