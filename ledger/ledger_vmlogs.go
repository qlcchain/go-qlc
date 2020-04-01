package ledger

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

type vmlogsStore interface {
	AddOrUpdateVmLogs(value *types.VmLogs, c storage.Cache) error
	GetVmLogs(key types.Hash, c ...storage.Cache) (*types.VmLogs, error)
	DeleteVmLogs(key types.Hash, c storage.Cache) error
	HasVmLogs(key types.Hash, c ...storage.Cache) (bool, error)
	SearchVmLogs(fn func(key types.Hash, value *types.VmLogs) error) error
	CountVmLogs() (uint64, error)
}

var (
	ErrVmLogsExists   = errors.New("the VmLogs is empty")
	ErrVmLogsNotFound = errors.New("VmLogs not found")
)

func (l *Ledger) AddOrUpdateVmLogs(value *types.VmLogs, c storage.Cache) error {
	key := value.Hash()
	k, err := storage.GetKeyOfParts(storage.KeyPrefixVmLogs, key)
	if err != nil {
		return err
	}
	return c.Put(k, value)
}

func (l *Ledger) GetVmLogs(key types.Hash, c ...storage.Cache) (*types.VmLogs, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixVmLogs, key)
	if err != nil {
		return nil, err
	}

	i, r, err := l.Get(k, c...)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrVmLogsNotFound
		}
		return nil, err
	}
	if i != nil {
		return i.(*types.VmLogs), nil
	}
	meta := new(types.VmLogs)
	if err := meta.Deserialize(r); err != nil {
		return nil, err
	}
	return meta, nil
}

func (l *Ledger) DeleteVmLogs(key types.Hash, c storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixVmLogs, key)
	if err != nil {
		return err
	}
	return c.Delete(k)
}

func (l *Ledger) HasVmLogs(key types.Hash, c ...storage.Cache) (bool, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixVmLogs, key)
	if err != nil {
		return false, err
	}
	_, _, err = l.Get(k, c...)
	if err != nil {
		if err == storage.KeyNotFound {
			return false, ErrVmLogsNotFound
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) SearchVmLogs(fn func(key types.Hash, value *types.VmLogs) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixVmLogs)

	return l.store.Iterator(prefix, nil, func(k []byte, v []byte) error {
		key, err := types.BytesToHash(k[1:])
		if err != nil {
			return fmt.Errorf("vmlogs key deserialize: %s", err)
		}
		value := new(types.VmLogs)
		if err := value.Deserialize(v); err != nil {
			return fmt.Errorf("vmlogs deserialize: %s", err)
		}
		if err := fn(key, value); err != nil {
			return fmt.Errorf("vmlogs fn: %s", err)
		}
		return nil
	})
}

func (l *Ledger) CountVmLogs() (uint64, error) {
	return l.store.Count([]byte{storage.KeyPrefixVmLogs})
}
