package ledger

import (
	"sync/atomic"

	"github.com/qlcchain/go-qlc/common/sync/hashmap"
	"github.com/qlcchain/go-qlc/common/sync/spinlock"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
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

//func (r *RepresentationCache) setToCache(key interface{}, value interface{}, txn db.StoreTxn) error {
//	benefit := value.(*types.Benefit)
//	address := key.(types.Address)
//	rKey, err := getKeyOfParts(idPrefixRepresentationCache, address, atomic.AddInt64(r.order, 1))
//	if err != nil {
//		return err
//	}
//	rVal, err := benefit.MarshalMsg(nil)
//	if err != nil {
//		r.logger.Errorf("MarshalMsg benefit error: %s ,address: %s, val: %s", err, address, benefit)
//		return err
//	}
//
//	if err := txn.Set(rKey, rVal); err != nil {
//		r.logger.Error(err)
//		return err
//	}
//	return nil
//}
//
//func (r *RepresentationCache) getFromCache(key interface{}, txn db.StoreTxn) (interface{}, error) {
//	return nil, nil
//}

type RepresentationCacheType struct {
	address types.Address
	benefit *types.Benefit
	order   int64
}

func (r *RepresentationCache) cacheToConfirmed(txn db.StoreTxn) error {
	// get all cache
	rcs := make([]*RepresentationCacheType, 0)
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
		order := int64(util.BE_BytesToUint64(cacheKey[1+types.AddressSize:]))

		rc := &RepresentationCacheType{
			address: addrCache,
			benefit: &beCache,
			order:   order,
		}

		rcs = append(rcs, rc)
		return nil
	})
	if err != nil {
		r.logger.Error(err)
		return err
	}

	if len(rcs) > 0 {
		// get last cache for each address
		t := make(map[types.Address]*RepresentationCacheType)
		for _, rc := range rcs {
			if lastCache, ok := t[rc.address]; ok {
				if rc.order > lastCache.order {
					t[rc.address] = rc
				}
			} else {
				t[rc.address] = rc
			}
		}

		// save last cache to confirmed
		for address, lastCache := range t {
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

		// delete cache
		for _, v := range rcs {
			cKey, err := getKeyOfParts(idPrefixRepresentationCache, v.address, v.order)
			if err != nil {
				return err
			}
			if err := txn.Delete(cKey); err != nil {
				r.logger.Error(err)
				return err
			}
		}
	}
	return nil
}
