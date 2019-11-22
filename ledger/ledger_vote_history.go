package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) AddVoteHistory(hash types.Hash, address types.Address, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixVoteHistory, hash, address)
	if err != nil {
		return err
	}

	return txn.Set(k, nil)
}

func (l *Ledger) HasVoteHistory(hash types.Hash, address types.Address, txns ...db.StoreTxn) bool {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixVoteHistory, hash, address)
	if err != nil {
		return false
	}

	if err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	}); err != nil {
		return false
	}

	return true
}

func (l *Ledger) CleanBlockVoteHistory(hash types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixVoteHistory, hash)
	if err != nil {
		return err
	}

	err = txn.PrefixIterator(k, func(key []byte, val []byte, b byte) error {
		k := make([]byte, len(key))
		copy(k[:], key)
		if er := txn.Delete(k); er != nil {
			l.logger.Error(er)
		}
		return nil
	})

	return nil
}

func (l *Ledger) CleanAllVoteHistory(txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixVoteHistory)
	if err != nil {
		return err
	}

	err = txn.PrefixIterator(k, func(key []byte, val []byte, b byte) error {
		k := make([]byte, len(key))
		copy(k[:], key)
		if er := txn.Delete(k); er != nil {
			l.logger.Error(er)
		}
		return nil
	})

	return nil
}
