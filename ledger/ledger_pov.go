package ledger

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"
)

var (
	ErrPovKeyNotFound      = errors.New("pov key not found")
	ErrPovHeaderNotFound   = errors.New("pov header not found")
	ErrPovBodyNotFound     = errors.New("pov body not found")
	ErrPovHeightNotFound   = errors.New("pov height not found")
	ErrPovHashNotFound     = errors.New("pov hash not found")
	ErrPovTxLookupNotFound = errors.New("pov tx lookup not found")
	ErrPovTDNotFound       = errors.New("pov total difficulty not found")

	ErrPovMinerStatNotFound = errors.New("pov miner state not found")
)

func (l *Ledger) AddPovBlock(blk *types.PovBlock, td *big.Int, txns ...db.StoreTxn) error {
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

	if err := l.addPovTD(blk.GetHash(), blk.GetHeight(), td, txn); err != nil {
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

	if err := l.deletePovTD(blk.GetHash(), blk.GetHeight(), txn); err != nil {
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
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovHeader, height, hash)
	if err != nil {
		return nil, err
	}

	hdr := new(types.PovHeader)
	err = txn.Get(key, func(val []byte, b byte) error {
		if err := hdr.Deserialize(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPovHeaderNotFound
		}
		return nil, err
	}
	return hdr, nil
}

func (l *Ledger) HasPovHeader(height uint64, hash types.Hash, txns ...db.StoreTxn) bool {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovHeader, height, hash)
	if err != nil {
		return false
	}

	err = txn.Get(key, func(val []byte, b byte) error {
		return nil
	})
	if err != nil {
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
	key, err := getKeyOfParts(idPrefixPovBody, height, hash)
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
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovBody, height, hash)
	if err != nil {
		return nil, err
	}

	body := new(types.PovBody)
	err = txn.Get(key, func(val []byte, b byte) error {
		if err := body.Deserialize(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPovBodyNotFound
		}
		return nil, err
	}
	return body, nil
}

func (l *Ledger) HasPovBody(height uint64, hash types.Hash, txns ...db.StoreTxn) bool {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovBody, height, hash)
	if err != nil {
		return false
	}

	err = txn.Get(key, func(val []byte, b byte) error {
		return nil
	})
	if err != nil {
		return false
	}

	return true
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

	blockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBytes, height)

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
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovHeight, hash)
	if err != nil {
		return 0, err
	}

	var height uint64
	err = txn.Get(key, func(val []byte, b byte) error {
		height = binary.BigEndian.Uint64(val)
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return 0, ErrPovHeightNotFound
		}
		return 0, err
	}
	return height, nil
}

func (l *Ledger) HasPovHeight(hash types.Hash, txns ...db.StoreTxn) bool {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovHeight, hash)
	if err != nil {
		return false
	}

	err = txn.Get(key, func(val []byte, b byte) error {
		return nil
	})
	if err != nil {
		return false
	}

	return true
}

func (l *Ledger) addPovTD(hash types.Hash, height uint64, td *big.Int, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovTD, height, hash)
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

	if err := txn.Set(key, td.Bytes()); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) AddPovTD(hash types.Hash, height uint64, td *big.Int, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return l.addPovTD(hash, height, td, txn)
}

func (l *Ledger) deletePovTD(hash types.Hash, height uint64, txn db.StoreTxn) error {
	key, err := getKeyOfParts(idPrefixPovTD, height, hash)
	if err != nil {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovTD(hash types.Hash, height uint64, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return l.deletePovTD(hash, height, txn)
}

func (l *Ledger) GetPovTD(hash types.Hash, height uint64, txns ...db.StoreTxn) (*big.Int, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovTD, height, hash)
	if err != nil {
		return nil, err
	}

	td := new(big.Int)
	err = txn.Get(key, func(val []byte, b byte) error {
		td.SetBytes(val)
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPovTDNotFound
		}
		return nil, err
	}
	return td, nil
}

func (l *Ledger) AddPovTxLookup(txHash types.Hash, txLookup *types.PovTxLookup, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovTxLookup, txHash)
	if err != nil {
		return err
	}

	dataTypes, err := txLookup.Serialize()
	if err != nil {
		return err
	}

	if err := txn.Set(key, dataTypes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovTxLookup(txHash types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovTxLookup, txHash)
	if err != nil {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovTxLookup(txHash types.Hash, txns ...db.StoreTxn) (*types.PovTxLookup, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovTxLookup, txHash)
	if err != nil {
		return nil, err
	}

	txLookup := new(types.PovTxLookup)
	err = txn.Get(key, func(val []byte, b byte) error {
		if err := txLookup.Deserialize(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPovTxLookupNotFound
		}
		return nil, err
	}
	return txLookup, nil
}

func (l *Ledger) HasPovTxLookup(txHash types.Hash, txns ...db.StoreTxn) bool {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovTxLookup, txHash)
	if err != nil {
		return false
	}

	err = txn.Get(key, func(val []byte, b byte) error {
		return nil
	})
	if err != nil {
		return false
	}

	return true
}

func (l *Ledger) AddPovBestHash(height uint64, hash types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovBestHash, height)
	if err != nil {
		return err
	}

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
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovBestHash, height)
	if err != nil {
		return err
	}

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

	return l.getPovBestHash(key, txn)
}

