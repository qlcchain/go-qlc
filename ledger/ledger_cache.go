package ledger

import (
	"encoding/hex"
	"sync/atomic"

	"github.com/bluele/gcache"
	"github.com/cornelk/hashmap"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/sync/spinlock"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
)

type LedgerCache interface {
	getFromMemory(key string) (interface{}, bool)
	setToMemory(key string, benefit *types.Benefit)
	updateMemory(key interface{}, value interface{}, txn db.StoreTxn) error
	iterMemory(fn func(string, interface{}) error) error

	cacheToConfirmed(txn db.StoreTxn) error
}

type RepresentationCache struct {
	order          *int64
	representLock  *hashmap.HashMap
	representation *hashmap.HashMap
	logger         *zap.SugaredLogger
}

func NewRepresentationCache() *RepresentationCache {
	return &RepresentationCache{
		order:          new(int64),
		representation: &hashmap.HashMap{},
		representLock:  &hashmap.HashMap{},
		logger:         log.NewLogger("ledger_cache"),
	}
}

func (r *RepresentationCache) iterMemory(fn func(string, interface{}) error) error {
	for kv := range r.representation.Iter() {
		k := kv.Key.(string)
		v := kv.Value.(*types.Benefit)
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (r *RepresentationCache) getLock(key string) *spinlock.SpinLock {
	i, _ := r.representLock.GetOrInsert(key, &spinlock.SpinLock{})
	spin, _ := i.(*spinlock.SpinLock)
	return spin
}

func (r *RepresentationCache) updateMemory(key interface{}, value interface{}, txn db.StoreTxn) error {
	benefit := value.(*types.Benefit)
	address := key.(types.Address)
	r.setToMemory(address.String(), benefit)

	rKey, err := getKeyOfParts(idPrefixRepresentationCache, address, atomic.AddInt64(r.order, 1))
	if err != nil {
		r.logger.Error(err)
		return err
	}
	rVal, err := benefit.MarshalMsg(nil)
	if err != nil {
		r.logger.Errorf("MarshalMsg benefit error: %s ,address: %s, val: %s", err, address, benefit)
		return err
	}

	if err := txn.Set(rKey, rVal); err != nil {
		r.logger.Error(err)
		return err
	}
	return nil
}

func (r *RepresentationCache) getFromMemory(address string) (interface{}, bool) {
	if lr, ok := r.representation.Get(address); ok {
		return lr.(*types.Benefit), true
	} else {
		return nil, false
	}
}

func (r *RepresentationCache) setToMemory(address string, benefit *types.Benefit) {
	r.representation.Set(address, benefit)
}

type RepresentationCacheType struct {
	address types.Address
	benefit *types.Benefit
	order   int64
}

func (r *RepresentationCache) cacheToConfirmed(txn db.StoreTxn) error {
	// get all cache
	rts := make(map[types.Address]*RepresentationCacheType)

	err := txn.Iterator(idPrefixRepresentationCache, func(cacheKey []byte, cacheVal []byte, b byte) error {
		addrCache, err := types.BytesToAddress(cacheKey[1 : 1+types.AddressSize])
		if err != nil {
			r.logger.Error(err)
			return err
		}
		var beCache types.Benefit
		if _, err := beCache.UnmarshalMsg(cacheVal); err != nil {
			r.logger.Error(err)
			return err
		}
		order := int64(util.BE_BytesToUint64(cacheKey[1+types.AddressSize : 1+types.AddressSize+8]))

		rc := &RepresentationCacheType{
			address: addrCache,
			benefit: &beCache,
			order:   order,
		}

		// get last cache for each address
		if lastCache, ok := rts[rc.address]; ok {
			if rc.order >= lastCache.order {
				rts[rc.address] = rc
			}
		} else {
			rts[rc.address] = rc
		}
		return nil
	})
	if err != nil {
		r.logger.Error(err)
		return err
	}

	if len(rts) > 0 {
		// save last cache to confirmed
		for address, lastCache := range rts {
			key, err := getKeyOfParts(idPrefixRepresentation, address)
			if err != nil {
				return err
			}
			val, err := lastCache.benefit.MarshalMsg(nil)
			if err != nil {
				r.logger.Errorf("MarshalMsg benefit error: %s ,address: %s, val: %s", err, address, lastCache.benefit)
				return err
			}
			if err := txn.Set(key, val); err != nil {
				return err
			}
		}
	}

	// delete cache
	if err := txn.Drop([]byte{idPrefixRepresentationCache}); err != nil {
		return err
	}
	return nil
}

func (r *RepresentationCache) memoryToConfirmed(txn db.StoreTxn) error {
	cache := make([]string, 0)
	if err := txn.Iterator(idPrefixRepresentationCache, func(cacheKey []byte, cacheVal []byte, b byte) error {
		if len(cache) < 5000 { // delete old record, txn limit
			cache = append(cache, hex.EncodeToString(cacheKey))
		}
		return nil
	}); err != nil {
		return err
	}
	err := r.iterMemory(func(s string, i interface{}) error {
		address, err := types.HexToAddress(s)
		if err != nil {
			return err
		}
		benefit := i.(*types.Benefit)
		key, err := getKeyOfParts(idPrefixRepresentation, address)
		if err != nil {
			return err
		}
		val, err := benefit.Serialize()
		if err != nil {
			r.logger.Errorf("MarshalMsg benefit error: %s ,address: %s, val: %s", err, address, benefit)
			return err
		}
		if err := txn.Set(key, val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, k := range cache {
		b, err := hex.DecodeString(k)
		if err != nil {
			return err
		}
		if err := txn.Delete(b); err != nil {
			return err
		}
	}
	return nil
}

const (
	cacheLimit = 512
)

type Cache struct {
	confirmedAccount   gcache.Cache
	confirmedBlock     gcache.Cache
	unConfirmedAccount gcache.Cache
	unConfirmedBlock   gcache.Cache
	accountPending     gcache.Cache
}

func NewCache() *Cache {
	return &Cache{
		confirmedAccount:   gcache.New(cacheLimit).LRU().Build(),
		confirmedBlock:     gcache.New(cacheLimit).LRU().Build(),
		unConfirmedAccount: gcache.New(cacheLimit).LRU().Build(),
		unConfirmedBlock:   gcache.New(cacheLimit).LRU().Build(),
		accountPending:     gcache.New(cacheLimit).LRU().Build(),
	}
}
