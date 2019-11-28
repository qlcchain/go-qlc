package db

import (
	"io"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
)

// Store is an interface that all stores need to implement.
type Store interface {
	io.Closer
	Purge() error
	Erase() error
	ViewInTx(fn func(txn StoreTxn) error) error
	UpdateInTx(fn func(txn StoreTxn) error) error
	NewTransaction(update bool) *BadgerStoreTxn
	NewWriteBatch() *BadgerStoreBatch
	UpdateInBatch(fn func(batch StoreBatch) error) error
	Size() (int64, int64)
}

type StoreTxn interface {
	Set(key, val []byte) error
	SetWithMeta(key, val []byte, meta byte) error
	Get(key []byte, fn func([]byte, byte) error) error
	Delete(key []byte) error
	Iterator(pre byte, fn func([]byte, []byte, byte) error) error
	PrefixIterator(prefix []byte, fn func([]byte, []byte, byte) error) error
	RangeIterator(startKey []byte, endKey []byte, fn func([]byte, []byte, byte) error) error
	Commit() error
	Discard()
	Drop(prefix []byte) error
	Upgrade(migrations []Migration) error
	Count(prefix []byte) (uint64, error)
	Stream(prefix []byte, filter func(item *badger.Item) bool, callback func(list *pb.KVList) error) error
}

type StoreBatch interface {
	Set(key, val []byte) error
	SetWithMeta(key, val []byte, meta byte) error
	Delete(key []byte) error
	Cancel()
	Flush() error
}
