/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package trie

import (
	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common/types"
)

type NodePool struct {
	cache    gcache.Cache
	limit    int
	clearNum int
}

const cacheSize = 10000

func NewTrieNodePool(limit int, clearNum int) *NodePool {
	if limit <= 0 {
		limit = cacheSize
	}
	p := &NodePool{limit: limit, clearNum: clearNum}
	p.cache = gcache.New(limit).LRU().Build()
	return p
}

func NewSimpleTrieNodePool() *NodePool {
	clearNum := cacheSize * 80 / 100
	return NewTrieNodePool(cacheSize, clearNum)
}

func (p *NodePool) Get(key *types.Hash) *TrieNode {
	if value, err := p.cache.Get(*key); err == nil {
		return value.(*TrieNode)
	}
	return nil
}

func (p *NodePool) Set(key *types.Hash, trieNode *TrieNode) {
	_ = p.cache.Set(*key, trieNode)
}

func (p *NodePool) Clear() {
	p.cache.Purge()
}

func (p *NodePool) Len() int {
	return p.cache.Len(false)
}

func (p *NodePool) clear() {
}
