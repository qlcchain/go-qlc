package ledger

import (
	"fmt"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

type PendingStore interface {
	GetPending(pendingKey *types.PendingKey) (*types.PendingInfo, error)
	GetPendings(fn func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error) error
	GetPendingsByAddress(address types.Address, fn func(key *types.PendingKey, value *types.PendingInfo) error) error
	GetPendingsByToken(account types.Address, token types.Hash, fn func(key *types.PendingKey, value *types.PendingInfo) error) error
	PendingAmount(address types.Address, token types.Hash) (types.Balance, error)
	AddPending(key *types.PendingKey, value *types.PendingInfo, c *Cache) error
	DeletePending(key *types.PendingKey, c *Cache) error
}

func (l *Ledger) AddPending(key *types.PendingKey, value *types.PendingInfo, c *Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPending, key)
	if err != nil {
		return err
	}
	if err := c.Put(k, value); err != nil {
		return err
	}
	return l.rcache.UpdateAccountPending(key, value, true)
}

func (l *Ledger) DeletePending(key *types.PendingKey, c *Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPending, key)
	if err != nil {
		return err
	}
	info, err := l.GetPending(key)
	if err == nil {
		if err := c.Delete(k); err != nil {
			return err
		}
		return l.rcache.UpdateAccountPending(key, info, false)
	}
	return nil
}

func (l *Ledger) GetPending(pendingKey *types.PendingKey) (*types.PendingInfo, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPending, pendingKey)
	if err != nil {
		return nil, err
	}

	r, err := l.getFromCache(k)
	if r != nil {
		return r.(*types.PendingInfo), nil
	} else {
		if err == ErrKeyDeleted {
			return nil, ErrPendingNotFound
		}
	}

	v, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrPendingNotFound
		}
		return nil, err
	}
	meta := new(types.PendingInfo)
	if err := meta.Deserialize(v); err != nil {
		return nil, err
	}
	return meta, nil
}

func (l *Ledger) GetPendings(fn func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPending)
	return l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		pendingKey := new(types.PendingKey)
		if err := pendingKey.Deserialize(key[1:]); err != nil {
			return fmt.Errorf("pendingKey deserialize: %s", err)
		}
		pendingInfo := new(types.PendingInfo)
		if err := pendingInfo.Deserialize(val); err != nil {
			return fmt.Errorf("pendingKey deserialize: %s", err)
		}
		if err := fn(pendingKey, pendingInfo); err != nil {
			return fmt.Errorf("pendingKey deserialize: %s", err)
		}
		return nil
	})
}

func (l *Ledger) GetPendingsByAddress(address types.Address, fn func(key *types.PendingKey, value *types.PendingInfo) error) error {
	pre := make([]byte, 0)
	pre = append(pre, byte(storage.KeyPrefixPending))
	pre = append(pre, address.Bytes()...)
	return l.store.Iterator(pre, nil, func(key []byte, val []byte) error {
		pendingKey := new(types.PendingKey)
		if err := pendingKey.Deserialize(key[1:]); err != nil {
			return err
		}
		pendingInfo := new(types.PendingInfo)
		if err := pendingInfo.Deserialize(val); err != nil {
			return err
		}
		if err := fn(pendingKey, pendingInfo); err != nil {
			return fmt.Errorf("process pending: %s", err)
		}
		return nil
	})
}

func (l *Ledger) GetPendingsByToken(account types.Address, token types.Hash, fn func(key *types.PendingKey, value *types.PendingInfo) error) error {
	err := l.GetPendingsByAddress(account, func(key *types.PendingKey, value *types.PendingInfo) error {
		if value.Type == token {
			return fn(key, value)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("process pending by token: %s", err)
	}
	return nil
}

func (l *Ledger) PendingAmount(address types.Address, token types.Hash) (types.Balance, error) {
	b, err := l.rcache.GetAccountPending(address, token)
	if err == nil {
		return b, nil
	}
	pendingAmount := types.ZeroBalance
	if err := l.GetPendingsByToken(address, token, func(pk *types.PendingKey, pv *types.PendingInfo) error {
		pendingAmount = pendingAmount.Add(pv.Amount)
		return nil
	}); err != nil {
		return types.ZeroBalance, err
	}
	if err := l.rcache.AddAccountPending(address, token, pendingAmount); err != nil {
		return types.ZeroBalance, err
	}
	return pendingAmount, nil
}
