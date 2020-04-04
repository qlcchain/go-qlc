package db

import (
	"bytes"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
	"github.com/pkg/errors"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

type BadgerStore struct {
	db *badger.DB
}

// NewBadgerStore initializes/opens a badger database in the given directory.
func NewBadgerStore(dir string) (storage.Store, error) {
	opts := badger.DefaultOptions(dir)

	opts.MaxTableSize = common.BadgerMaxTableSize
	opts.Logger = nil
	opts.ValueLogLoadingMode = options.FileIO
	_ = util.CreateDirIfNotExist(dir)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerStore{db: db}, nil
}

func (b *BadgerStore) Get(k []byte) ([]byte, error) {
	txn := b.db.NewTransaction(false)
	defer txn.Discard()
	item, err := txn.Get(k)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, storage.KeyNotFound
		}
		return nil, err
	}
	var v []byte
	err = item.Value(func(val []byte) error {
		v = val
		return nil
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (b *BadgerStore) Put(k, v []byte) error {
	txn := b.db.NewTransaction(true)
	if err := txn.Set(k, v); err != nil {
		return err
	}
	return txn.Commit()
}

func (b *BadgerStore) Delete(k []byte) error {
	txn := b.db.NewTransaction(true)
	if err := txn.Delete(k); err != nil {
		return err
	}
	return txn.Commit()
}

func (b *BadgerStore) Has(k []byte) (bool, error) {
	txn := b.db.NewTransaction(false)
	defer txn.Discard()
	_, err := txn.Get(k)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (b *BadgerStore) Batch(canRead bool) storage.Batch {
	if canRead {
		return &BadgerTransaction{
			txn: b.db.NewTransaction(true),
		}
	} else {
		return &BadgerWriteBatch{
			batch: b.db.NewWriteBatch(),
		}
	}
}

func (b *BadgerStore) PutBatch(batch storage.Batch) error {
	if bb, ok := batch.(*BadgerTransaction); ok {
		defer bb.txn.Discard()
		return bb.txn.Commit()
	} else if bb, ok := batch.(*BadgerWriteBatch); ok {
		return bb.batch.Flush()
	}
	return errors.New("error batch type")
}

func (b *BadgerStore) BatchWrite(canRead bool, fn func(batch storage.Batch) error) error {
	if !canRead {
		b := &BadgerWriteBatch{
			batch: b.db.NewWriteBatch(),
		}
		if err := fn(b); err != nil {
			b.batch.Cancel()
			return err
		}
		return b.batch.Flush()

	} else {
		tx := &BadgerTransaction{
			txn: b.db.NewTransaction(true),
		}
		if err := fn(tx); err != nil {
			tx.txn.Discard()
			return err
		}
		return tx.txn.Commit()
	}
}

//func (b *BadgerStore) BatchView(fn func(batch storage.Batch) error) error {
//	tx := &BadgerTransaction{
//		txn: b.db.NewTransaction(false),
//	} //logger.Debugf("BatchView NewTransaction %p", txn)
//	defer func() {
//		//logger.Debugf("BatchView Discard %p", txn)
//		tx.txn.Discard()
//	}()
//
//	if err := fn(tx); err != nil {
//		return err
//	}
//	return nil
//}

func (b *BadgerStore) Iterator(prefix []byte, end []byte, fn func(k, v []byte) error) error {
	if len(prefix) <= 0 {
		return errors.New("invalid prefix")
	}

	txn := b.db.NewTransaction(false)
	defer txn.Discard()
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	if end == nil {
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				return fn(key, val)
			})
			if err != nil {
				return err
			}
		}
	} else {
		for it.Seek(prefix); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			if bytes.Compare(key, end) >= 0 {
				break
			}
			err := item.Value(func(val []byte) error {
				return fn(key, val)
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *BadgerStore) Count(prefix []byte) (uint64, error) {
	var i uint64
	err := t.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			i++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (b *BadgerStore) Purge() error {
	return b.db.RunValueLogGC(0.5)
}

func (b *BadgerStore) Drop(prefix []byte) error {
	if prefix == nil {
		return b.db.DropAll()
	} else {
		txn := b.db.NewTransaction(true)
		defer txn.Commit()

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			k := make([]byte, len(key))
			copy(k, key)
			if err := txn.Delete(k); err != nil {
				return err
			}
		}
		return nil
	}
}

func (b *BadgerStore) Close() error {
	return b.db.Close()
}

func (b *BadgerStore) Action(at storage.ActionType) (interface{}, error) {
	switch at {
	case storage.GC:
		err := b.db.RunValueLogGC(0.5)
		if err != nil {
			return nil, err
		}
		return nil, nil
	case storage.Size:
		lsm, vlog := b.db.Size()
		s := make(map[string]int64)
		s["lsm"] = lsm
		s["vlog"] = vlog
		return s, nil
	default:
		return "", errors.New("invalid action type")
	}
}

type BadgerTransaction struct {
	txn *badger.Txn
}

func (b *BadgerTransaction) Get(k []byte) (interface{}, error) {
	item, err := b.txn.Get(k)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, storage.KeyNotFound
		}
		return nil, err
	}
	v := make([]byte, 0)
	err = item.Value(func(val []byte) error {
		v = val
		return err
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (b *BadgerTransaction) Iterator(prefix []byte, end []byte, fn func(k, v []byte) error) error {
	if len(prefix) <= 0 {
		return errors.New("invalid prefix")
	}

	it := b.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	if end == nil {
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				return fn(key, val)
			})
			if err != nil {
				return err
			}
		}
	} else {
		for it.Seek(prefix); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			if bytes.Compare(key, end) >= 0 {
				break
			}
			err := item.Value(func(val []byte) error {
				return fn(key, val)
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *BadgerTransaction) Delete(k []byte) error {
	return b.txn.Delete(k)
}

func (b *BadgerTransaction) Put(k []byte, v interface{}) error {
	return b.txn.Set(k, v.([]byte))
}

func (b *BadgerTransaction) Drop(prefix []byte) error {
	if len(prefix) <= 0 {
		return errors.New("invalid prefix")
	}

	it := b.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		key := item.Key()
		k := make([]byte, len(key))
		copy(k, key)
		if err := b.Delete(k); err != nil {
			return err
		}
	}
	return nil
}

func (b *BadgerTransaction) Discard() {
	b.txn.Discard()
}

type BadgerWriteBatch struct {
	batch *badger.WriteBatch
}

func (b *BadgerWriteBatch) Get([]byte) (interface{}, error) {
	return nil, errors.New("BatchWrite can write only")
}

func (b *BadgerWriteBatch) Iterator(prefix []byte, end []byte, f func(k, v []byte) error) error {
	return errors.New("BatchWrite can write only")
}

func (b *BadgerWriteBatch) Delete(k []byte) error {
	return b.batch.Delete(k)
}

func (b *BadgerWriteBatch) Put(k []byte, v interface{}) error {
	return b.batch.Set(k, v.([]byte))
}

func (b *BadgerWriteBatch) Drop(prefix []byte) error {
	panic("not implemented")
}

func (b *BadgerWriteBatch) Discard() {
	b.batch.Cancel()
}

func (b *BadgerStore) Upgrade(version int) error {
	keys := make([]storage.KeyPrefix, 0)
	keys = append(keys, storage.KeyPrefixUncheckedBlockLink)
	keys = append(keys, storage.KeyPrefixUncheckedBlockPrevious)
	keys = append(keys, storage.KeyPrefixUncheckedTokenInfo)
	keys = append(keys, storage.KeyPrefixGapPublish)
	keys = append(keys, storage.KeyPrefixUncheckedPovHeight)

	unchecks := make([]*uncheckedMap, 0)
	for _, k := range keys {
		txn := b.db.NewTransaction(false)
		it := txn.NewIterator(badger.DefaultIteratorOptions)

		prefix, _ := storage.GetKeyOfParts(k)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				blk := new(types.StateBlock)
				if err := blk.Deserialize(val); err != nil {
					return err
				}
				sync := types.SynchronizedKind(item.UserMeta())
				uncheckVal := &types.Unchecked{
					Block: blk,
					Kind:  sync,
				}
				k := make([]byte, len(key))
				copy(k[:], key)
				uncheckedMap := &uncheckedMap{
					key:   k,
					value: uncheckVal,
				}
				unchecks = append(unchecks, uncheckedMap)
				return nil
			})
			if err != nil {
				return err
			}
		}
		it.Close()
		txn.Discard()
	}
	batch := b.db.NewTransaction(true)
	for _, u := range unchecks {
		k := u.key
		v, err := u.value.Serialize()
		if err != nil {
			return fmt.Errorf("uncheck Serialize error: %s", err)
		}

		if err := batch.Set(k, v); err != nil {
			return err
		}
	}
	if err := batch.Commit(); err != nil {
		return fmt.Errorf("batch flush: %s", err)
	}
	return nil
}

type uncheckedMap struct {
	key   []byte
	value *types.Unchecked
}