func (l *Ledger) getPovBestHash(key []byte, txn db.StoreTxn) (types.Hash, error) {
	var hash types.Hash
	err := txn.Get(key, func(val []byte, b byte) error {
		err := hash.UnmarshalBinary(val)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return types.ZeroHash, ErrPovHashNotFound
		}
		return types.ZeroHash, err
	}
	return hash, nil
}

func (l *Ledger) GetAllPovBestHashes(fn func(height uint64, hash types.Hash) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPovBestHash, func(key []byte, val []byte, b byte) error {
		var height uint64
		height = util.BE_BytesToUint64(key[1:])

		var hash types.Hash
		if err := hash.UnmarshalBinary(val); err != nil {
			return err
		}

		if err := fn(height, hash); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) AddPovMinerStat(dayStat *types.PovMinerDayStat, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovMinerStat, dayStat.DayIndex)
	if err != nil {
		return err
	}

	valBytes, err := dayStat.Serialize()
	if err != nil {
		return err
	}

	if err := txn.Set(key, valBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovMinerStat(dayIndex uint32, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPovMinerStat, dayIndex)
	if err != nil {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) getPovMinerStat(key []byte, txn db.StoreTxn) (*types.PovMinerDayStat, error) {
	dayStat := new(types.PovMinerDayStat)
	err := txn.Get(key, func(val []byte, b byte) error {
		err := dayStat.Deserialize(val)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPovMinerStatNotFound
		}
		return nil, err
	}
	return dayStat, nil
}

func (l *Ledger) GetPovMinerStat(dayIndex uint32, txns ...db.StoreTxn) (*types.PovMinerDayStat, error) {
	key, err := getKeyOfParts(idPrefixPovMinerStat, dayIndex)
	if err != nil {
		return nil, err
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return l.getPovMinerStat(key, txn)
}

func (l *Ledger) GetLatestPovMinerStat(txns ...db.StoreTxn) (*types.PovMinerDayStat, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var latestKey []byte
	err := txn.Iterator(idPrefixPovMinerStat, func(key []byte, val []byte, meta byte) error {
		// just safe copy, iterator will reuse item(key, value)
		latestKey = append(latestKey[:0], key...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(latestKey) <= 0 {
		return nil, err
	}

	return l.getPovMinerStat(latestKey, txn)
}

func (l *Ledger) GetAllPovMinerStats(fn func(*types.PovMinerDayStat) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPovMinerStat, func(key []byte, val []byte, meta byte) error {
		dayStat := new(types.PovMinerDayStat)
		err := dayStat.Deserialize(val)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovBlockByHeightAndHash(height uint64, hash types.Hash, txns ...db.StoreTxn) (*types.PovBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	header, err := l.GetPovHeader(height, hash, txns...)
	if err != nil {
		return nil, err
	}

	body, err := l.GetPovBody(height, hash, txns...)
	if err != nil {
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

	return l.GetPovBlockByHeightAndHash(height, hash, txns...)
}

func (l *Ledger) GetPovHeaderByHeight(height uint64, txns ...db.StoreTxn) (*types.PovHeader, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	hash, err := l.GetPovBestHash(height, txns...)
	if err != nil {
		return nil, err
	}

	return l.GetPovHeader(height, hash, txn)
}

func (l *Ledger) BatchGetPovHeadersByHeightAsc(height uint64, count uint64, txns ...db.StoreTxn) ([]*types.PovHeader, error) {
	startKey, err := getKeyOfParts(idPrefixPovBestHash, height)
	if err != nil {
		return nil, err
	}
	endKey, err := getKeyOfParts(idPrefixPovBestHash, uint64(height+count))
	if err != nil {
		return nil, err
	}

	return l.batchGetPovHeadersByHeight(startKey, endKey, txns...)
}

func (l *Ledger) BatchGetPovHeadersByHeightDesc(height uint64, count uint64, txns ...db.StoreTxn) ([]*types.PovHeader, error) {
	if height < count {
		return nil, errors.New("height should greater than count")
	}

	startKey, err := getKeyOfParts(idPrefixPovBestHash, uint64(height-count+1))
	if err != nil {
		return nil, err
	}
	endKey, err := getKeyOfParts(idPrefixPovBestHash, height+1)
	if err != nil {
		return nil, err
	}

	headers, err := l.batchGetPovHeadersByHeight(startKey, endKey, txns...)
	if err != nil {
		return nil, err
	}

	// reversing
	for i := len(headers)/2 - 1; i >= 0; i-- {
		opp := len(headers) - 1 - i
		headers[i], headers[opp] = headers[opp], headers[i]
	}

	return headers, nil
}

func (l *Ledger) batchGetPovHeadersByHeight(startKey []byte, endKey []byte, txns ...db.StoreTxn) ([]*types.PovHeader, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var headers []*types.PovHeader

	err := txn.RangeIterator(startKey, endKey, func(key []byte, val []byte, meta byte) error {
		var height uint64
		height = util.BE_BytesToUint64(key[1:])

		var hash types.Hash
		err := hash.UnmarshalBinary(val)
		if err != nil {
			return err
		}

		header, err := l.GetPovHeader(height, hash)
		if err != nil {
			return err
		}

		headers = append(headers, header)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return headers, nil
}

func (l *Ledger) GetPovHeaderByHash(hash types.Hash, txns ...db.StoreTxn) (*types.PovHeader, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	height, err := l.GetPovHeight(hash, txns...)
	if err != nil {
		return nil, err
	}

	return l.GetPovHeader(height, hash, txns...)
}

func (l *Ledger) GetAllPovHeaders(fn func(header *types.PovHeader) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPovHeader, func(key []byte, val []byte, b byte) error {
		header := new(types.PovHeader)
		if err := header.Deserialize(val); err != nil {
			return err
		}
		if err := fn(header); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
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

func (l *Ledger) GetAllPovBestHeaders(fn func(header *types.PovHeader) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := l.GetAllPovBestHashes(func(height uint64, hash types.Hash) error {
		header, err := l.GetPovHeader(height, hash)
		if err != nil {
			return err
		}

		if err := fn(header); err != nil {
			return err
		}

		return nil
	}, txn)

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetAllPovBestBlocks(fn func(*types.PovBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := l.GetAllPovBestHashes(func(height uint64, hash types.Hash) error {
		block, err := l.GetPovBlockByHeightAndHash(height, hash)
		if err != nil {
			return err
		}

		if err := fn(block); err != nil {
			return err
		}

		return nil
	}, txn)

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetLatestPovBestHash(txns ...db.StoreTxn) (types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var latestKey []byte
	err := txn.Iterator(idPrefixPovBestHash, func(key []byte, val []byte, meta byte) error {
		// just safe copy, iterator will reuse item(key, value)
		latestKey = append(latestKey[:0], key...)
		return nil
	})
	if err != nil {
		return types.ZeroHash, err
	}
	if len(latestKey) <= 0 {
		return types.ZeroHash, fmt.Errorf("latest best hash key is zero")
	}

	latestHash, err := l.getPovBestHash(latestKey, txn)
	if err != nil {
		return types.ZeroHash, err
	}
	if latestHash.IsZero() {
		return types.ZeroHash, fmt.Errorf("latest best hash value is zero")
	}

	return latestHash, nil
}

func (l *Ledger) GetLatestPovHeader(txns ...db.StoreTxn) (*types.PovHeader, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	latestHash, err := l.GetLatestPovBestHash(txn)
	if err != nil {
		return nil, err
	}

	height, err := l.GetPovHeight(latestHash, txns...)
	if err != nil {
		return nil, err
	}

	return l.GetPovHeader(height, latestHash, txn)
}

func (l *Ledger) GetLatestPovBlock(txns ...db.StoreTxn) (*types.PovBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	latestHash, err := l.GetLatestPovBestHash(txn)
	if err != nil {
		return nil, err
	}

	return l.GetPovBlockByHash(latestHash, txn)
}

func (l *Ledger) HasPovBlock(height uint64, hash types.Hash, txns ...db.StoreTxn) bool {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	if !l.HasPovHeader(height, hash, txn) {
		return false
	}

	if !l.HasPovBody(height, hash, txn) {
		return false
	}

	return true
}

func (l *Ledger) DropAllPovBlocks() error {
	txn, flag := l.getTxn(true)
	defer l.releaseTxn(txn, flag)

	prefix, _ := getKeyOfParts(idPrefixPovHeader)
	_ = txn.Drop(prefix)

	prefix, _ = getKeyOfParts(idPrefixPovBody)
	_ = txn.Drop(prefix)

	prefix, _ = getKeyOfParts(idPrefixPovHeight)
	_ = txn.Drop(prefix)

	prefix, _ = getKeyOfParts(idPrefixPovTxLookup)
	_ = txn.Drop(prefix)

	prefix, _ = getKeyOfParts(idPrefixPovBestHash)
	_ = txn.Drop(prefix)

	prefix, _ = getKeyOfParts(idPrefixPovTD)
	_ = txn.Drop(prefix)

	prefix, _ = getKeyOfParts(idPrefixPovMinerStat)
	_ = txn.Drop(prefix)

	return nil
}

func (l *Ledger) CountPovBlocks(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	return txn.Count([]byte{idPrefixPovHeader})
}

func (l *Ledger) CountPovTxs(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	return txn.Count([]byte{idPrefixPovTxLookup})
}

func (l *Ledger) CountPovBestHashs(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	return txn.Count([]byte{idPrefixPovBestHash})
}
