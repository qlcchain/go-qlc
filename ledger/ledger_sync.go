package ledger

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

func (l *Ledger) getUncheckedSyncBlockKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = byte(storage.KeyPrefixUncheckedSync)
	copy(key[1:], hash[:])
	return key[:]
}

func (l *Ledger) AddUncheckedSyncBlock(previous types.Hash, blk *types.StateBlock) error {
	blockBytes, err := blk.Serialize()
	if err != nil {
		return err
	}

	key := l.getUncheckedSyncBlockKey(previous)
	return l.store.Put(key, blockBytes)
}

func (l *Ledger) GetUncheckedSyncBlock(hash types.Hash) (*types.StateBlock, error) {
	key := l.getUncheckedSyncBlockKey(hash)

	blk := new(types.StateBlock)

	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	if err = blk.Deserialize(val); err != nil {
		return nil, err
	}
	return blk, nil
}

func (l *Ledger) HasUncheckedSyncBlock(hash types.Hash) (bool, error) {
	key := l.getUncheckedSyncBlockKey(hash)
	return l.store.Has(key)
}

func (l *Ledger) CountUncheckedSyncBlocks() (uint64, error) {
	var count uint64
	count, err := l.store.Count([]byte{byte(storage.KeyPrefixUncheckedSync)})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (l *Ledger) DeleteUncheckedSyncBlock(hash types.Hash) error {
	key := l.getUncheckedSyncBlockKey(hash)

	return l.store.Delete(key)
}

func (l *Ledger) getUnconfirmedSyncBlockKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = byte(storage.KeyPrefixUnconfirmedSync)
	copy(key[1:], hash[:])
	return key[:]
}

func (l *Ledger) AddUnconfirmedSyncBlock(hash types.Hash, blk *types.StateBlock) error {
	blockBytes, err := blk.Serialize()
	if err != nil {
		return err
	}

	key := l.getUnconfirmedSyncBlockKey(hash)
	return l.store.Put(key, blockBytes)
}

func (l *Ledger) GetUnconfirmedSyncBlock(hash types.Hash) (*types.StateBlock, error) {
	key := l.getUnconfirmedSyncBlockKey(hash)
	blk := new(types.StateBlock)
	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	if err = blk.Deserialize(val); err != nil {
		return nil, err
	}

	return blk, nil
}

func (l *Ledger) HasUnconfirmedSyncBlock(hash types.Hash) (bool, error) {
	key := l.getUnconfirmedSyncBlockKey(hash)
	return l.store.Has(key)
}

func (l *Ledger) CountUnconfirmedSyncBlocks() (uint64, error) {
	var count uint64
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixUnconfirmedSync)
	count, err := l.store.Count(prefix)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (l *Ledger) DeleteUnconfirmedSyncBlock(hash types.Hash) error {
	key := l.getUnconfirmedSyncBlockKey(hash)
	return l.store.Delete(key)
}

func (l *Ledger) WalkSyncCache(visit common.SyncCacheWalkFunc) {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixUnconfirmedSync)
	_ = l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		visit(common.SyncCacheUnconfirmed, key)
		return nil
	})

	prefix2, _ := storage.GetKeyOfParts(storage.KeyPrefixUncheckedSync)
	_ = l.store.Iterator(prefix2, nil, func(key []byte, val []byte) error {
		visit(common.SyncCacheUnchecked, key)
		return nil
	})
}

func (l *Ledger) CleanSyncCache() {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixUnconfirmedSync)
	_ = l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		hash, _ := types.BytesToHash(key[1:])
		_ = l.DeleteUnconfirmedSyncBlock(hash)
		return nil
	})

	prefix2, _ := storage.GetKeyOfParts(storage.KeyPrefixUncheckedSync)
	_ = l.store.Iterator(prefix2, nil, func(key []byte, val []byte) error {
		hash, _ := types.BytesToHash(key[1:])
		_ = l.DeleteUncheckedSyncBlock(hash)
		return nil
	})
}
