package db

import (
	"context"
	"log"
	"sort"

	"github.com/dgraph-io/badger"
	badgerOpts "github.com/dgraph-io/badger/options"
	"github.com/dgraph-io/badger/pb"
	"github.com/qlcchain/go-qlc/common/util"
)

// BadgerStore represents a block lattice store backed by a badger database.
type BadgerStore struct {
	db *badger.DB
}

type BadgerStoreTxn struct {
	db  *badger.DB
	txn *badger.Txn
}

//var logger = log2.NewLogger("badger")

// NewBadgerStore initializes/opens a badger database in the given directory.
func NewBadgerStore(dir string) (Store, error) {
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	opts.Logger = nil
	opts.ValueLogLoadingMode = badgerOpts.FileIO
	_ = util.CreateDirIfNotExist(dir)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerStore{db: db}, nil
}

func (s *BadgerStore) Erase() error {
	return s.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := txn.Delete(k)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})
}

func (s *BadgerStore) NewTransaction(update bool) *BadgerStoreTxn {
	txn := &BadgerStoreTxn{txn: s.db.NewTransaction(update), db: s.db}
	return txn
}

// Close closes the database
func (s *BadgerStore) Close() error {
	return s.db.Close()
}

// Purge purges any old/deleted keys from the database.
func (s *BadgerStore) Purge() error {
	return s.db.RunValueLogGC(0.5)
}

func (s *BadgerStore) ViewInTx(fn func(txn StoreTxn) error) error {
	return s.db.View(func(txn *badger.Txn) error {
		return fn(&BadgerStoreTxn{txn: txn, db: s.db})
	})
}

func (s *BadgerStore) UpdateInTx(fn func(txn StoreTxn) error) error {
	return s.db.Update(func(txn *badger.Txn) error {
		t := &BadgerStoreTxn{txn: txn, db: s.db}
		return fn(t)
	})
}

func (t *BadgerStoreTxn) Set(key []byte, val []byte) error {
	if err := t.txn.Set(key, val); err != nil {
		return err
	}
	return nil
}

func (t *BadgerStoreTxn) SetWithMeta(key, val []byte, meta byte) error {
	if err := t.txn.SetWithMeta(key[:], val, meta); err != nil {
		return err
	}
	return nil
}

func (t *BadgerStoreTxn) Get(key []byte, fn func([]byte, byte) error) error {
	item, err := t.txn.Get(key)
	if err != nil {
		return err
	}
	err = item.Value(func(val []byte) error {
		err = fn(val, item.UserMeta())
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (t *BadgerStoreTxn) Delete(key []byte) error {
	return t.txn.Delete(key)
}

func (t *BadgerStoreTxn) Iterator(pre byte, fn func([]byte, []byte, byte) error) error {
	it := t.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := [...]byte{pre}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		item := it.Item()
		key := item.Key()
		err := item.Value(func(val []byte) error {
			return fn(key, val, item.UserMeta())
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func (t *BadgerStoreTxn) Drop(prefix []byte) error {
	if prefix == nil {
		return t.db.DropAll()
	} else {
		return t.db.DropPrefix(prefix)
	}
}

func (t *BadgerStoreTxn) Upgrade(migrations []Migration) error {
	sort.Sort(Migrations(migrations))
	for _, m := range migrations {
		err := m.Migrate(t)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *BadgerStoreTxn) Commit(callback func(error)) error {
	return t.txn.Commit()
}

func (t *BadgerStoreTxn) Discard() {
	if t.txn != nil {
		t.txn.Discard()
	}
}

func (t *BadgerStoreTxn) Count(prefix []byte) (uint64, error) {
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

func (t *BadgerStoreTxn) Stream(prefix []byte, filter func(item *badger.Item) bool, callback func(list *pb.KVList) error) error {
	stream := t.db.NewStream()
	stream.Prefix = prefix // Leave nil for iteration over the whole DB.
	//stream.LogPrefix = "Badger.Streaming" // For identifying stream logs. Outputs to Logger.

	// ChooseKey is called concurrently for every key. If left nil, assumes true by default.
	stream.ChooseKey = filter

	// KeyToList is called concurrently for chosen keys. This can be used to convert
	// Badger data into custom key-values. If nil, uses stream.ToList, a default
	// implementation, which picks all valid key-values.
	stream.KeyToList = nil

	// -- End of optional settings.

	// Send is called serially, while Stream.Orchestrate is running.
	stream.Send = callback

	// Run the stream
	if err := stream.Orchestrate(context.Background()); err != nil {
		return err
	}
	return nil
}
