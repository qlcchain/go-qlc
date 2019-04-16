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

	if err := l.addPovHeader(blk.GetHeader(), txn); err != nil {
		return err
	}

	if err := l.addPovBody(blk.GetHeight(), blk.GetHash(), blk.GetBody(), txn); err != nil {
		return err
	}

	if err := l.addPovHeight(blk.GetHash(), blk.GetHeight(), txn); err != nil {
		return err
	}

	l.releaseTxn(txn, flag)
	return nil
}

func (l *Ledger) DeletePovBlock(blk *types.PovBlock, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)

	if err := l.deletePovHeader(blk.GetHeight(), blk.GetHash(), txn); err != nil {
		return err
	}

	if err := l.deletePovBody(blk.GetHeight(), blk.GetHash(), txn); err != nil {
		return err
	}

	if err := l.deletePovHeight(blk.GetHash(), txn); err != nil {
		return err
	}

	l.releaseTxn(txn, flag)

	return nil
}

func (l *Ledger) addPovHeader(header *types.PovHeader, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovHeader, header.Height, header.Hash)
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

func (l *Ledger) AddPovHeader(header *types.PovHeader, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return l.addPovHeader(header, txn)
}

func (l *Ledger) deletePovHeader(height uint64, hash types.Hash, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovHeader, height, hash)
	if err != nil {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return l.deletePovHeader(height, hash, txn)
}

func (l *Ledger) GetPovHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovHeader, error) {
	key, err := getKeyOfParts(idPrefixPovHeader, height, hash)
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

func (l *Ledger) HasPovHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) bool {
	header, err := l.GetPovHeader(height, hash)
	if err != nil || header == nil {
		return false
	}

	return true
}

func (l *Ledger) addPovBody(height uint64, hash types.Hash, body *types.PovBody, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovBody, height, hash)
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

func (l *Ledger) AddPovBody(height uint64, hash types.Hash, body *types.PovBody, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return l.addPovBody(height, hash, body, txn)
}

func (l *Ledger) deletePovBody(height uint64, hash types.Hash, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovBody, height)
	if err != nil {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovBody(height uint64, hash types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return l.deletePovBody(height, hash, txn)
}

func (l *Ledger) GetPovBody(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovBody, error) {
	key, err := getKeyOfParts(idPrefixPovBody, height, hash)
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

func (l *Ledger) addPovHeight(hash types.Hash, height uint64, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovHeight, hash)
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

func (l *Ledger) AddPovHeight(hash types.Hash, height uint64, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return l.addPovHeight(hash, height, txn)
}

func (l *Ledger) deletePovHeight(hash types.Hash, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovHeight, hash)
	if err != nil {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovHeight(hash types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return l.deletePovHeight(hash, txn)
}

func (l *Ledger) GetPovHeight(hash types.Hash, txns ...db.StoreTxn) (uint64, error) {
	key, err := getKeyOfParts(idPrefixPovHeight, hash)
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

func (l *Ledger) AddPovBestHash(height uint64, hash types.Hash, txns ...db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovBestHash, height)
	if err != nil {
		return err
	}

	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	valBytes := make([]byte, hash.Len())
	err = hash.MarshalBinaryTo(valBytes)
	if err != nil {
		return err
	}

	if err := txn.Set(key, valBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovBestHash(height uint64, txns ...db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovBestHash, height)
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

func (l *Ledger) GetPovBestHash(height uint64, txns ...db.StoreTxn) (types.Hash, error) {
	key, err := getKeyOfParts(idPrefixPovBestHash, height)
	if err != nil {
		return types.ZeroHash, err
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var hash types.Hash
	err = txn.Get(key, func(val []byte, b byte) error {
		err := hash.UnmarshalBinary(val)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return types.ZeroHash, ErrBlockNotFound
		}
		return types.ZeroHash, err
	}
	return hash, nil
}

func (l *Ledger) GetPovBlockByHeightAndHash(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	header, err := l.GetPovHeader(height, hash, txns...)
	if err != nil {
		return nil, err
	}

	body, err := l.GetPovBody(height, hash, txns...)
	if err != nil{
		return nil, err
	}

	blk := types.NewPovBlockWithBody(header, body)
	return blk, nil
}

func (l *Ledger) GetPovBlockByHeight(height uint64, txns ...db.StoreTxn) (*types.PovBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	hash, err := l.GetPovBestHash(height, txns...)
	if err != nil {
		return nil, err
	}

	return l.GetPovBlockByHeightAndHash(height, hash, txn)
}

func (l *Ledger) GetPovBlockByHash(hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	height, err := l.GetPovHeight(hash, txns...)
	if err != nil {
		return nil, err
	}

	return l.GetPovBlockByHeight(height, txns...)
}

func (l *Ledger) GetAllPovBlocks(fn func(*types.PovBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPovHeader, func(key []byte, val []byte, b byte) error {
		header := new(types.PovHeader)
		if err := header.Deserialize(val); err != nil {
			return err
		}

		body, err := l.GetPovBody(header.Height, header.GetHash(), txns...)
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
	var latestVal []byte
	err := txn.Iterator(idPrefixPovBestHash, func(key []byte, val []byte, meta byte) error {
		latestKey = key
		latestVal = val
		return nil
	})
	if err != nil {
		return nil, err
	}
	if latestKey == nil || latestVal == nil {
		return nil, err
	}

	latestHeight := util.BytesToUint64(latestKey)
	if latestHeight < 0 {
		return nil, err
	}
	var latestHash types.Hash
	latestHash.UnmarshalBinary(latestVal)
	if latestHash == types.ZeroHash {
		return nil, err
	}

	return l.GetPovBlockByHeightAndHash(latestHeight, latestHash, txn)
}
