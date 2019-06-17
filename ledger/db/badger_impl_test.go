package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
)

var db Store

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")
	dir := filepath.Join(config.QlcTestDataDir(), "badger")
	var err error
	db, err = NewBadgerStore(dir)

	if err != nil {
		t.Fatal(err)
	}
	return func(t *testing.T) {
		t.Log("teardown test case")
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestBadgerStoreTxn_Set(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	err := db.UpdateInTx(func(txn StoreTxn) error {
		blk := new(types.StateBlock)
		key := blk.GetHash()
		val, _ := blk.Serialize()
		if err := txn.Set(key[:], val); err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadgerStoreTxn_SetWithMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	err := db.UpdateInTx(func(txn StoreTxn) error {
		blk := new(types.StateBlock)
		key := blk.GetHash()
		val, _ := blk.Serialize()
		if err := txn.SetWithMeta(key[:], val, 0); err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadgerStoreTxn_Get(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	err := db.UpdateInTx(func(txn StoreTxn) error {
		blk := new(types.StateBlock)
		key := blk.GetHash()
		val, _ := blk.Serialize()
		if err := txn.SetWithMeta(key[:], val, 0); err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.ViewInTx(func(txn StoreTxn) error {
		block := new(types.StateBlock)
		key := block.GetHash()
		blk := new(types.StateBlock)
		err := txn.Get(key[:], func(val []byte, b byte) error {
			if err2 := blk.Deserialize(val); err2 != nil {
				t.Fatal(err2)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log(blk)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadgerStoreTxn_Iterator(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	err := db.ViewInTx(func(txn StoreTxn) error {
		err := txn.Iterator(206, func(key []byte, val []byte, b byte) error {
			t.Log(key)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadgerStoreTxn_KeyIterator(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	keys := [][]byte{[]byte{1, 206}, []byte{2, 207}, []byte{1, 208}}
	vals := keys
	err := db.UpdateInTx(func(txn StoreTxn) error {
		for i, key := range keys {
			if err := txn.SetWithMeta(key, vals[i], 0); err != nil {
				t.Fatal(err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	expectKeyCnt := 0
	err = db.ViewInTx(func(txn StoreTxn) error {
		err := txn.KeyIterator([]byte{1}, func(key []byte) error {
			t.Log(key)
			expectKeyCnt++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if expectKeyCnt != 2 {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadgerStoreTxn_Delete(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	err := db.UpdateInTx(func(txn StoreTxn) error {
		blk := new(types.StateBlock)
		key := blk.GetHash()
		if err := txn.Delete(key[:]); err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadgerStoreTxn_Drop(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	txn := db.NewTransaction(true)
	blk := new(types.StateBlock)
	key := blk.GetHash()
	val, _ := blk.Serialize()
	if err := txn.Set(key[:], val); err != nil {
		t.Fatal(err)
	}
	if err := txn.Commit(nil); err != nil {
		t.Fatal(err)
	}

	txn = db.NewTransaction(true)
	if err := txn.Drop(nil); err != nil {
		t.Fatal(err)
	}
	txn.Commit(nil)
	txn = db.NewTransaction(false)
	err := txn.Get(key[:], func(bytes []byte, b byte) error {
		return nil
	})
	if err != badger.ErrKeyNotFound {
		t.Fatal(err)
	}

}
