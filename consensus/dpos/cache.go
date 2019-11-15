package dpos

import (
	"container/list"
	"sync"
	"time"
)

type EvictedFunc func(interface{}, interface{})

type item struct {
	key        interface{}
	value      interface{}
	expiration *time.Time
}

type cache struct {
	lock        sync.RWMutex
	size        int
	expiration  time.Duration
	items       map[interface{}]*list.Element
	evictList   *list.List
	evictedFunc EvictedFunc
}

func newCache(size int, expiration time.Duration) *cache {
	return &cache{
		size:       size,
		expiration: expiration,
		items:      make(map[interface{}]*list.Element, size),
		evictList:  list.New(),
	}
}

func (c *cache) set(key, val interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.items[key]; ok {
		return
	}

	if c.evictList.Len() >= c.size {
		ent := c.evictList.Back()
		if ent == nil {
			return
		} else {
			c.evict(ent)
		}
	}

	expire := time.Now().Add(c.expiration)
	it := &item{
		key:        key,
		value:      val,
		expiration: &expire,
	}
	c.items[key] = c.evictList.PushFront(it)
}

func (c *cache) has(key interface{}) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if val, ok := c.items[key]; ok {
		it := val.Value.(*item)
		if !it.isExpired() {
			return true
		}
	}

	return false
}

func (c *cache) evict(em *list.Element) {
	c.evictList.Remove(em)
	entry := em.Value.(*item)
	delete(c.items, entry.key)

	if c.evictedFunc != nil {
		c.evictedFunc(entry.key, entry.value)
	}
}

func (c *cache) gc() {
	for _, val := range c.items {
		it := val.Value.(*item)
		if it.isExpired() {
			c.evict(val)
		}
	}
}

func (it *item) isExpired() bool {
	return it.expiration.Before(time.Now())
}
