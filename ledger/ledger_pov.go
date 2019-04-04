package ledger

import (
	"encoding/binary"
	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) AddPovBlock(blk *types.PovBlock, txns ...db.StoreTxn) error {
	key, _ := getKeyOfParts(idPrefixPovBlock, blk.GetHeight(), blk.GetHash())
	txn, flag := l.getTxn(true, txns...)

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

	if err := l.addPovBlockHash(blk.GetHash(), blk.GetHeight(), txn); err != nil {
		return err
	}

	l.releaseTxn(txn, flag)
	return nil
}

func (l *Ledger) addPovBlockHash(hash types.Hash, height uint64, txn db.StoreTxn) error {
	key := getKeyOfHash(hash, idPrefixPovBlockHash)

	err := txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	blockBytes := util.Uint64ToBytes(height)

	if err := txn.Set(key, blockBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) AddPovTxLookup(txHash types.Hash, povHeight uint64, txns... db.StoreTxn) error {
	key := getKeyOfHash(txHash, idPrefixPovTxHash)
	txn, flag := l.getTxn(true, txns...)
	l.releaseTxn(txn, flag)

	blockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBytes, povHeight)
	util.Uint64ToBytes(povHeight)

	if err := txn.Set(key, blockBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovBlock(blk *types.PovBlock, txns ...db.StoreTxn) error {
	key, _ := getKeyOfParts(idPrefixPovBlock, blk.Height, blk.GetHash())
	txn, flag := l.getTxn(true, txns...)

	if err := txn.Delete(key); err != nil {
		return err
	}

	if err := l.deletePovBlockHash(blk.GetHash(), txn); err != nil {
		return err
	}

	l.releaseTxn(txn, flag)

	return nil
}

func (l *Ledger) deletePovBlockHash(hash types.Hash, txn db.StoreTxn) error {
	key := getKeyOfHash(hash, idPrefixPovBlock)

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovTxLookup(txHash types.Hash, txns... db.StoreTxn) error {
	key := getKeyOfHash(txHash, idPrefixPovTxHash)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovBlockHeight(hash types.Hash, txns ...db.StoreTxn) (uint64, error) {
	key := getKeyOfHash(hash, idPrefixPovBlockHash)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var height uint64
	err := txn.Get(key, func(val []byte, b byte) error {
		height = binary.BigEndian.Uint64(val)
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return 0, ErrBlockNotFound
		}
		return 0, err
	}
	return height, nil
}

func (l *Ledger) GetPovBlockByHash(hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error) {
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

func (l *Ledger) GetAllPovBlocks(fn func(*types.PovBlock) error, txns ...db.StoreTxn) error {
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

func (l *Ledger) GetLatestPovBlock(txns ...db.StoreTxn) (*types.PovBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var latestKey []byte
	err := txn.KeyIterator(idPrefixPovBlockHash, func(key []byte) error {
		latestKey = key
		return nil
	})

	if err != nil {
		return nil, err
	}

	var latestHash types.Hash
	copy(latestHash[:], latestKey)

	return l.GetPovBlockByHash(latestHash, txns...)
}
