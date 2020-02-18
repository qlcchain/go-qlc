package ledger

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

var (
	ErrPovKeyNotFound      = errors.New("pov key not found")
	ErrPovHeaderNotFound   = errors.New("pov header not found")
	ErrPovBodyNotFound     = errors.New("pov body not found")
	ErrPovHeightNotFound   = errors.New("pov height not found")
	ErrPovHashNotFound     = errors.New("pov hash not found")
	ErrPovTxLookupNotFound = errors.New("pov tx lookup not found")
	ErrPovTxLookupNotBest  = errors.New("pov tx lookup not best")
	ErrPovTDNotFound       = errors.New("pov total difficulty not found")

	ErrPovMinerStatNotFound = errors.New("pov miner statistics not found")
	ErrPovDiffStatNotFound  = errors.New("pov difficulty statistics not found")
)

type PovStore interface {
	AddPovBlock(blk *types.PovBlock, td *types.PovTD, batch ...storage.Batch) error
	DeletePovBlock(blk *types.PovBlock) error

	AddPovHeader(header *types.PovHeader) error
	DeletePovHeader(height uint64, hash types.Hash) error
	GetPovHeader(height uint64, hash types.Hash) (*types.PovHeader, error)
	HasPovHeader(height uint64, hash types.Hash, batch ...storage.Batch) bool

	AddPovBody(height uint64, hash types.Hash, body *types.PovBody) error
	DeletePovBody(height uint64, hash types.Hash) error
	GetPovBody(height uint64, hash types.Hash) (*types.PovBody, error)
	HasPovBody(height uint64, hash types.Hash, batch ...storage.Batch) bool

	AddPovHeight(hash types.Hash, height uint64) error
	DeletePovHeight(hash types.Hash) error
	GetPovHeight(hash types.Hash) (uint64, error)
	HasPovHeight(hash types.Hash) bool

	AddPovTD(hash types.Hash, height uint64, td *types.PovTD) error
	DeletePovTD(hash types.Hash, height uint64) error
	GetPovTD(hash types.Hash, height uint64, batch ...storage.Batch) (*types.PovTD, error)

	AddPovTxLookup(txHash types.Hash, txLookup *types.PovTxLookup, batch ...storage.Batch) error
	DeletePovTxLookup(txHash types.Hash, batch ...storage.Batch) error
	GetPovTxLookup(txHash types.Hash) (*types.PovTxLookup, error)
	HasPovTxLookup(txHash types.Hash) bool

	AddPovTxLookupInBatch(txHash types.Hash, txLookup *types.PovTxLookup, batch storage.Batch) error
	DeletePovTxLookupInBatch(txHash types.Hash, batch storage.Batch) error
	SetPovTxlScanCursor(height uint64, batch ...storage.Batch) error
	GetPovTxlScanCursor() (uint64, error)

	AddPovBestHash(height uint64, hash types.Hash, batch ...storage.Batch) error
	DeletePovBestHash(height uint64, batch ...storage.Batch) error
	GetPovBestHash(height uint64, batch ...storage.Batch) (types.Hash, error)
	GetAllPovBestHashes(fn func(height uint64, hash types.Hash) error) error

	SetPovLatestHeight(height uint64, batch ...storage.Batch) error
	GetPovLatestHeight() (uint64, error)

	AddPovMinerStat(dayStat *types.PovMinerDayStat) error
	DeletePovMinerStat(dayIndex uint32) error
	GetPovMinerStat(dayIndex uint32, batch ...storage.Batch) (*types.PovMinerDayStat, error)
	HasPovMinerStat(dayIndex uint32) bool
	GetLatestPovMinerStat(batch ...storage.Batch) (*types.PovMinerDayStat, error)
	GetAllPovMinerStats(fn func(*types.PovMinerDayStat) error) error

	AddPovDiffStat(dayStat *types.PovDiffDayStat) error
	DeletePovDiffStat(dayIndex uint32) error
	GetPovDiffStat(dayIndex uint32) (*types.PovDiffDayStat, error)
	GetLatestPovDiffStat() (*types.PovDiffDayStat, error)

	GetAllPovDiffStats(fn func(*types.PovDiffDayStat) error) error
	GetPovBlockByHeightAndHash(height uint64, hash types.Hash) (*types.PovBlock, error)
	GetPovBlockByHeight(height uint64) (*types.PovBlock, error)
	GetPovBlockByHash(hash types.Hash) (*types.PovBlock, error)
	GetPovHeaderByHeight(height uint64) (*types.PovHeader, error)

	BatchGetPovHeadersByHeightAsc(height uint64, count uint64) ([]*types.PovHeader, error)
	BatchGetPovHeadersByHeightDesc(height uint64, count uint64) ([]*types.PovHeader, error)
	GetPovHeaderByHash(hash types.Hash) (*types.PovHeader, error)

	GetAllPovHeaders(fn func(header *types.PovHeader) error) error
	GetAllPovBlocks(fn func(*types.PovBlock) error) error
	GetAllPovBestHeaders(fn func(header *types.PovHeader) error) error
	GetAllPovBestBlocks(fn func(*types.PovBlock) error) error

	GetLatestPovBestHash() (types.Hash, error)
	GetLatestPovHeader() (*types.PovHeader, error)
	GetLatestPovBlock() (*types.PovBlock, error)
	HasPovBlock(height uint64, hash types.Hash, batch ...storage.Batch) bool
	DropAllPovBlocks() error

	CountPovBlocks() (uint64, error)
	CountPovTxs() (uint64, error)
	CountPovBestHashs() (uint64, error)
}

