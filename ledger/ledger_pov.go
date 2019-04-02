package ledger

import (
	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) AddPovBlock(blk *types.PovBlock, txns ...db.StoreTxn) error {
	key := getKeyOfHash(blk.GetHash(), idPrefixPovBlock)
	txn, flag := l.getTxn(true, txns...)

	//never overwrite implicitly
	err := txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	blockBytes, err := blk.Serialize()
	if err != nil {
		return err
	}
	if err := txn.Set(key, blockBytes); err != nil {
		return err
	}

	l.releaseTxn(txn, flag)
	return nil
}

func (l *Ledger) GetPovBlock(hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error) {
	key := getKeyOfHash(hash, idPrefixPovBlock)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	blk := new(types.PovBlock)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := blk.Deserialize(val); err != nil {
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

func (l *Ledger) GetPovBlocks(fn func(*types.PovBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPovBlock, func(key []byte, val []byte, b byte) error {
		blk := new(types.PovBlock)
		if err := blk.Deserialize(val); err != nil {
			return err
		}
		if err := fn(blk); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) DeletePovBlock(hash types.Hash, txns ...db.StoreTxn) error {
	key := getKeyOfHash(hash, idPrefixPovBlock)
	txn, flag := l.getTxn(true, txns...)

	blk := new(types.StateBlock)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := blk.Deserialize(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil && err != ErrBlockNotFound {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}

	l.releaseTxn(txn, flag)
	return nil
}
