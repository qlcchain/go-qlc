package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), storage.Store) {
	t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "store", uuid.New().String())
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	return func(t *testing.T) {
		//err := l.DBStore.Erase()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, db
}

func TestNewBadgerStore(t *testing.T) {
	teardownTestCase, db := setupTestCase(t)
	defer teardownTestCase(t)

	if err := db.Put([]byte{1, 2, 3}, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	r, err := db.Get([]byte{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	prefix := []byte{1}
	if err := db.Drop(prefix); err != nil {
		t.Fatal(err)
	}
	if i, err := db.Count(prefix); i != 0 || err != nil {
		t.Fatal(i, err)
	}
	t.Log(r)
}

func TestNewBadgerStore_BadgerTransaction(t *testing.T) {
	teardownTestCase, db := setupTestCase(t)
	defer teardownTestCase(t)

	batch := db.Batch(true)
	for i := 0; i < 10000; i++ {
		block := mock.StateBlockWithoutWork()
		k := make([]byte, 0)
		k = append(k, byte(1))
		k = append(k, block.GetHash().Bytes()...)
		v, _ := block.Serialize()
		if err := batch.Put(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.PutBatch(batch); err != nil {
		t.Fatal(err)
	}

	prefix := []byte{1}
	batch2 := db.Batch(true)
	count := 0
	err := batch2.Iterator(prefix, nil, func(k, v []byte) error {
		count++
		return nil
	})
	if err := db.PutBatch(batch2); err != nil {
		t.Fatal(err)
	}
	if err != nil || count != 10000 {
		t.Fatal(err)
	}

	count = 0
	err = db.Iterator(prefix, []byte{1, 100}, func(k, v []byte) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(count)

	batchDrop := db.Batch(true)
	if err := batchDrop.Drop(prefix); err != nil {
		t.Fatal(err)
	}
	if err := db.PutBatch(batchDrop); err != nil {
		t.Fatal(err)
	}

	if i, err := db.Count(prefix); i != 0 || err != nil {
		t.Fatal(i, err)
	}
}

func TestNewBadgerStore_BadgerWriteBatch(t *testing.T) {
	teardownTestCase, db := setupTestCase(t)
	defer teardownTestCase(t)

	batch := db.Batch(false)
	for i := 0; i < 1000; i++ {
		block := mock.StateBlockWithoutWork()
		k := make([]byte, 0)
		k = append(k, byte(1))
		k = append(k, block.GetHash().Bytes()...)
		v, _ := block.Serialize()
		if err := batch.Put(k, v); err != nil {
			t.Fatal(err)
		}
	}

	blk := mock.StateBlockWithoutWork()
	k := make([]byte, 0)
	k = append(k, byte(1))
	k = append(k, blk.GetHash().Bytes()...)
	v, _ := blk.Serialize()
	if err := batch.Put(k, v); err != nil {
		t.Fatal(err)
	}

	if err := db.PutBatch(batch); err != nil {
		t.Fatal(err)
	}
	prefix := []byte{1}
	if i, err := db.Count(prefix); i != 1001 || err != nil {
		t.Fatal(err)
	}

	batchDel := db.Batch(false)
	if err := batchDel.Delete(k); err != nil {
		t.Fatal(err)
	}
	if err := db.PutBatch(batchDel); err != nil {
		t.Fatal(err)
	}

	if i, err := db.Count(prefix); i != 1000 || err != nil {
		t.Fatal(err)
	}
}

func TestNewBadgerStore_BadgerBatch(t *testing.T) {
	teardownTestCase, db := setupTestCase(t)
	defer teardownTestCase(t)

	if err := db.BatchWrite(true, func(batch storage.Batch) error {
		if err := db.Put([]byte{1, 2, 3}, []byte{4, 5, 6}); err != nil {
			t.Fatal(err)
		}
		if err := db.Put([]byte{1, 2, 4}, []byte{4, 5, 6}); err != nil {
			t.Fatal(err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	prefix := []byte{1}
	if i, err := db.Count(prefix); i != 2 || err != nil {
		t.Fatal(err)
	}

	if err := db.BatchWrite(false, func(batch storage.Batch) error {
		if err := db.Put([]byte{1, 2, 5}, []byte{4, 5, 6}); err != nil {
			t.Fatal(err)
		}
		if err := db.Put([]byte{1, 2, 6}, []byte{4, 5, 6}); err != nil {
			t.Fatal(err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if i, err := db.Count(prefix); i != 4 || err != nil {
		t.Fatal(err)
	}
}

func TestBadgerStore_Get(t *testing.T) {
	teardownTestCase, db := setupTestCase(t)
	defer teardownTestCase(t)

	key := []byte{1, 2, 3}
	if err := db.Put(key, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if b, err := db.Has(key); err != nil || !b {
		t.Fatal()
	}
	if err := db.Delete(key); err != nil {
		t.Fatal(err)
	}
	if b, _ := db.Has(key); b {
		t.Fatal()
	}
}

func TestBadgerStore_Count(t *testing.T) {
	teardownTestCase, db := setupTestCase(t)
	defer teardownTestCase(t)

	if err := db.Put([]byte{1, 2, 3}, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if err := db.Put([]byte{1, 2, 4}, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if err := db.Put([]byte{1, 2, 5}, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if c, err := db.Count([]byte{1, 2}); c != 3 || err != nil {
		t.Fatal(err)
	}
}

func TestBadgerStore_Upgrade(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "badger", uuid.New().String())
	opts := badger.DefaultOptions(dir)
	_ = util.CreateDirIfNotExist(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	store := &BadgerStore{db: db}
	defer func() {
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}()

	hash := mock.Hash()
	k, _ := storage.GetKeyOfParts(storage.KeyPrefixUncheckedBlockPrevious, hash)

	blk := mock.StateBlockWithoutWork()
	val, _ := blk.Serialize()
	txn := db.NewTransaction(true)
	if err := txn.SetEntry(&badger.Entry{
		Key:      k,
		Value:    val,
		UserMeta: 0,
	}); err != nil {
		t.Fatal(err)
	}
	if err := txn.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := store.Upgrade(1); err != nil {
		t.Fatal(err)
	}

}
