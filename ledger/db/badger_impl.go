package db

import (
	"os"

	"github.com/dgraph-io/badger"
	badgerOpts "github.com/dgraph-io/badger/options"
)

// BadgerStore represents a block lattice store backed by a badger database.
type BadgerStore struct {
	db *badger.DB
}

type BadgerStoreTxn struct {
	txn *badger.Txn
	db  *badger.DB
	ops uint64
}

// NewBadgerStore initializes/opens a badger database in the given directory.
func NewBadgerStore(dir string) (*BadgerStore, error) {
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

// Close closes the database
func (s *BadgerStore) Close() error {
	return s.db.Close()
}

// Purge purges any old/deleted keys from the database.
func (s *BadgerStore) Purge() error {
	return nil
}

func (s *BadgerStore) View(fn func(txn StoreTxn) error) error {
	return nil
}

func (s *BadgerStore) Update(fn func(txn StoreTxn) error) error {
	return nil
}