func (l *Ledger) AddPovBlock(blk *types.PovBlock, td *types.PovTD, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	if err := l.addPovHeader(blk.GetHeader(), b); err != nil {
		return err
	}

	if err := l.addPovBody(blk.GetHeight(), blk.GetHash(), blk.GetBody(), b); err != nil {
		return err
	}

	if err := l.addPovHeight(blk.GetHash(), blk.GetHeight(), b); err != nil {
		return err
	}

	if err := l.addPovTD(blk.GetHash(), blk.GetHeight(), td, b); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) DeletePovBlock(blk *types.PovBlock) error {
	batch := l.store.Batch(true)

	if err := l.deletePovHeader(blk.GetHeight(), blk.GetHash(), batch); err != nil {
		return err
	}

	if err := l.deletePovBody(blk.GetHeight(), blk.GetHash(), batch); err != nil {
		return err
	}

	if err := l.deletePovHeight(blk.GetHash(), batch); err != nil {
		return err
	}

	if err := l.deletePovTD(blk.GetHash(), blk.GetHeight(), batch); err != nil {
		return err
	}
	return l.store.PutBatch(batch)
}

func (l *Ledger) addPovHeader(header *types.PovHeader, batch storage.Batch) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovHeader, header.GetHeight(), header.GetHash())
	if err != nil {
		return err
	}

	_, err = batch.Get(key)
	if err == nil {
		return ErrBlockExists
	} else if err != storage.KeyNotFound {
		return err
	}

	dataBytes, err := header.Serialize()
	if err != nil {
		return err
	}

	return batch.Put(key, dataBytes)
}

func (l *Ledger) AddPovHeader(header *types.PovHeader) error {
	batch := l.store.Batch(true)
	if err := l.addPovHeader(header, batch); err != nil {
		return err
	}
	return l.store.PutBatch(batch)
}

func (l *Ledger) deletePovHeader(height uint64, hash types.Hash, txn storage.Batch) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovHeader, height, hash)
	if err != nil {
		return err
	}

	return txn.Delete(key)
}

func (l *Ledger) DeletePovHeader(height uint64, hash types.Hash) error {
	batch := l.store.Batch(true)
	if err := l.deletePovHeader(height, hash, batch); err != nil {
		l.logger.Error(err)
		return err
	}
	return l.store.PutBatch(batch)
}

func (l *Ledger) GetPovHeader(height uint64, hash types.Hash) (*types.PovHeader, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovHeader, height, hash)
	if err != nil {
		return nil, err
	}

	hdr := new(types.PovHeader)
	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrPovHeaderNotFound
		}
		return nil, err
	}

	if err := hdr.Deserialize(val); err != nil {
		return nil, err
	}
	return hdr, nil
}

func (l *Ledger) HasPovHeader(height uint64, hash types.Hash, batch ...storage.Batch) bool {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovHeader, height, hash)
	if err != nil {
		return false
	}

	_, err = l.getFromStore(key, batch...)
	if err != nil {
		return false
	}
	return true
}

