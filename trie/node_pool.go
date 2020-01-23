/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package trie

import (
	"github.com/cornelk/hashmap"

	"github.com/qlcchain/go-qlc/common/types"
)

type NodePool struct {
	cache    *hashmap.HashMap
	limit    int
	clearNum int
}

const cacheSize = 10000

func NewTrieNodePool(limit int, clearNum int) *NodePool {
	if limit <= 0 {
		limit = cacheSize
	}
	return &NodePool{limit: limit, cache: hashmap.New(uintptr(limit)), clearNum: clearNum}
}

func NewSimpleTrieNodePool() *NodePool {
	return &NodePool{limit: cacheSize, cache: hashmap.New(uintptr(cacheSize)), clearNum: cacheSize / 2}
}

func (p *NodePool) Get(key *types.Hash) *TrieNode {
	if value, ok := p.cache.Get(key[:]); ok {
		return value.(*TrieNode)
	}
	return nil
}

func (p *NodePool) Set(key *types.Hash, trieNode *TrieNode) {
	p.cache.Set(key[:], trieNode)

	if p.cache.Len() >= p.limit {
		p.clear()
	}
}

func (p *NodePool) Clear() {
	quitCh := make(chan struct{})
	for k := range p.cache.Iter(quitCh) {
		p.cache.Del(k.Key)
	}
	close(quitCh)
}

func (p *NodePool) Len() int {
	return p.cache.Len()
}

func (p *NodePool) clear() {
	i := 0
	quitCh := make(chan struct{})
	for key := range p.cache.Iter(quitCh) {
		p.cache.Del(key.Key)
		i++
		if i >= p.clearNum {
			close(quitCh)
			return
		}
	}
	close(quitCh)
}
