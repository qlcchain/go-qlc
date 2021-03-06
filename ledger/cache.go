package ledger

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluele/gcache"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
)

const defaultBlockFlushSecs = 1 * time.Second
const defaultBlockCapacity = 100000
const defaultUncheckFlushSecs = 2 * time.Minute
const defaultUncheckCapacity = 500000

type cacheType byte

const (
	block cacheType = iota
	unchecked
)

type MemoryCache struct {
	l          *Ledger
	caches     []*Cache
	cacheCount int
	writeIndex int
	readIndex  int

	lastFlush     time.Time
	flushInterval time.Duration
	flushStatue   bool
	flushChan     chan bool
	closedChan    chan bool
	lock          sync.Mutex
	typ           cacheType

	tempCaches []*Cache
	tempLock   sync.Mutex

	logger *zap.SugaredLogger
}

func NewMemoryCache(ledger *Ledger, flushInterval time.Duration, size int, typ cacheType) *MemoryCache {
	lc := &MemoryCache{
		l:             ledger,
		cacheCount:    6,
		writeIndex:    1,
		readIndex:     0,
		caches:        make([]*Cache, 0),
		lastFlush:     time.Now(),
		flushInterval: flushInterval,
		flushStatue:   false,
		flushChan:     make(chan bool, 10),
		closedChan:    make(chan bool),
		lock:          sync.Mutex{},
		tempLock:      sync.Mutex{},
		typ:           typ,
		logger:        log.NewLogger("ledger/cache"),
	}
	for i := 0; i < lc.cacheCount; i++ {
		lc.caches = append(lc.caches, newCache(i, size, typ))
	}
	for i := 0; i < 30; i++ {
		lc.tempCaches = append(lc.tempCaches, newTempCache())
	}
	go lc.flushCache()
	return lc
}

func (lc *MemoryCache) ResetCapacity(size int) {
	for _, m := range lc.caches {
		m.cache = gcache.New(size).Build()
	}
}

