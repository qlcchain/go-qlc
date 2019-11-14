package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"
	"time"
)

func (l *Ledger) AddVoteHistory(hash types.Hash, address types.Address, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixVoteHistory, hash, address)
	if err != nil {
		return err
	}

	return txn.Set(k, util.BE_Uint64ToBytes(uint64(time.Now().Unix())))
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

func (l *Ledger) CleanVoteHistoryTimeout(txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	now := uint64(time.Now().Unix())
	err := txn.Iterator(idPrefixVoteHistory, func(key []byte, val []byte, b byte) error {
		addTime := util.BE_BytesToUint64(val)
		if now - addTime > 600 {
			er := txn.Delete(key)
			if er != nil {
				l.logger.Error(er)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}