func (l *Ledger) addPovBody(height uint64, hash types.Hash, body *types.PovBody, txn storage.Batch) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovBody, height, hash)
	if err != nil {
		return err
	}

	_, err = txn.Get(key)
	if err == nil {
		return ErrBlockExists
	} else if err != storage.KeyNotFound {
		return err
	}

	dataBytes, err := body.Serialize()
	if err != nil {
		return err
	}

	if err := txn.Put(key, dataBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) AddPovBody(height uint64, hash types.Hash, body *types.PovBody) error {
	batch := l.store.Batch(true)
	if err := l.addPovBody(height, hash, body, batch); err != nil {
		l.logger.Error(err)
		return err
	}
	return l.store.PutBatch(batch)
}

func (l *Ledger) deletePovBody(height uint64, hash types.Hash, batch storage.Batch) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovBody, height, hash)
	if err != nil {
		return err
	}

	return batch.Delete(key)
}

func (l *Ledger) DeletePovBody(height uint64, hash types.Hash) error {
	batch := l.store.Batch(true)
	if err := l.deletePovBody(height, hash, batch); err != nil {
		return err
	}
	return l.store.PutBatch(batch)
}

func (l *Ledger) GetPovBody(height uint64, hash types.Hash) (*types.PovBody, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovBody, height, hash)
	if err != nil {
		return nil, err
	}

	body := new(types.PovBody)
	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrPovBodyNotFound
		}
		return nil, err
	}
	if err := body.Deserialize(val); err != nil {
		return nil, err
	}

	return body, nil
}

func (l *Ledger) HasPovBody(height uint64, hash types.Hash, batch ...storage.Batch) bool {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovBody, height, hash)
	if err != nil {
		return false
	}

	_, err = l.getFromStore(key, batch...)
	if err != nil {
		return false
	}
	return true
}

func (l *Ledger) addPovHeight(hash types.Hash, height uint64, txn storage.Batch) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovHeight, hash)
	if err != nil {
		return err
	}

	_, err = txn.Get(key)
	if err == nil {
		return ErrBlockExists
	} else if err != storage.KeyNotFound {
		return err
	}

	blockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBytes, height)

	if err := txn.Put(key, blockBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) AddPovHeight(hash types.Hash, height uint64) error {
	batch := l.store.Batch(true)
	if err := l.addPovHeight(hash, height, batch); err != nil {
		return err
	}
	return l.store.PutBatch(batch)
}

func (l *Ledger) deletePovHeight(hash types.Hash, batch storage.Batch) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovHeight, hash)
	if err != nil {
		return err
	}

	return batch.Delete(key)
}

func (l *Ledger) DeletePovHeight(hash types.Hash) error {
	batch := l.store.Batch(true)
	if err := l.deletePovHeight(hash, batch); err != nil {
		return err
	}
	return l.store.PutBatch(batch)
}

func (l *Ledger) GetPovHeight(hash types.Hash) (uint64, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovHeight, hash)
	if err != nil {
		return 0, err
	}

	var height uint64
	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return 0, ErrPovHeightNotFound
		}
		return 0, err
	}

	height = binary.BigEndian.Uint64(val)
	return height, nil
}

func (l *Ledger) HasPovHeight(hash types.Hash) bool {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovHeight, hash)
	if err != nil {
		return false
	}

	if b, _ := l.store.Has(key); b {
		return true
	} else {
		return false
	}
}

func (l *Ledger) addPovTD(hash types.Hash, height uint64, td *types.PovTD, txn storage.Batch) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTD, height, hash)
	if err != nil {
		return err
	}

	_, err = txn.Get(key)
	if err == nil {
		return ErrBlockExists
	} else if err != storage.KeyNotFound {
		return err
	}

	val, err := td.Serialize()
	if err != nil {
		return err
	}

	if err := txn.Put(key, val); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) AddPovTD(hash types.Hash, height uint64, td *types.PovTD) error {
	batch := l.store.Batch(true)
	if err := l.addPovTD(hash, height, td, batch); err != nil {
		return err
	}
	return l.store.PutBatch(batch)
}