// get write cache index
func (lc *MemoryCache) GetCache() *Cache {
	//lc.logger.Info("GetCache")
	lc.lock.Lock()
	defer lc.lock.Unlock()
	if lc.needsFlush() {
		//lc.logger.Debug("current write cache need flush: ", lc.writeIndex)
		newWriteIndex := (lc.writeIndex + 1) % lc.cacheCount // next write cache index
		st := time.Now()
		for {
			if newWriteIndex != lc.readIndex { // next write cache index must flush done
				break
			} else {
				time.Sleep(500 * time.Millisecond)
			}
			if time.Now().Sub(st) > 2*time.Second {
				lc.logger.Error("cache flush timeout ", newWriteIndex)
			}
		}
		lc.lastFlush = time.Now()

		if len(lc.flushChan) > 1 { // if disk write too slowly
			<-lc.flushChan
		}
		lc.flushChan <- true
		lc.writeIndex = newWriteIndex
		//lc.logger.Debug("new write cache index: ", lc.writeIndex)
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

func (lc *MemoryCache) flushCache() {
	ticker := time.NewTicker(lc.flushInterval)
	for {
		select {
		case <-ticker.C:
			lc.GetCache()
		case <-lc.flushChan:
			lc.flush()
		case <-lc.l.ctx.Done():
			lc.close()
			lc.closedChan <- true
			return
			//default:
			//	time.Sleep(1 * time.Second)
		}
	}
}

func (lc *MemoryCache) flush() {
	if lc.flushStatue {
		return
	}
	lc.flushStatue = true
	defer func() {
		lc.flushStatue = false
	}()
	index := lc.readIndex
	index = (index + 1) % lc.cacheCount // next read cache index to dump
	for index != lc.writeIndex {
		//lc.logger.Debug("     begin flush cache: ", index)
		if err := lc.caches[index].flush(lc.l); err != nil {
			lc.logger.Error(err)
		}
		//lc.logger.Debug("     flush done and new read index: ", index)
		lc.readIndex = index
		index = (index + 1) % lc.cacheCount // next read cache index to dump
	}
}

func (lc *MemoryCache) closed() {
	<-lc.closedChan
	lc.logger.Info("cache closed")
}

func (lc *MemoryCache) close() error {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	index := lc.readIndex
	index = (index + 1) % lc.cacheCount // next read cache index to dump
	finish := false
	for !finish {
		if index == lc.writeIndex {
			finish = true
		}
		//lc.logger.Debug("  begin flush cache: ", index)
		if err := lc.caches[index].flush(lc.l); err != nil {
			lc.logger.Error(err)
		}
		//lc.logger.Debug("  flush done and new read index: ", index)
		lc.readIndex = index
		index = (index + 1) % lc.cacheCount // next read cache index to dump
	}
	return nil
}

func (lc *MemoryCache) rebuild() error {
	if err := lc.close(); err != nil {
		return err
	}
	lc.writeIndex = 1
	lc.readIndex = 0
	return nil
}

func (lc *MemoryCache) getTempCache() *Cache {
	lc.tempLock.Lock()
	defer lc.tempLock.Unlock()
	for i := 0; i < len(lc.tempCaches); i++ {
		c := lc.tempCaches[i]
		if c.quote == 0 {
			if c.capacity() > 0 {
				lc.logger.Error("cache should empty")
				break
			}
			c.quote = 1
			return c
		}
	}
	return newTempCache()
}

func (lc *MemoryCache) releaseTempCache(c *Cache) {
	c.purge()
	c.quote = 0
}

func (lc *MemoryCache) Get(key []byte) (interface{}, error) {
	index := lc.writeIndex
	readIndex := lc.readIndex
	count := 0
	for index != readIndex {
		count++
		if count == lc.cacheCount {
			lc.logger.Error("cache get loop")
		}
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

func (lc *MemoryCache) Delete(key []byte) error {
	c := lc.GetCache()
	return c.Delete(key)
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

func (lc *MemoryCache) prefixIterator(prefix []byte, fn func(k []byte, v []byte) error) ([][]byte, error) {
	keys := make([][]byte, 0)
	index := lc.writeIndex
	readIndex := lc.readIndex
	for index != readIndex {
		items := lc.caches[index].cache.GetALL(false)
		for k, v := range items {
			key := originalKey(k.(string))
			if bytes.HasPrefix(key, prefix) {
				if !contain(keys, key) && !isDeleteKey(v) {
					if err := fn(key, v.([]byte)); err != nil {
						return nil, fmt.Errorf("ledger iterator: %s", err)
					}
				}
				keys = append(keys, key)
			}
		}
		index = (index - 1) % lc.cacheCount
		if index < 0 {
			index = lc.cacheCount - 1
		}
	}
	return keys, nil
}

func (lc *MemoryCache) prefixIteratorObject(prefix []byte, fn func(k []byte, v interface{}) error) ([][]byte, error) {
	keys := make([][]byte, 0)
	index := lc.writeIndex
	readIndex := lc.readIndex
	for index != readIndex {
		items := lc.caches[index].cache.GetALL(false)
		for k, v := range items {
			key := originalKey(k.(string))
			if bytes.HasPrefix(key, prefix) {
				if !contain(keys, key) && !isDeleteKey(v) {
					if err := fn(key, v); err != nil {
						return nil, fmt.Errorf("ledger iterator: %s", err)
					}
				}
				keys = append(keys, key)
			}
		}
		index = (index - 1) % lc.cacheCount
		if index < 0 {
			index = lc.cacheCount - 1
		}
	}
	return keys, nil
}

func contain(kvs [][]byte, key []byte) bool {
	for _, kv := range kvs {
		if bytes.EqualFold(kv, key) {
			return true
		}
	}
	return false
}

func (lc *MemoryCache) BatchUpdate(fn func(c *Cache) error) error {
	c := lc.getTempCache()
	defer lc.releaseTempCache(c)
	if err := fn(c); err != nil {
		return err
	}

	mc := lc.GetCache()
	mc.Quoted()
	defer mc.Release()
	for k, v := range c.cache.GetALL(false) {
		if err := mc.Put(originalKey(k.(string)), v); err != nil {
			return err
		}
	}
	return nil
	//mc := lc.GetCache()
	//mc.Quoted()
	//defer mc.Release()
	//if err := fn(mc); err != nil {
	//	return err
	//}
	//return nil
}

type Cache struct {
	//used is true
	index       int
	quote       int32
	flushLock   sync.Mutex
	flushStatue bool
	typ         cacheType

	cache  gcache.Cache
	logger *zap.SugaredLogger
}

func newCache(index int, size int, typ cacheType) *Cache {
	return &Cache{
		index:  index,
		cache:  gcache.New(size).Build(),
		typ:    typ,
		logger: log.NewLogger("ledger/cache"),
	}
}

func newTempCache() *Cache {
	return &Cache{
		cache:  gcache.New(10000).Build(),
		logger: log.NewLogger("ledger/cache"),
	}
}

func (c *Cache) flush(l *Ledger) error {
	cs := new(CacheStat)
	cs.Index = c.index
	cs.Key = c.capacity()
	cs.Start = time.Now().UnixNano()
	defer func() {
		cs.End = time.Now().UnixNano()
	}()
	l.updateCacheStat(cs, c.typ)

	c.flushLock.Lock()
	defer func() {
		c.flushLock.Unlock()
	}()
	st := time.Now()
	for {
		if c.quote == 0 {
			break
		} else {
			time.Sleep(100 * time.Millisecond)
		}
		if time.Now().Sub(st) > 1*time.Second {
			c.logger.Error("cache quote timeout")
		}
	}

	// Nothing to do if there is no data to flush.
	if c.capacity() == 0 {
		return nil
	}

	batch := l.store.Batch(false)
	defer batch.Discard()
	for k, v := range c.cache.GetALL(false) {
		key := originalKey(k.(string))
		if bytes.EqualFold(key[:1], []byte{byte(storage.KeyPrefixBlock)}) {
			cs.Block = cs.Block + 1
		}
		if isDeleteKey(v) {
			cs.Delete = cs.Delete + 1
		}
		if err := c.dumpToLevelDb(key, v, batch); err != nil {
			return fmt.Errorf("dump to store: %s ", err)
		}
		if err := c.dumpToRelation(key, v, l); err != nil {
			return fmt.Errorf("dump to relation: %s ", err)
		}
	}
	if err := l.store.PutBatch(batch); err != nil {
		return fmt.Errorf("store put batch: %s ", err)
	}
	c.purge()
	return nil
}

func (c *Cache) dumpToLevelDb(key []byte, v interface{}, b storage.Batch) error {
	if !isDeleteKey(v) {
		switch o := v.(type) {
		case types.Serializer:
			val, err := o.Serialize()
			if err != nil {
				return fmt.Errorf("serialize error,  %s", key[:1])
			}
			return b.Put(key, val)
		case []byte:
			return b.Put(key, o)
		default:
			return fmt.Errorf("missing method serialize: %s", key[:1])
		}
	} else {
		return b.Delete(key)
	}
}

func (c *Cache) dumpToRelation(key []byte, v interface{}, l *Ledger) error {
	if !isDeleteKey(v) {
		if val, ok := v.(types.Convert); ok {
			objs, err := val.ConvertToSchema()
			if err != nil {
				return fmt.Errorf("table convert: %s", err)
			}
			l.relation.Add(objs)
		}
	}
	for _, obj := range l.deletedSchema {
		l.relation.Delete(obj)
	}
	l.deletedSchema = l.deletedSchema[:0]
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
	panic("not implemented")
}

func (c *Cache) Discard() {
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

type CacheStore interface {
	Cache() *MemoryCache
	GetCacheStat() []*CacheStat
	GetCacheStatue() map[string]string
	UCache() *MemoryCache
	GetUCacheStat() []*CacheStat
	GetUCacheStatue() map[string]string
}

func (l *Ledger) Cache() *MemoryCache {
	return l.cache
}

func (l *Ledger) GetCacheStat() []*CacheStat {
	return l.cacheStats
}

func (l *Ledger) GetCacheStatue() map[string]string {
	r := make(map[string]string)
	for i, c := range l.cache.caches {
		r["c"+strconv.Itoa(i)] = strconv.Itoa(c.capacity())
	}
	r["read"] = strconv.Itoa(l.cache.readIndex)
	r["write"] = strconv.Itoa(l.cache.writeIndex)
	r["lastflush"] = l.cache.lastFlush.Format("2006-01-02 15:04:05")
	r["flushStatue"] = strconv.FormatBool(l.cache.flushStatue)
	r["flushChan"] = strconv.Itoa(len(l.cache.flushChan))
	return r
}

func (l *Ledger) UCache() *MemoryCache {
	return l.unCheckCache
}

func (l *Ledger) GetUCacheStat() []*CacheStat {
	return l.uCacheStats
}

func (l *Ledger) GetUCacheStatue() map[string]string {
	r := make(map[string]string)
	if l.hasUncheckedCache {
		for i, c := range l.unCheckCache.caches {
			r["c"+strconv.Itoa(i)] = strconv.Itoa(c.capacity())
		}
		r["read"] = strconv.Itoa(l.unCheckCache.readIndex)
		r["write"] = strconv.Itoa(l.unCheckCache.writeIndex)
		r["lastflush"] = l.unCheckCache.lastFlush.Format("2006-01-02 15:04:05")
		r["flushStatue"] = strconv.FormatBool(l.unCheckCache.flushStatue)
		r["flushChan"] = strconv.Itoa(len(l.unCheckCache.flushChan))
	}
	return r
}
