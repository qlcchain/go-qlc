package ledger

import (
	"encoding/json"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) AddRepresentation(key types.Address, diff *types.Benefit, txns ...db.StoreTxn) error {
	spin := l.representCache.getLock(key.String())
	spin.Lock()
	defer spin.Unlock()

	value, err := l.GetRepresentation(key, txns...)
	if err != nil && err != ErrRepresentationNotFound {
		l.logger.Errorf("getRepresentation error: %s ,address: %s", err, key)
		return err
	}

	value.Balance = value.Balance.Add(diff.Balance)
	value.Vote = value.Vote.Add(diff.Vote)
	value.Network = value.Network.Add(diff.Network)
	value.Oracle = value.Oracle.Add(diff.Oracle)
	value.Storage = value.Storage.Add(diff.Storage)
	value.Total = value.Total.Add(diff.Total)

	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)
	return l.representCache.updateMemory(key, value, txn)
}

func (l *Ledger) SubRepresentation(key types.Address, diff *types.Benefit, txns ...db.StoreTxn) error {
	spin := l.representCache.getLock(key.String())
	spin.Lock()
	defer spin.Unlock()

	value, err := l.GetRepresentation(key, txns...)
	if err != nil {
		l.logger.Errorf("GetRepresentation error: %s ,address: %s", err, key)
		return err
	}
	value.Balance = value.Balance.Sub(diff.Balance)
	value.Vote = value.Vote.Sub(diff.Vote)
	value.Network = value.Network.Sub(diff.Network)
	value.Oracle = value.Oracle.Sub(diff.Oracle)
	value.Storage = value.Storage.Sub(diff.Storage)
	value.Total = value.Total.Sub(diff.Total)

	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)
	return l.representCache.updateMemory(key, value, txn)
}

func (l *Ledger) getRepresentation(key types.Address, txns ...db.StoreTxn) (*types.Benefit, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixRepresentation, key)
	if err != nil {
		return nil, err
	}

	value := new(types.Benefit)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			l.logger.Errorf("Unmarshal benefit error: %s ,address: %s, val: %s", err, key, string(v))
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrRepresentationNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetRepresentation(key types.Address, txns ...db.StoreTxn) (*types.Benefit, error) {
	if lr, ok := l.representCache.getFromMemory(key.String()); ok {
		return lr.(*types.Benefit), nil
	}
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixRepresentation, key)
	if err != nil {
		return nil, err
	}

	value := new(types.Benefit)
	err = txn.Get(k, func(v []byte, b byte) (err error) {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return &types.Benefit{
				Vote:    types.ZeroBalance,
				Network: types.ZeroBalance,
				Storage: types.ZeroBalance,
				Oracle:  types.ZeroBalance,
				Balance: types.ZeroBalance,
				Total:   types.ZeroBalance,
			}, ErrRepresentationNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetRepresentations(fn func(types.Address, *types.Benefit) error, txns ...db.StoreTxn) error {
	err := l.representCache.iterMemory(func(s string, i interface{}) error {
		address, err := types.HexToAddress(s)
		if err != nil {
			return err
		}
		benefit := i.(*types.Benefit)
		if err := fn(address, benefit); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	err = txn.Iterator(idPrefixRepresentation, func(key []byte, val []byte, b byte) error {
		address, err := types.BytesToAddress(key[1:])
		if err != nil {
			return err
		}
		if _, ok := l.representCache.getFromMemory(address.String()); !ok {
			benefit := new(types.Benefit)
			if err = benefit.Deserialize(val); err != nil {
				l.logger.Errorf("Unmarshal benefit error: %s ,val: %s", err, string(val))
				return err
			}
			if err := fn(address, benefit); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetRepresentationsCache(address types.Address, fn func(address types.Address, am *types.Benefit, amCache *types.Benefit) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	if !address.IsZero() {
		be, err := l.getRepresentation(address)
		if err != nil && err != ErrRepresentationNotFound {
			return err
		}
		if v, ok := l.representCache.getFromMemory(address.String()); ok {
			beMemory := v.(*types.Benefit)
			if err := fn(address, be, beMemory); err != nil {
				return err
			}
		} else {
			if err := fn(address, be, nil); err != nil {
				return err
			}
		}
		return nil
	} else {
		err := l.representCache.iterMemory(func(s string, i interface{}) error {
			addr, err := types.HexToAddress(s)
			if err != nil {
				return err
			}
			beMemory := i.(*types.Benefit)
			be, err := l.getRepresentation(addr)
			if err != nil && err != ErrRepresentationNotFound {
				return err
			}
			if err := fn(addr, be, beMemory); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		err = txn.Iterator(idPrefixRepresentation, func(key []byte, val []byte, b byte) error {
			addr, err := types.BytesToAddress(key[1:])
			if err != nil {
				l.logger.Error(err)
				return err
			}
			be := new(types.Benefit)
			_, err = be.UnmarshalMsg(val)
			if err != nil {
				return err
			}
			if _, ok := l.representCache.getFromMemory(addr.String()); !ok {
				if err := fn(addr, be, nil); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
		return nil
	}
}

func (l *Ledger) GetOnlineRepresentations(txns ...db.StoreTxn) ([]types.Address, error) {
	key := []byte{idPrefixOnlineReps}
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)
	var result []types.Address
	err := txn.Get(key, func(bytes []byte, b byte) error {
		if len(bytes) > 0 {
			return json.Unmarshal(bytes, &result)
		}
		return nil
	})

	return result, err
}

func (l *Ledger) SetOnlineRepresentations(addresses []*types.Address, txns ...db.StoreTxn) error {
	bytes, err := json.Marshal(addresses)
	if err != nil {
		return err
	}

	key := []byte{idPrefixOnlineReps}
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Set(key, bytes)
}
