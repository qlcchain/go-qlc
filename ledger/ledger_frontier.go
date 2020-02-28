package ledger

import (
	"sort"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

type FrontierStore interface {
	GetFrontier(hash types.Hash, cache ...storage.Cache) (*types.Frontier, error)
	GetFrontiers() ([]*types.Frontier, error)
	CountFrontiers() (uint64, error)
}

func (l *Ledger) GetFrontier(hash types.Hash, c ...storage.Cache) (*types.Frontier, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixFrontier, hash)
	if err != nil {
		return nil, err
	}
	frontier := types.Frontier{HeaderBlock: hash}
	if r, err := l.getFromCache(k, c...); r != nil {
		h := r.(*types.Hash)
		frontier.OpenBlock = *h
		return &frontier, nil
	} else {
		if err == ErrKeyDeleted {
			return nil, ErrFrontierNotFound
		}
	}

	v, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrFrontierNotFound
		}
		return nil, err
	}
	open := new(types.Hash)
	if err := open.Deserialize(v); err != nil {
		l.logger.Error(err)
		return nil, err
	}
	//if _, err := open.UnmarshalMsg(v); err != nil {
	//	l.logger.Error(err)
	//	return nil, err
	//}
	frontier.OpenBlock = *open
	return &frontier, nil
}

func (l *Ledger) GetFrontiers() ([]*types.Frontier, error) {
	var frontiers []*types.Frontier
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixFrontier)

	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		var frontier types.Frontier
		copy(frontier.HeaderBlock[:], key[1:])
		open := new(types.Hash)
		if err := open.Deserialize(val); err != nil {
			l.logger.Error(err)
			return err
		}
		frontier.OpenBlock = *open
		frontiers = append(frontiers, &frontier)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Sort(types.Frontiers(frontiers))
	return frontiers, nil
}

func (l *Ledger) CountFrontiers() (uint64, error) {
	return l.store.Count([]byte{byte(storage.KeyPrefixFrontier)})
}

func (l *Ledger) AddFrontier(frontier *types.Frontier, c *Cache) error {
	v := frontier.OpenBlock
	k, err := storage.GetKeyOfParts(storage.KeyPrefixFrontier, frontier.HeaderBlock)
	if err != nil {
		return err
	}

	return c.Put(k, &v)
}

func (l *Ledger) DeleteFrontier(key types.Hash, c *Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixFrontier, key)
	if err != nil {
		return err
	}
	return c.Delete(k)
}
