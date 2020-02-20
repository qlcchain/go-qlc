package ledger

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/hashmap"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/sync/spinlock"
	"github.com/qlcchain/go-qlc/common/types"
)

type RepresentationStore interface {
	GetRepresentation(key types.Address, c ...storage.Cache) (*types.Benefit, error)
	GetRepresentations(fn func(types.Address, *types.Benefit) error) error
	CountRepresentations() (uint64, error)

	GetOnlineRepresentations() ([]types.Address, error)
	SetOnlineRepresentations(addresses []*types.Address) error
}

func (l *Ledger) GetRepresentation(key types.Address, c ...storage.Cache) (*types.Benefit, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixRepresentation, key)
	if err != nil {
		return nil, err
	}

	r, err := l.getFromCache(k, c...)
	if r != nil {
		return r.(*types.Benefit).Clone(), nil
	} else {
		if err == ErrKeyDeleted {
			return types.ZeroBenefit.Clone(), ErrRepresentationNotFound
		}
	}

	v, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return types.ZeroBenefit.Clone(), ErrRepresentationNotFound
		}
		return nil, err
	}
	meta := new(types.Benefit)
	if err := meta.Deserialize(v); err != nil {
		return nil, err
	}
	return meta, nil
}

func (l *Ledger) GetRepresentations(fn func(types.Address, *types.Benefit) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixRepresentation)

	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		address, err := types.BytesToAddress(key[1:])
		if err != nil {
			l.logger.Error(err)
			return err
		}
		benefit := new(types.Benefit)
		if err := benefit.Deserialize(val); err != nil {
			l.logger.Error(err)
			return err
		}
		if err := fn(address, benefit); err != nil {
			l.logger.Error(err)
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) CountRepresentations() (uint64, error) {
	return l.store.Count([]byte{byte(storage.KeyPrefixRepresentation)})
}

func (l *Ledger) AddRepresentation(address types.Address, diff *types.Benefit, c *Cache) error {
	spin := l.representCache.getLock(address.String())
	spin.Lock()
	defer spin.Unlock()

	value, err := l.GetRepresentation(address, c)
	if err != nil && err != ErrRepresentationNotFound {
		l.logger.Errorf("getRepresentation error: %s ,address: %s", err, address)
		return err
	}

	value.Balance = value.Balance.Add(diff.Balance)
	value.Vote = value.Vote.Add(diff.Vote)
	value.Network = value.Network.Add(diff.Network)
	value.Oracle = value.Oracle.Add(diff.Oracle)
	value.Storage = value.Storage.Add(diff.Storage)
	value.Total = value.Total.Add(diff.Total)

	return l.setRepresentation(address, value, c)
}

func (l *Ledger) SubRepresentation(address types.Address, diff *types.Benefit, c *Cache) error {
	spin := l.representCache.getLock(address.String())
	spin.Lock()
	defer spin.Unlock()

	value, err := l.GetRepresentation(address, c)
	if err != nil {
		l.logger.Errorf("GetRepresentation error: %s ,address: %s", err, address)
		return err
	}
	value.Balance = value.Balance.Sub(diff.Balance)
	value.Vote = value.Vote.Sub(diff.Vote)
	value.Network = value.Network.Sub(diff.Network)
	value.Oracle = value.Oracle.Sub(diff.Oracle)
	value.Storage = value.Storage.Sub(diff.Storage)
	value.Total = value.Total.Sub(diff.Total)

	return l.setRepresentation(address, value, c)
}

func (l *Ledger) setRepresentation(address types.Address, benefit *types.Benefit, cache *Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixRepresentation, address)
	if err != nil {
		return err
	}
	return cache.Put(k, benefit)
}

func (l *Ledger) GetOnlineRepresentations() ([]types.Address, error) {
	key := []byte{byte(storage.KeyPrefixOnlineReps)}
	var result []types.Address
	bytes, err := l.store.Get(key)
	if err != nil {
		return nil, err
	}
	if len(bytes) > 0 {
		if err := json.Unmarshal(bytes, &result); err != nil {
			return nil, err
		}
	}
	return result, err
}

func (l *Ledger) SetOnlineRepresentations(addresses []*types.Address) error {
	bytes, err := json.Marshal(addresses)
	if err != nil {
		return err
	}
	key := []byte{byte(storage.KeyPrefixOnlineReps)}
	return l.store.Put(key, bytes)
}

type RepresentationCache struct {
	representLock *hashmap.HashMap
}

func NewRepresentationCache() *RepresentationCache {
	return &RepresentationCache{
		representLock: &hashmap.HashMap{},
	}
}

func (r *RepresentationCache) getLock(key string) *spinlock.SpinLock {
	i, _ := r.representLock.GetOrInsert(key, &spinlock.SpinLock{})
	spin, _ := i.(*spinlock.SpinLock)
	return spin
}
