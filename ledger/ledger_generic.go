// Note: This file generate CURD code by genny automatically (https://github.com/cheekybits/genny)
// Usage:
// - install gen, command is  ` go get github.com/cheekybits/genny
// - create type in /common/types package, example: StateBlock, Hash
// - generate code, command is ` cat ledger_generic.go | genny gen "GenericType=StateBlock GenericKey=Hash" > ledger_generic_gen.go `

package ledger

import (
	"errors"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

const idPrefixGenericType byte = iota

var (
	ErrGenericTypeExists   = errors.New("the GenericType is empty")
	ErrGenericTypeNotFound = errors.New("GenericType not found")
)

func (l *Ledger) AddGenericType(key types.GenericKey, value *types.GenericType, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixGenericType, key)
	if err != nil {
		return nil
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrGenericTypeExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetGenericType(key types.GenericKey, txns ...db.StoreTxn) (*types.GenericType, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixGenericType, key)
	if err != nil {
		return nil, err
	}

	value := new(types.GenericType)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrGenericTypeNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) DeleteGenericType(key types.GenericKey, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixGenericType, key)
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err != nil {
		return err
	}

	return txn.Delete(k)
}

func (l *Ledger) UpdateGenericType(key types.GenericKey, value *types.GenericType, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixGenericType, key)
	if err != nil {
		return nil
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrGenericTypeNotFound
		}
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) AddOrUpdateGenericType(key types.GenericKey, value *types.GenericType, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixGenericType, key)
	if err != nil {
		return nil
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetGenericTypes(fn func(key types.GenericKey, value *types.GenericType) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixGenericType, func(k []byte, v []byte, b byte) error {
		key := new(types.GenericKey)
		if err := key.Deserialize(k); err != nil {
			return err
		}
		value := new(types.GenericType)
		if err := value.Deserialize(v); err != nil {
			return err
		}
		if err := fn(*key, value); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) HasGenericType(key types.GenericKey, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixGenericType, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) CountGenericTypes(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	return txn.Count([]byte{idPrefixGenericType})
}
