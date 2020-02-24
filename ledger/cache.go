package ledger

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type MemoryCache struct {
	l          *Ledger
	caches     []*Cache
	cacheCount int
	writeIndex int
	readIndex  int

	lastFlush     time.Time
	flushInterval time.Duration
	//flushStatue   bool
	flushChan  chan bool
	closedChan chan bool
	lock       sync.Mutex
	logger     *zap.SugaredLogger
}

func NewMemoryCache(ledger *Ledger) *MemoryCache {
	lc := &MemoryCache{
		l:             ledger,
		cacheCount:    5,
		writeIndex:    1,
		readIndex:     0,
		caches:        make([]*Cache, 0),
		lastFlush:     time.Now(),
		flushInterval: defaultFlushSecs,
		//flushStatue:   false,
		flushChan:  make(chan bool, 1),
		closedChan: make(chan bool),
		lock:       sync.Mutex{},
		logger:     log.NewLogger("ledger/dbcache"),
	}
	for i := 0; i < lc.cacheCount; i++ {
		lc.caches = append(lc.caches, newCache())
	}
	go lc.flushCache()
	return lc
}

// get write cache index
func (lc *MemoryCache) GetCache() *Cache {
	//lc.logger.Info("GetCache")
	if lc.needsFlush() {
		lc.lock.Lock()
		if lc.needsFlush() {
			lc.logger.Debug("current write cache need flush: ", lc.writeIndex)
			lc.writeIndex = (lc.writeIndex + 1) % lc.cacheCount // next write cache index, and must flush done
			st := time.Now()
			for {
				if !lc.caches[lc.writeIndex].flushStatue {
					break
				} else {
					time.Sleep(100 * time.Millisecond)
				}
				if time.Now().Sub(st) > 2*time.Second {
					lc.logger.Error("cache flush timeout")
				}
			}
			lc.logger.Debug("new write cache index: ", lc.writeIndex)
			lc.lastFlush = time.Now()
			lc.flushChan <- true
		}
		lc.lock.Unlock()
	}
	//lc.logger.Debug("return write cache index: ", lc.writeIndex)
	cache := lc.caches[lc.writeIndex]
	return cache
}

func (lc *MemoryCache) needsFlush() bool {
	if (time.Since(lc.lastFlush) > lc.flushInterval) && lc.caches[lc.writeIndex].capacity() > 0 {
		return true
	}
	//if lc.GetCache().capacity() > defaultCapacity {
	//	return true
	//}
	return false
}

func (lc *MemoryCache) flushCache() *Cache {
	ticker := time.NewTicker(defaultFlushSecs)
	for {
		select {
		case <-ticker.C:
			lc.GetCache()
		case <-lc.flushChan:
			lc.flush()
		case <-lc.l.ctx.Done():
			lc.close()
			lc.closedChan <- true
			//default:
			//	time.Sleep(1 * time.Second)
		}
	}
}

func (lc *MemoryCache) flush() {
	//if lc.flushStatue {
	//	return
	//}
	//lc.flushStatue = true
	//defer func() {
	//	lc.flushStatue = false
	//}()
	//lc.logger.Debug("flush... ")
	index := lc.readIndex
	index = (index + 1) % lc.cacheCount // next read cache index to dump
	for index != lc.writeIndex {
		//lc.logger.Debug("  begin flush cache: ", index)
		if err := lc.caches[index].flush(lc.l, index); err != nil {
			lc.logger.Error(err)
		}
		//lc.logger.Debug("  flush done and new read index: ", index)
		lc.readIndex = index
		index = (index + 1) % lc.cacheCount // next read cache index to dump
	}
}

func (lc *MemoryCache) closed() {
	<-lc.closedChan
	lc.logger.Info("cache closed")
}

func (lc *MemoryCache) close() error {
	index := lc.readIndex
	index = (index + 1) % lc.cacheCount // next read cache index to dump
	finish := false
	for !finish {
		if index == lc.writeIndex {
			finish = true
		}
		//lc.logger.Debug("  begin flush cache: ", index)
		if err := lc.caches[index].flush(lc.l, index); err != nil {
			lc.logger.Error(err)
		}
		//lc.logger.Debug("  flush done and new read index: ", index)
		lc.readIndex = index
		index = (index + 1) % lc.cacheCount // next read cache index to dump
	}
	return nil
}