func (l *Ledger) deletePovTD(hash types.Hash, height uint64, batch storage.Batch) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTD, height, hash)
	if err != nil {
		return err
	}

	return batch.Delete(key)
}

func (l *Ledger) DeletePovTD(hash types.Hash, height uint64) error {
	batch := l.store.Batch(true)
	if err := l.deletePovTD(hash, height, batch); err != nil {
		return err
	}
	return l.store.PutBatch(batch)
}

func (l *Ledger) GetPovTD(hash types.Hash, height uint64, batch ...storage.Batch) (*types.PovTD, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTD, height, hash)
	if err != nil {
		return nil, err
	}

	td := new(types.PovTD)
	val, err := l.getFromStore(key, batch...)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrPovTDNotFound
		}
		return nil, err
	}
	if err := td.Deserialize(val); err != nil {
		return nil, err
	}
	return td, nil
}

func (l *Ledger) AddPovTxLookup(txHash types.Hash, txLookup *types.PovTxLookup, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTxLookup, txHash)
	if err != nil {
		return err
	}

	dataTypes, err := txLookup.Serialize()
	if err != nil {
		return err
	}

	if err := b.Put(key, dataTypes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovTxLookup(txHash types.Hash, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTxLookup, txHash)
	if err != nil {
		return err
	}

	if err := b.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovTxLookup(txHash types.Hash) (*types.PovTxLookup, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTxLookup, txHash)
	if err != nil {
		return nil, err
	}

	txLookup := new(types.PovTxLookup)
	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrPovTxLookupNotFound
		}
		return nil, err
	}

	if err := txLookup.Deserialize(val); err != nil {
		return nil, err
	}
	bestHash, err := l.GetPovBestHash(txLookup.BlockHeight)
	if err != nil {
		return nil, ErrPovTxLookupNotBest
	}
	if bestHash != txLookup.BlockHash {
		return nil, ErrPovTxLookupNotBest
	}

	return txLookup, nil
}

func (l *Ledger) HasPovTxLookup(txHash types.Hash) bool {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTxLookup, txHash)
	if err != nil {
		return false
	}

	txLookup := new(types.PovTxLookup)
	val, err := l.store.Get(key)
	if err != nil {
		return false
	}

	if err := txLookup.Deserialize(val); err != nil {
		return false
	}

	bestHash, err := l.GetPovBestHash(txLookup.BlockHeight)
	if err != nil {
		return false
	}
	if bestHash != txLookup.BlockHash {
		return false
	}

	return true
}

func (l *Ledger) AddPovTxLookupInBatch(txHash types.Hash, txLookup *types.PovTxLookup, batch storage.Batch) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTxLookup, txHash)
	if err != nil {
		return err
	}

	dataTypes, err := txLookup.Serialize()
	if err != nil {
		return err
	}

	if err := batch.Put(key, dataTypes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovTxLookupInBatch(txHash types.Hash, batch storage.Batch) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTxLookup, txHash)
	if err != nil {
		return err
	}

	if err := batch.Delete(key); err != nil {
		return err
	}

	return nil
}

//getTxn get txn by `update` mode
func (l *Ledger) getBatch(update bool, batch ...storage.Batch) (storage.Batch, bool) {
	if len(batch) > 0 {
		//logger.Debugf("getTxn %p", txns[0])
		return batch[0], false
	} else {
		b := l.store.Batch(true)
		return b, true
	}
}

// releaseTxn commit change and close txn
func (l *Ledger) releaseBatch(batch storage.Batch, flag bool) {
	if flag {
		err := l.store.PutBatch(batch)
		if err != nil {
			l.logger.Error(err)
		}
	}
}

func (l *Ledger) SetPovTxlScanCursor(height uint64, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTxlScanCursor)
	if err != nil {
		return err
	}

	valBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(valBytes, height)

	if err := b.Put(key, valBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovTxlScanCursor() (uint64, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovTxlScanCursor)
	if err != nil {
		return 0, err
	}

	var height uint64
	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return 0, ErrPovKeyNotFound
		}
		return 0, err
	}
	height = binary.BigEndian.Uint64(val)
	return height, nil
}

