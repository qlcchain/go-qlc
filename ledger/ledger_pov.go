package ledger

import (
	"encoding/binary"
	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) AddPovBlock(blk *types.PovBlock, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)

	if err := l.addPovBlockHeader(blk.ToHeader(), txn); err != nil {
		return err
	}

	if err := l.addPovBlockBody(blk.GetHeight(), blk.ToBody(), txn); err != nil {
		return err
	}

	if err := l.addPovBlockNumber(blk.GetHash(), blk.GetHeight(), txn); err != nil {
		return err
	}

	l.releaseTxn(txn, flag)
	return nil
}

func (l *Ledger) DeletePovBlock(blk *types.PovBlock, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)

	if err := l.deletePovBlockHeader(blk.GetHeight(), blk.GetHash(), txn); err != nil {
		return err
	}

	if err := l.deletePovBlockBody(blk.GetHeight(), txn); err != nil {
		return err
	}

	if err := l.deletePovBlockNumber(blk.GetHash(), txn); err != nil {
		return err
	}

	l.releaseTxn(txn, flag)

	return nil
}

func (l *Ledger) addPovBlockHeader(header *types.PovHeader, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovBlockHeader, header.Height, header.Hash)
	if err != nil {
		return err
	}

	err = txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	dataBytes, err := header.Serialize()
	if err != nil {
		return err
	}

	if err := txn.Set(key, dataBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) deletePovBlockHeader(height uint64, hash types.Hash, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovBlockHeader, height, hash)
	if err != nil {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovBlockHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovHeader, error) {
	key, err := getKeyOfParts(idPrefixPovBlockHeader, height, hash)
	if err != nil {
		return nil, err
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	hdr := new(types.PovHeader)
	err = txn.Get(key, func(val []byte, b byte) error {
		if err := hdr.Deserialize(val); err != nil {
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
	return hdr, nil
}

func (l *Ledger) addPovBlockBody(height uint64, body *types.PovBody, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovBlockBody, height)
	if err != nil {
		return err
	}

	err = txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	dataBytes, err := body.Serialize()
	if err != nil {
		return err
	}

	if err := txn.Set(key, dataBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) deletePovBlockBody(height uint64, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovBlockBody, height)
	if err != nil {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovBlockBody(height uint64, txns ...db.StoreTxn) (*types.PovBody, error) {
	key, err := getKeyOfParts(idPrefixPovBlockBody, height)
	if err != nil {
		return nil, err
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	body := new(types.PovBody)
	err = txn.Get(key, func(val []byte, b byte) error {
		if err := body.Deserialize(val); err != nil {
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
	return body, nil
}

func (l *Ledger) addPovBlockNumber(hash types.Hash, height uint64, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovBlockNumber, hash)
	if err != nil {
		return err
	}

	err = txn.Get(key, func(bytes []byte, b byte) error {
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

func (l *Ledger) deletePovBlockNumber(hash types.Hash, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovBlockNumber, hash)
	if err != nil {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovBlockNumber(hash types.Hash, txns ...db.StoreTxn) (uint64, error) {
	key, err := getKeyOfParts(idPrefixPovBlockNumber, hash)
	if err != nil {
		return 0, err
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var height uint64
	err = txn.Get(key, func(val []byte, b byte) error {
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

func (l *Ledger) AddPovTxLookup(txHash types.Hash, txLookup *types.PovTxLookup, txns... db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovTxLookup, txHash)
	if err != nil {
		return err
	}

	txn, flag := l.getTxn(true, txns...)
	l.releaseTxn(txn, flag)

	dataTypes, err := txLookup.Serialize()
	if err != nil {
		return err
	}

	if err := txn.Set(key, dataTypes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovTxLookup(txHash types.Hash, txns... db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovTxLookup, txHash)
	if err != nil {
		return err
	}

	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovBlockByHeight(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	header, err := l.GetPovBlockHeader(height, hash, txns...)
	if err != nil {
		return nil, err
	}

	body, err := l.GetPovBlockBody(height, txns...)
	if err != nil{
		return nil, err
	}

	blk := types.NewPovBlockWithBody(header, body)
	return blk, nil
}

func (l *Ledger) GetPovBlockByHash(hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	height, err := l.GetPovBlockNumber(hash, txns...)
	if err != nil {
		return nil, err
	}

	return l.GetPovBlockByHeight(height, hash, txns...)
}

func (l *Ledger) GetAllPovBlocks(fn func(*types.PovBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPovBlockHeader, func(key []byte, val []byte, b byte) error {
		header := new(types.PovHeader)
		if err := header.Deserialize(val); err != nil {
			return err
		}

		body, err := l.GetPovBlockBody(header.Height, txns...)
		if err != nil {
			return err
		}

		blk := types.NewPovBlockWithBody(header, body)
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
	err := txn.KeyIterator(idPrefixPovBlockNumber, func(key []byte) error {
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
