// Note: This file generate CURD code by genny automatically (https://github.com/cheekybits/genny)
// Usage:
// - install gen, command is  ` go get github.com/cheekybits/genny
// - create type in /common/types package, example: StateBlock, Hash
// - generate code, command is ` cat ledger_generic.go | genny gen "GenericType=StateBlock GenericKey=Hash" > ledger_type.go `
// - generate testcase code, command is ` cat ledger_generic_test.go | genny gen "GenericType=StateBlock GenericKey=Hash" > ledger_type_test.go `

package ledger

import (
	"errors"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	ErrGenericTypeExists   = errors.New("the GenericType is empty")
	ErrGenericTypeNotFound = errors.New("GenericType not found")
)

func (l *Ledger) AddGenericType(key *types.GenericKey, value *types.GenericType) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericType, key)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	if _, err := l.store.Get(k); err == nil {
		return ErrGenericTypeExists
	} else if err != storage.KeyNotFound {
		return err
	}
	return l.store.Put(k, v)
}

func (l *Ledger) GetGenericType(key *types.GenericKey) (*types.GenericType, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericType, key)
	if err != nil {
		return nil, err
	}

	value := new(types.GenericType)
	v, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrGenericTypeNotFound
		}
		return nil, err
	}
	if err := value.Deserialize(v); err != nil {
		return nil, err
	}
	return value, nil
}

func (l *Ledger) DeleteGenericType(key *types.GenericKey) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericType, key)
	if err != nil {
		return err
	}
	_, err = l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return ErrGenericTypeNotFound
		}
		return err
	}

	return l.store.Delete(k)
}

func (l *Ledger) UpdateGenericType(key *types.GenericKey, value *types.GenericType) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericType, key)
	if err != nil {
		return err
	}
	_, err = l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return ErrGenericTypeNotFound
		}
		return err
	}

	v, err := value.Serialize()
	if err != nil {
		return err
	}
	return l.store.Put(k, v)
}

func (l *Ledger) AddOrUpdateGenericType(key *types.GenericKey, value *types.GenericType) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericType, key)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}
	return l.store.Put(k, v)
}

func (l *Ledger) GetGenericTypes(fn func(key *types.GenericKey, value *types.GenericType) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixGenericType)

	err := l.store.Iterator(prefix, nil, func(k []byte, v []byte) error {
		key := new(types.GenericKey)
		if err := key.Deserialize(k[1:]); err != nil {
			return err
		}
		value := new(types.GenericType)
		if err := value.Deserialize(v); err != nil {
			return err
		}
		if err := fn(key, value); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) HasGenericType(key *types.GenericKey) (bool, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixGenericType, key)
	if err != nil {
		return false, err
	}
	return l.store.Has(k)
}

func (l *Ledger) CountGenericTypes() (uint64, error) {
	return l.store.Count([]byte{storage.KeyPrefixGenericType})
}
