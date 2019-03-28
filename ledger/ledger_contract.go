package ledger

import (
	"bytes"
	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func getStorageKey(prefix, key []byte) []byte {
	var storageKey []byte
	storageKey = append(storageKey, []byte{idPrefixStorage}...)
	storageKey = append(storageKey, prefix...)
	storageKey = append(storageKey, key...)
	return storageKey
}

func (l *Ledger) GetStorage(prefix, key []byte, txns ...db.StoreTxn) ([]byte, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	storageKey := getStorageKey(prefix, key)
	var storage []byte
	err := txn.Get(storageKey, func(val []byte, b byte) (err error) {
		storage = val
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}
	return storage, nil
}

func (l *Ledger) SetStorage(prefix, key []byte, value []byte, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)
	storageKey := getStorageKey(prefix, key)
	err := txn.Get(storageKey, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrStorageExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(storageKey, value)
}

func (l *Ledger) Iterator(prefix []byte, fn func(key []byte, value []byte) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixStorage, func(key []byte, val []byte, b byte) error {
		if bytes.HasPrefix(key, prefix) {
			err := fn(key, val)
			if err != nil {
				l.logger.Error(err)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}