func (l *Ledger) AddPovBestHash(height uint64, hash types.Hash, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovBestHash, height)
	if err != nil {
		return err
	}

	valBytes := make([]byte, hash.Len())
	err = hash.MarshalBinaryTo(valBytes)
	if err != nil {
		return err
	}

	if err := b.Put(key, valBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovBestHash(height uint64, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovBestHash, height)
	if err != nil {
		return err
	}

	if err := b.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovBestHash(height uint64, batch ...storage.Batch) (types.Hash, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovBestHash, height)
	if err != nil {
		return types.ZeroHash, err
	}

	return l.getPovBestHash(key, batch...)
}

func (l *Ledger) getPovBestHash(key []byte, batch ...storage.Batch) (types.Hash, error) {
	var hash types.Hash
	val, err := l.getFromStore(key, batch...)
	if err != nil {
		if err == storage.KeyNotFound {
			return types.ZeroHash, ErrPovHashNotFound
		}
		return types.ZeroHash, err
	}
	err = hash.UnmarshalBinary(val)
	if err != nil {
		return types.ZeroHash, err
	}
	return hash, nil
}

func (l *Ledger) GetAllPovBestHashes(fn func(height uint64, hash types.Hash) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPovBestHash)

	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
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

func (l *Ledger) SetPovLatestHeight(height uint64, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovLatestHeight)
	if err != nil {
		return err
	}

	valBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(valBytes, height)

	if err := b.Put(key, valBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovLatestHeight() (uint64, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovLatestHeight)
	if err != nil {
		return 0, err
	}

	var height uint64
	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return 0, ErrPovHeightNotFound
		}
		return 0, err
	}
	height = binary.BigEndian.Uint64(val)
	return height, nil
}

func (l *Ledger) AddPovMinerStat(dayStat *types.PovMinerDayStat) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovMinerStat, dayStat.DayIndex)
	if err != nil {
		return err
	}

	valBytes, err := dayStat.Serialize()
	if err != nil {
		return err
	}

	if err := l.store.Put(key, valBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovMinerStat(dayIndex uint32) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovMinerStat, dayIndex)
	if err != nil {
		return err
	}

	if err := l.store.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) getPovMinerStat(key []byte, batch ...storage.Batch) (*types.PovMinerDayStat, error) {
	dayStat := new(types.PovMinerDayStat)
	val, err := l.getFromStore(key, batch...)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrPovMinerStatNotFound
		}
		return nil, err
	}
	err = dayStat.Deserialize(val)
	if err != nil {
		return nil, err
	}
	return dayStat, nil
}

func (l *Ledger) GetPovMinerStat(dayIndex uint32, batch ...storage.Batch) (*types.PovMinerDayStat, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovMinerStat, dayIndex)
	if err != nil {
		return nil, err
	}
	return l.getPovMinerStat(key, batch...)
}

func (l *Ledger) HasPovMinerStat(dayIndex uint32) bool {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovMinerStat, dayIndex)
	if err != nil {
		return false
	}

	if b, _ := l.store.Has(key); b {
		return true
	} else {
		return false
	}
}

func (l *Ledger) GetLatestPovMinerStat(batch ...storage.Batch) (*types.PovMinerDayStat, error) {
	var latestKey []byte
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPovMinerStat)
	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		// just safe copy, iterator will reuse item(key, value)
		if latestKey == nil || len(latestKey) != len(key) {
			latestKey = make([]byte, len(key))
		}
		copy(latestKey, key)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(latestKey) <= 0 {
		return nil, err
	}

	return l.getPovMinerStat(latestKey, batch...)
}

func (l *Ledger) GetAllPovMinerStats(fn func(*types.PovMinerDayStat) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPovMinerStat)
	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		dayStat := new(types.PovMinerDayStat)
		err := dayStat.Deserialize(val)
		if err != nil {
			return err
		}

		if err := fn(dayStat); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (l *Ledger) AddPovDiffStat(dayStat *types.PovDiffDayStat) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovDiffStat, dayStat.DayIndex)
	if err != nil {
		return err
	}

	valBytes, err := dayStat.Serialize()
	if err != nil {
		return err
	}

	if err := l.store.Put(key, valBytes); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeletePovDiffStat(dayIndex uint32) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovDiffStat, dayIndex)
	if err != nil {
		return err
	}

	if err := l.store.Delete(key); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) getPovDiffStat(key []byte) (*types.PovDiffDayStat, error) {
	dayStat := new(types.PovDiffDayStat)
	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrPovDiffStatNotFound
		}
		return nil, err
	}
	err = dayStat.Deserialize(val)
	if err != nil {
		return nil, err
	}
	return dayStat, nil
}