func (lc *MemoryCache) Get(key []byte) (interface{}, error) {
	index := lc.writeIndex
	readIndex := lc.readIndex
	for index != readIndex {
		if v, err := lc.caches[index].Get(key); err == nil {
			return v, nil
		} else {
			if err == ErrKeyDeleted {
				return nil, err
			}
		}
		index = (index - 1) % lc.cacheCount
		if index < 0 {
			index = lc.cacheCount - 1
		}
	}
	return nil, ErrKeyNotInCache
}

func (lc *MemoryCache) Put(key []byte, value interface{}) error {
	c := lc.GetCache()
	return c.Put(key, value)
}

func (lc *MemoryCache) Has(key []byte) (bool, error) {
	_, err := lc.Get(key)
	if err != nil {
		if err == ErrKeyDeleted {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (lc *MemoryCache) prefixIterator(prefix []byte) []*kv {
	kvs := make([]*kv, 0)
	index := lc.writeIndex
	readIndex := lc.readIndex
	for index != readIndex {
		items := lc.caches[index].cache.GetALL(false)
		for k, v := range items {
			key := originalKey(k.(string))
			if bytes.HasPrefix(key, prefix) {
				if !contain(kvs, key) && !isDeleteKey(v) {
					value := v.([]byte)
					temp := &kv{
						key:   key,
						value: value,
					}
					kvs = append(kvs, temp)
				}
			}
		}
		index = (index - 1) % lc.cacheCount
		if index < 0 {
			index = lc.cacheCount - 1
		}
	}
	return kvs
}

func contain(kvs []*kv, key []byte) bool {
	for _, kv := range kvs {
		if bytes.EqualFold(kv.key, key) {
			return true
		}
	}
	return false
}

func (lc *MemoryCache) BatchUpdate(fn func(c *Cache) error) error {
	c := newCache()
	defer func() {
		c.purge()
	}()
	if err := fn(c); err != nil {
		return err
	}

	mc := lc.GetCache()
	mc.Quoted()
	for k, v := range c.cache.GetALL(false) {
		if err := mc.Put(originalKey(k.(string)), v); err != nil {
			return err
		}
	}
	mc.Release()
	return nil
	//mc := lc.GetCache()
	//mc.Quoted()
	//defer mc.Release()
	//if err := fn(mc); err != nil {
	//	return err
	//}
	//return nil
}

const defaultFlushSecs = 1 * time.Second
const defaultCapacity = 100000

type Cache struct {
	//used is true
	quote       int32
	flushLock   sync.Mutex
	flushStatue bool

	cache  gcache.Cache
	logger *zap.SugaredLogger
}

func newCache() *Cache {
	return &Cache{
		cache:  gcache.New(defaultCapacity).Build(),
		logger: log.NewLogger("ledger/cache"),
	}
}

func (c *Cache) flush(l *Ledger, index int) error {
	cs := new(CacheStat)
	cs.Index = index
	cs.Key = c.capacity()
	cs.Start = time.Now().UnixNano()

	c.flushLock.Lock()
	defer func() {
		c.flushLock.Unlock()
	}()
	c.flushStatue = true
	defer func() {
		c.flushStatue = false
	}()
	st := time.Now()
	for {
		if c.quote == 0 {
			break
		} else {
			time.Sleep(100 * time.Millisecond)
		}
		if time.Now().Sub(st) > 3*time.Second {
			c.logger.Error("cache quote timeout")
		}
	}

	// Nothing to do if there is no data to flush.
	if c.isEmpty() {
		return nil
	}

	batch := l.store.Batch(false)
	for k, v := range c.cache.GetALL(false) {
		t := originalKey(k.(string))
		if bytes.EqualFold(t[:1], []byte{byte(storage.KeyPrefixBlock)}) {
			cs.Block = cs.Block + 1
		}
		if err := c.dumpToLevelDb(k, v, batch); err != nil {
			c.logger.Error(err)
			batch.Cancel()
			return err
		}
		if err := c.dumpToRelation(k, v, l); err != nil {
			c.logger.Error(err)
			return err
		}
	}
	if err := l.store.PutBatch(batch); err != nil {
		c.logger.Error(err)
		return err
	}
	c.purge()

	cs.End = time.Now().UnixNano()
	l.updateCacheStat(cs)
	return nil
}

func (c *Cache) dumpToLevelDb(k, v interface{}, b storage.Batch) error {
	key := originalKey(k.(string))
	if !isDeleteKey(v) {
		switch o := v.(type) {
		case types.Serialize:
			val, err := o.Serialize()

			if err != nil {
				c.logger.Error("serialize error  ", key[:1])
				return err
			}
			if err := b.Put(key, val); err != nil {
				return err
			}
		case []byte:
			if err := b.Put(key, o); err != nil {
				return err
			}
		default:
			c.logger.Error("missing method serialize:  ", key[:1])
			return fmt.Errorf("unknown type: %s", key[:1])
		}
	} else {
		if err := b.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cache) dumpToRelation(k, v interface{}, l *Ledger) error {
	if obj, ok := v.(types.Schema); ok {
		if !isDeleteKey(v) {
			l.relation.Add(obj)
		} else {
			l.relation.Delete(obj)
		}
	}
	return nil
}

func (c *Cache) Quoted() {
	atomic.AddInt32(&c.quote, 1)
}

func (c *Cache) Release() {
	atomic.AddInt32(&c.quote, -1)
}

// Clear the Cache since it has been flushed.
func (c *Cache) purge() {
	c.cache.Purge()
}

func (c *Cache) isEmpty() bool {
	return c.cache.Len(false) == 0
}

func (c *Cache) capacity() int {
	return c.cache.Len(false)
}

func (c *Cache) Put(key []byte, value interface{}) error {
	return c.cache.Set(transformKey(key), value)
}

func (c *Cache) Get(key []byte) (interface{}, error) {
	r, err := c.cache.Get(transformKey(key))
	if err != nil {
		return nil, err
	}
	if isDeleteKey(r) {
		return nil, ErrKeyDeleted
	}
	return r, nil
}

func (c *Cache) Iterator(prefix []byte, end []byte, f func(k, v []byte) error) error {
	//items := c.cache.GetALL(false)
	//for k, v := range items {
	//	key := originalKey(k.(string))
	//	if bytes.HasPrefix(key, prefix) {
	//		if err := f(key, v.([]byte)); err != nil {
	//			return fmt.Errorf("cache iterator error: %s", err)
	//		}
	//	}
	//}
	//return nil
	panic("not implemented")
}

func (c *Cache) Cancel() {
	panic("not implemented")
}

func (c *Cache) Delete(key []byte) error {
	return c.cache.Set(transformKey(key), deleteKeyTag)
}

func (b *Cache) Drop(prefix []byte) error {
	panic("not implemented")
}

func (b *Cache) Len() int64 {
	panic("not implemented")
}

func transformKey(k []byte) string {
	return string(k)
}

func originalKey(k string) []byte {
	return []byte(k)
}

const (
	cacheLimit = 512
)

type rCache struct {
	accountPending gcache.Cache
}

func NewrCache() *rCache {
	return &rCache{
		accountPending: gcache.New(cacheLimit).LRU().Build(),
	}
}

func isDeleteKey(v interface{}) bool {
	if _, ok := v.(*deleteKey); ok {
		return true
	}
	return false
}

type deleteKey struct {
}

var deleteKeyTag = new(deleteKey)

//func isDeleteKey(v interface{}) bool {
//	if v == nil {
//		return true
//	} else {
//		return false
//	}
//}
//
//type deleteKey struct {
//}
//var deleteKeyTag []byte = nil

var ErrKeyDeleted = errors.New("key is deleted")
var ErrKeyNotInCache = errors.New("key not in cache")
