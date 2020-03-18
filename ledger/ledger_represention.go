package ledger

import (
	"encoding/json"
	"fmt"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
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
			l.logger.Error(err)
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

	return l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		address, err := types.BytesToAddress(key[1:])
		if err != nil {
			return fmt.Errorf("benefit BytesToAddress: %s ", err)
		}
		benefit := new(types.Benefit)
		if err := benefit.Deserialize(val); err != nil {
			return fmt.Errorf("benefit Deserialize: %s ", err)
		}
		if err := fn(address, benefit); err != nil {
			return fmt.Errorf("benefit fn: %s ", err)
		}
		return nil
	})
}

func (l *Ledger) CountRepresentations() (uint64, error) {
	return l.store.Count([]byte{byte(storage.KeyPrefixRepresentation)})
}

func (l *Ledger) AddRepresentation(address types.Address, diff *types.Benefit, c *Cache) error {
	value, err := l.GetRepresentation(address, c)
	if err != nil && err != ErrRepresentationNotFound {
		return fmt.Errorf("GetRepresentation err: %s ,address: %s", err, address)
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
	value, err := l.GetRepresentation(address, c)
	if err != nil {
		return fmt.Errorf("GetRepresentation error: %s ,address: %s", err, address)
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
	return cache.Put(k, benefit.Clone())
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

func (l *Ledger) updateRepresentation() error {
	representMap := make(map[types.Address]*types.Benefit)
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixAccount)
	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		am := new(types.AccountMeta)
		if err := am.Deserialize(val); err != nil {
			return err
		}
		tm := am.Token(config.ChainToken())
		if tm != nil {
			if _, ok := representMap[tm.Representative]; !ok {
				representMap[tm.Representative] = types.ZeroBenefit.Clone()
			}
			representMap[tm.Representative].Balance = representMap[tm.Representative].Balance.Add(am.CoinBalance)
			representMap[tm.Representative].Vote = representMap[tm.Representative].Vote.Add(am.CoinVote)
			representMap[tm.Representative].Network = representMap[tm.Representative].Network.Add(am.CoinNetwork)
			representMap[tm.Representative].Total = representMap[tm.Representative].Total.Add(am.VoteWeight())
		}
		return nil
	})
	if err != nil {
		return err
	}
	batch := l.store.Batch(true)
	if err := batch.Drop([]byte{byte(storage.KeyPrefixRepresentation)}); err != nil {
		return err
	}
	for address, benefit := range representMap {
		key, err := storage.GetKeyOfParts(storage.KeyPrefixRepresentation, address)
		if err != nil {
			return err
		}
		val, err := benefit.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := batch.Put(key, val); err != nil {
			return err
		}
	}
	return l.store.PutBatch(batch)
}