func (l *Ledger) GetPovDiffStat(dayIndex uint32) (*types.PovDiffDayStat, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPovDiffStat, dayIndex)
	if err != nil {
		return nil, err
	}

	return l.getPovDiffStat(key)
}

func (l *Ledger) GetLatestPovDiffStat() (*types.PovDiffDayStat, error) {
	var latestKey []byte
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPovDiffStat)
	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		// just safe copy, iterator will reuse item(key, value)
		if latestKey == nil || len(latestKey) != len(key) {
			latestKey = make([]byte, len(key))
		}
		copy(latestKey, key)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(latestKey) <= 0 {
		return nil, ErrPovDiffStatNotFound
	}

	return l.getPovDiffStat(latestKey)
}

func (l *Ledger) GetAllPovDiffStats(fn func(*types.PovDiffDayStat) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPovDiffStat)
	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		dayStat := new(types.PovDiffDayStat)
		err := dayStat.Deserialize(val)
		if err != nil {
			return err
		}

		if err := fn(dayStat); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (l *Ledger) GetPovBlockByHeightAndHash(height uint64, hash types.Hash) (*types.PovBlock, error) {
	header, err := l.GetPovHeader(height, hash)
	if err != nil {
		return nil, err
	}

	body, err := l.GetPovBody(height, hash)
	if err != nil {
		return nil, err
	}

	blk := types.NewPovBlockWithBody(header, body)
	return blk, nil
}

func (l *Ledger) GetPovBlockByHeight(height uint64) (*types.PovBlock, error) {
	hash, err := l.GetPovBestHash(height)
	if err != nil {
		return nil, err
	}

	return l.GetPovBlockByHeightAndHash(height, hash)
}

func (l *Ledger) GetPovBlockByHash(hash types.Hash) (*types.PovBlock, error) {
	height, err := l.GetPovHeight(hash)
	if err != nil {
		return nil, err
	}

	return l.GetPovBlockByHeightAndHash(height, hash)
}

func (l *Ledger) GetPovHeaderByHeight(height uint64) (*types.PovHeader, error) {
	hash, err := l.GetPovBestHash(height)
	if err != nil {
		return nil, err
	}

	return l.GetPovHeader(height, hash)
}

func (l *Ledger) BatchGetPovHeadersByHeightAsc(height uint64, count uint64) ([]*types.PovHeader, error) {
	startKey, err := storage.GetKeyOfParts(storage.KeyPrefixPovBestHash, height)
	if err != nil {
		return nil, err
	}
	endKey, err := storage.GetKeyOfParts(storage.KeyPrefixPovBestHash, uint64(height+count))
	if err != nil {
		return nil, err
	}
	return l.batchGetPovHeadersByHeight(startKey, endKey)
}

