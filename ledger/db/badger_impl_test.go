package db

import (
	"os"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

var db Store

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")
	dir := util.QlcDir("test", "badger")
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
		val, _ := blk.MarshalMsg(nil)
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
		val, _ := blk.MarshalMsg(nil)
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
		val, _ := blk.MarshalMsg(nil)
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
			if _, err2 := blk.UnmarshalMsg(val); err2 != nil {
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
