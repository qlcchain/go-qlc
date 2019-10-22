package ledger

import (
	"github.com/dgraph-io/badger"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) getUncheckedSyncBlockKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixUncheckedSync
	copy(key[1:], hash[:])
	return key[:]
}

func (l *Ledger) AddUncheckedSyncBlock(previous types.Hash, blk *types.StateBlock, txns ...db.StoreTxn) error {
	blockBytes, err := blk.Serialize()
	if err != nil {
		return err
	}

	key := l.getUncheckedSyncBlockKey(previous)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Set(key, blockBytes)
}

func (l *Ledger) GetUncheckedSyncBlock(hash types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	key := l.getUncheckedSyncBlockKey(hash)

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	blk := new(types.StateBlock)
	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if err = blk.Deserialize(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}

	return blk, nil
}

func (l *Ledger) HasUncheckedSyncBlock(hash types.Hash, txns ...db.StoreTxn) (bool, error) {
	key := l.getUncheckedSyncBlockKey(hash)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) error {
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

func (l *Ledger) CountUncheckedSyncBlocks(txns ...db.StoreTxn) (uint64, error) {
	var count uint64
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	count, err := txn.Count([]byte{idPrefixUncheckedSync})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (l *Ledger) DeleteUncheckedSyncBlock(hash types.Hash, txns ...db.StoreTxn) error {
	key := l.getUncheckedSyncBlockKey(hash)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Delete(key)
}

func (l *Ledger) getUnconfirmedSyncBlockKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixUnconfirmedSync
	copy(key[1:], hash[:])
	return key[:]
}

func (l *Ledger) AddUnconfirmedSyncBlock(hash types.Hash, blk *types.StateBlock, txns ...db.StoreTxn) error {
	blockBytes, err := blk.Serialize()
	if err != nil {
		return err
	}

	key := l.getUnconfirmedSyncBlockKey(hash)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Set(key, blockBytes)
}

func (l *Ledger) GetUnconfirmedSyncBlock(hash types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	key := l.getUnconfirmedSyncBlockKey(hash)

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	blk := new(types.StateBlock)
	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if err = blk.Deserialize(val); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return blk, nil
}

func (l *Ledger) HasUnconfirmedSyncBlock(hash types.Hash, txns ...db.StoreTxn) (bool, error) {
	key := l.getUnconfirmedSyncBlockKey(hash)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) error {
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

func (l *Ledger) CountUnconfirmedSyncBlocks(txns ...db.StoreTxn) (uint64, error) {
	var count uint64
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	count, err := txn.Count([]byte{idPrefixUnconfirmedSync})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (l *Ledger) DeleteUnconfirmedSyncBlock(hash types.Hash, txns ...db.StoreTxn) error {
	key := l.getUnconfirmedSyncBlockKey(hash)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Delete(key)
}

func (l *Ledger) WalkSyncCache(visit common.SyncCacheWalkFunc, txns ...db.StoreTxn) {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	_ = txn.Iterator(idPrefixUnconfirmedSync, func(key []byte, val []byte, b byte) error {
		visit(common.SyncCacheUnconfirmed, key)
		return nil
	})

	_ = txn.Iterator(idPrefixUncheckedSync, func(key []byte, val []byte, b byte) error {
		visit(common.SyncCacheUnchecked, key)
		return nil
	})
}