func (l *Ledger) BatchGetPovHeadersByHeightDesc(height uint64, count uint64) ([]*types.PovHeader, error) {
	if height < count {
		return nil, errors.New("height should greater than count")
	}

	startKey, err := storage.GetKeyOfParts(storage.KeyPrefixPovBestHash, uint64(height-count+1))
	if err != nil {
		return nil, err
	}
	endKey, err := storage.GetKeyOfParts(storage.KeyPrefixPovBestHash, height+1)
	if err != nil {
		return nil, err
	}

	headers, err := l.batchGetPovHeadersByHeight(startKey, endKey)
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

func (l *Ledger) batchGetPovHeadersByHeight(startKey []byte, endKey []byte) ([]*types.PovHeader, error) {
	var headers []*types.PovHeader
	err := l.store.Iterator(startKey, endKey, func(key []byte, val []byte) error {
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

func (l *Ledger) GetPovHeaderByHash(hash types.Hash) (*types.PovHeader, error) {
	height, err := l.GetPovHeight(hash)
	if err != nil {
		return nil, err
	}

	return l.GetPovHeader(height, hash)
}

func (l *Ledger) GetAllPovHeaders(fn func(header *types.PovHeader) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPovHeader)
	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
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

func (l *Ledger) GetAllPovBlocks(fn func(*types.PovBlock) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPovHeader)
	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		header := new(types.PovHeader)
		if err := header.Deserialize(val); err != nil {
			return err
		}

		body, err := l.GetPovBody(header.GetHeight(), header.GetHash())
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

func (l *Ledger) GetAllPovBestHeaders(fn func(header *types.PovHeader) error) error {
	err := l.GetAllPovBestHashes(func(height uint64, hash types.Hash) error {
		header, err := l.GetPovHeader(height, hash)
		if err != nil {
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

func (l *Ledger) GetAllPovBestBlocks(fn func(*types.PovBlock) error) error {
	err := l.GetAllPovBestHashes(func(height uint64, hash types.Hash) error {
		block, err := l.GetPovBlockByHeightAndHash(height, hash)
		if err != nil {
			return err
		}

		if err := fn(block); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetLatestPovBestHash() (types.Hash, error) {
	latestHeight, err := l.GetPovLatestHeight()
	if err != nil {
		return types.ZeroHash, err
	}

	latestHash, err := l.GetPovBestHash(latestHeight)
	if err != nil {
		return types.ZeroHash, err
	}
	if latestHash.IsZero() {
		return types.ZeroHash, fmt.Errorf("latest best hash value is zero")
	}

	return latestHash, nil
}

func (l *Ledger) GetLatestPovHeader() (*types.PovHeader, error) {
	latestHash, err := l.GetLatestPovBestHash()
	if err != nil {
		return nil, err
	}

	height, err := l.GetPovHeight(latestHash)
	if err != nil {
		return nil, err
	}

	return l.GetPovHeader(height, latestHash)
}

func (l *Ledger) GetLatestPovBlock() (*types.PovBlock, error) {
	latestHash, err := l.GetLatestPovBestHash()
	if err != nil {
		return nil, err
	}

	return l.GetPovBlockByHash(latestHash)
}

func (l *Ledger) HasPovBlock(height uint64, hash types.Hash, batch ...storage.Batch) bool {
	if !l.HasPovHeader(height, hash, batch...) {
		fmt.Println("==========")
		return false
	}

	if !l.HasPovBody(height, hash, batch...) {
		return false
	}

	return true
}

func (l *Ledger) DropAllPovBlocks() error {
	batch := l.store.Batch(true)
	defer func() {
		if err := l.store.PutBatch(batch); err != nil {
			l.logger.Error(err)
		}
	}()
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPovHeader)
	_ = batch.Drop(prefix)

	prefix, _ = storage.GetKeyOfParts(storage.KeyPrefixPovBody)
	_ = batch.Drop(prefix)

	prefix, _ = storage.GetKeyOfParts(storage.KeyPrefixPovHeight)
	_ = batch.Drop(prefix)

	prefix, _ = storage.GetKeyOfParts(storage.KeyPrefixPovTxLookup)
	_ = batch.Drop(prefix)

	prefix, _ = storage.GetKeyOfParts(storage.KeyPrefixPovBestHash)
	_ = batch.Drop(prefix)

	prefix, _ = storage.GetKeyOfParts(storage.KeyPrefixPovTD)
	_ = batch.Drop(prefix)

	prefix, _ = storage.GetKeyOfParts(storage.KeyPrefixPovMinerStat)
	_ = batch.Drop(prefix)

	prefix, _ = storage.GetKeyOfParts(storage.KeyPrefixPovLatestHeight)
	_ = batch.Drop(prefix)

	prefix, _ = storage.GetKeyOfParts(storage.KeyPrefixPovTxlScanCursor)
	_ = batch.Drop(prefix)

	prefix, _ = storage.GetKeyOfParts(storage.KeyPrefixPovDiffStat)
	_ = batch.Drop(prefix)

	return nil
}

func (l *Ledger) CountPovBlocks() (uint64, error) {
	return l.store.Count([]byte{byte(storage.KeyPrefixPovHeader)})
}

func (l *Ledger) CountPovTxs() (uint64, error) {
	return l.store.Count([]byte{byte(storage.KeyPrefixPovTxLookup)})
}

func (l *Ledger) CountPovBestHashs() (uint64, error) {
	return l.store.Count([]byte{byte(storage.KeyPrefixPovBestHash)})
}
