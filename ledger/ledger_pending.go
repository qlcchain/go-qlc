package ledger

import (
	"errors"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) AddPending(key *types.PendingKey, value *types.PendingInfo, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPending, key)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrPendingExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}

	if err := txn.SetWithMeta(k, v, byte(types.PendingNotUsed)); err != nil {
		return err
	}
	return l.cache.UpdateAccountPending(key, value, true)
	//return txn.Set(key, pendingBytes)
}

func (l *Ledger) UpdatePending(key *types.PendingKey, kind types.PendingKind, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPending, key)
	if err != nil {
		return err
	}
	var pending types.PendingInfo
	err = txn.Get(k[:], func(v []byte, b byte) (err error) {
		if err := pending.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil
		}
		return err
	}

	v, err := pending.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return txn.SetWithMeta(k, v, byte(kind))
}

func (l *Ledger) GetPending(key *types.PendingKey, txns ...db.StoreTxn) (*types.PendingInfo, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPending, key)
	if err != nil {
		return nil, err
	}

	value := new(types.PendingInfo)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPendingNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetPendings(fn func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	err := txn.Iterator(idPrefixPending, func(key []byte, val []byte, b byte) error {
		pendingKey := new(types.PendingKey)
		if err := pendingKey.Deserialize(key[1:]); err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		pendingInfo := new(types.PendingInfo)
		if err := pendingInfo.Deserialize(val); err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		used := types.PendingKind(b)
		if used == types.PendingNotUsed {
			if err := fn(pendingKey, pendingInfo); err != nil {
				l.logger.Error("process pending error %s", err)
				errStr = append(errStr, err.Error())
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}

func (l *Ledger) SearchPending(address types.Address, fn func(key *types.PendingKey, value *types.PendingInfo) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Stream([]byte{idPrefixPending}, func(item *badger.Item) bool {
		key := &types.PendingKey{}
		if err := key.Deserialize(item.Key()[1:]); err == nil {
			return key.Address == address
		} else {
			l.logger.Error(util.ToString(key), err)
		}

		return false
	}, func(list *pb.KVList) error {
		for _, v := range list.Kv {
			pk := &types.PendingKey{}
			pi := &types.PendingInfo{}
			used := types.PendingKind(v.UserMeta[0])

			if used == types.PendingNotUsed {
				if err := pk.Deserialize(v.Key[1:]); err != nil {
					continue
				}
				if err := pi.Deserialize(v.Value); err != nil {
					continue
				}
				err := fn(pk, pi)
				if err != nil {
					l.logger.Error(err)
				}
			}
		}
		return nil
	})
}

func (l *Ledger) SearchAllKindPending(address types.Address, fn func(key *types.PendingKey, value *types.PendingInfo, kind types.PendingKind) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Stream([]byte{idPrefixPending}, func(item *badger.Item) bool {
		key := &types.PendingKey{}
		if err := key.Deserialize(item.Key()[1:]); err == nil {
			return key.Address == address
		} else {
			l.logger.Error(util.ToString(key), err)
		}
		return false
	}, func(list *pb.KVList) error {
		for _, v := range list.Kv {
			pk := &types.PendingKey{}
			pi := &types.PendingInfo{}
			used := types.PendingKind(v.UserMeta[0])
			if err := pk.Deserialize(v.Key[1:]); err != nil {
				continue
			}
			if err := pi.Deserialize(v.Value); err != nil {
				continue
			}
			err := fn(pk, pi, used)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (l *Ledger) DeletePending(key *types.PendingKey, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPending, key)
	if err != nil {
		return err
	}

	value := new(types.PendingInfo)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		if err := txn.Delete(k); err != nil {
			return err
		}
		return l.cache.UpdateAccountPending(key, value, true)
	}
	return nil
}

func (l *Ledger) Pending(account types.Address, txns ...db.StoreTxn) ([]*types.PendingKey, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var cache []*types.PendingKey
	err := txn.Iterator(idPrefixPending, func(key []byte, val []byte, b byte) error {
		pendingKey := types.PendingKey{}
		_, err := pendingKey.UnmarshalMsg(key[1:])
		if err != nil {
			return err
		}
		if pendingKey.Address == account {
			cache = append(cache, &pendingKey)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return cache, nil
}

func (l *Ledger) PendingAmount(address types.Address, token types.Hash, txns ...db.StoreTxn) (types.Balance, error) {
	b, err := l.cache.GetAccountPending(address, token)
	if err == nil {
		return b, nil
	}
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	pendingKeys, err := l.TokenPending(address, token)
	if err != nil {
		return types.ZeroBalance, err
	}
	pendingAmount := types.ZeroBalance
	for _, key := range pendingKeys {
		pendinginfo, err := l.GetPending(key)
		if err != nil {
			return types.ZeroBalance, err
		}
		pendingAmount = pendingAmount.Add(pendinginfo.Amount)
	}
	if err := l.cache.AddAccountPending(address, token, pendingAmount); err != nil {
		return types.ZeroBalance, err
	}
	return pendingAmount, nil
}

func (l *Ledger) TokenPending(account types.Address, token types.Hash, txns ...db.StoreTxn) ([]*types.PendingKey, error) {
	var cache []*types.PendingKey
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPending, func(key []byte, val []byte, b byte) error {
		pendingKey := types.PendingKey{}
		_, err := pendingKey.UnmarshalMsg(key[1:])
		if err != nil {
			return nil
		}
		if pendingKey.Address == account {
			var pending types.PendingInfo
			_, err := pending.UnmarshalMsg(val)
			if err != nil {
				return err
			}
			if pending.Type == token {
				cache = append(cache, &pendingKey)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return cache, nil
}

func (l *Ledger) TokenPendingInfo(account types.Address, token types.Hash, txns ...db.StoreTxn) ([]*types.PendingInfo, error) {
	var cache []*types.PendingInfo
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPending, func(key []byte, val []byte, b byte) error {
		pendingKey := types.PendingKey{}
		_, err := pendingKey.UnmarshalMsg(key[1:])
		if err != nil {
			return nil
		}
		if pendingKey.Address == account {
			var pendingInfo types.PendingInfo
			_, err := pendingInfo.UnmarshalMsg(val)
			if err != nil {
				return err
			}
			if pendingInfo.Type == token {
				cache = append(cache, &pendingInfo)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return cache, nil
}
