// Note: This file generate CURD code by genny automatically (https://github.com/cheekybits/genny)
// Usage:
// - install gen, command is  ` go get github.com/cheekybits/genny
// - create type in /common/types package, example: StateBlock, Hash
// - generate code, command is ` cat ledger_generic_cache.go | genny gen "GenericTypeC=StateBlock GenericKeyC=Hash" > ledger_type.go `
// - generate testcase code, command is ` cat ledger_generic_cache_test.go | genny gen "GenericTypeC=StateBlock GenericKeyC=Hash" > ledger_type_test.go `

package ledger

import (
	"errors"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	ErrGenericTypeCExists   = errors.New("the GenericTypeC is empty")
	ErrGenericTypeCNotFound = errors.New("GenericTypeC not found")
)

func (l *Ledger) AddGenericTypeC(key *types.GenericKeyC, value *types.GenericTypeC, c storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericTypeC, key)
	if err != nil {
		return err
	}
	return c.Put(k, value)
}

func (l *Ledger) GetGenericTypeC(key *types.GenericKeyC, c ...storage.Cache) (*types.GenericTypeC, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericTypeC, key)
	if err != nil {
		return nil, err
	}

	i, r, err := l.Get(k, c...)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrGenericTypeCNotFound
		}
		return nil, err
	}
	if i != nil {
		return i.(*types.GenericTypeC), nil
	}
	meta := new(types.GenericTypeC)
	if err := meta.Deserialize(r); err != nil {
		return nil, err
	}
	return meta, nil
}

func (l *Ledger) DeleteGenericTypeC(key *types.GenericKeyC, c storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericTypeC, key)
	if err != nil {
		return err
	}
	_, _, err = l.Get(k, c)
	if err != nil {
		if err == storage.KeyNotFound {
			return ErrGenericTypeCNotFound
		}
		return err
	}

	return c.Delete(k)
}

func (l *Ledger) UpdateGenericTypeC(key *types.GenericKeyC, value *types.GenericTypeC, c storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericTypeC, key)
	if err != nil {
		return err
	}
	_, _, err = l.Get(k, c)
	if err != nil {
		if err == storage.KeyNotFound {
			return ErrGenericTypeCNotFound
		}
		return err
	}

	return c.Put(k, value)
}

func (l *Ledger) AddOrUpdateGenericTypeC(key *types.GenericKeyC, value *types.GenericTypeC, c storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericTypeC, key)
	if err != nil {
		return err
	}
	return c.Put(k, value)
}

func (l *Ledger) HasGenericTypeC(key *types.GenericKeyC, c ...storage.Cache) (bool, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericTypeC, key)
	if err != nil {
		return false, err
	}
	_, _, err = l.Get(k, c...)
	if err != nil {
		if err == storage.KeyNotFound {
			return false, ErrGenericTypeCNotFound
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) GetGenericTypeCs(fn func(key *types.GenericKeyC, value *types.GenericTypeC) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixGenericTypeC)

	return l.store.Iterator(prefix, nil, func(k []byte, v []byte) error {
		key := new(types.GenericKeyC)
		if err := key.Deserialize(k[1:]); err != nil {
			return err
		}
		value := new(types.GenericTypeC)
		if err := value.Deserialize(v); err != nil {
			return err
		}
		if err := fn(key, value); err != nil {
			return err
		}
		return nil
	})
}

func (l *Ledger) CountGenericTypeCs() (uint64, error) {
	return l.store.Count([]byte{storage.KeyPrefixGenericTypeC})
}
