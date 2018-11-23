package db

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/dgraph-io/badger"
	badgerOpts "github.com/dgraph-io/badger/options"
)

const (
	badgerMaxOps = 10000
)

// BadgerStore represents a block lattice store backed by a badger database.
type BadgerStore struct {
	db *badger.DB
}

type BadgerStoreTxn struct {
	db  *badger.DB
	txn *badger.Txn
}

// NewBadgerStore initializes/opens a badger database in the given directory.
func NewBadgerStore() (Store, error) {
	dir := getBadgerDir()
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	opts.ValueLogLoadingMode = badgerOpts.FileIO

	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := os.Mkdir(dir, 0700); err != nil {
			return nil, err
		}
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerStore{db: db}, nil
}

func getBadgerDir() string {
	root, _ := os.Getwd()
	s := root[:strings.LastIndex(root, "go-qlc")]
	dir := path.Join(path.Join(path.Join(path.Join(s, "go-qlc"), "ledger"), "db"), "testdatabase")
	return dir
}

// Close closes the database
func (s *BadgerStore) Close() error {
	return s.db.Close()
}

// Purge purges any old/deleted keys from the database.
func (s *BadgerStore) Purge() error {
	return s.db.RunValueLogGC(0.5)
}

func (s *BadgerStore) View(fn func(txn StoreTxn) error) error {
	//t := &BadgerStoreTxn{txn: s.db.NewTransaction(true), db: s.db}
	//defer t.txn.Discard()
	//if err := fn(t); err != nil {
	//	return err
	//}
	//return nil
	return s.db.View(func(txn *badger.Txn) error {
		return fn(&BadgerStoreTxn{txn: txn, db: s.db})
	})
}

func (s *BadgerStore) Update(fn func(txn StoreTxn) error) error {
	t := &BadgerStoreTxn{txn: s.db.NewTransaction(true), db: s.db}
	defer t.txn.Discard()
	fmt.Println("new txn,", t)

	if err := fn(t); err != nil {
		fmt.Println("txn err,", err)
		return err
	}
	fmt.Println("commit txn,", t)
	return t.txn.Commit(nil)
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
