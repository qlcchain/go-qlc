package ledger

import (
	"encoding/json"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) AddOrUpdatePerformance(value *types.PerformanceTime, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPerformance, value.Hash)
	if err != nil {
		return err
	}
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return txn.Set(k, v)
}

func (l *Ledger) PerformanceTimes(fn func(*types.PerformanceTime), txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPerformance, func(key []byte, val []byte, b byte) error {
		pt := new(types.PerformanceTime)
		err := json.Unmarshal(val, pt)
		if err != nil {
			return err
		}
		fn(pt)
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetPerformanceTime(key types.Hash, txns ...db.StoreTxn) (*types.PerformanceTime, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPerformance, key)
	if err != nil {
		return nil, err
	}

	value := types.NewPerformanceTime()
	err = txn.Get(k, func(val []byte, b byte) (err error) {
		return json.Unmarshal(val, &value)
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPerformanceNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) IsPerformanceTimeExist(key types.Hash, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	if _, err := l.GetPerformanceTime(key); err == nil {
		return true, nil
	} else {
		if err == ErrPerformanceNotFound {
			return false, nil
		} else {
			return false, err
		}
	}
}
