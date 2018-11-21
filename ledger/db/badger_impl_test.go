package db

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestBadgerStoreTxn_Set(t *testing.T) {
	db, err := NewBadgerStore()
	defer db.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn StoreTxn) error {
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
	db, err := NewBadgerStore()
	defer db.Close()

	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn StoreTxn) error {
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
	db, err := NewBadgerStore()
	defer db.Close()

	if err != nil {
		t.Fatal(err)
	}
	err = db.View(func(txn StoreTxn) error {
		block := new(types.StateBlock)
		key := block.GetHash()
		blk := new(types.StateBlock)
		err := txn.Get(key[:], func(val []byte, b byte) error {
			if _, err = blk.UnmarshalMsg(val); err != nil {
				t.Fatal(err)
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
	db, err := NewBadgerStore()
	defer db.Close()

	if err != nil {
		t.Fatal(err)
	}
	err = db.View(func(txn StoreTxn) error {
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
	db, err := NewBadgerStore()
	defer db.Close()

	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn StoreTxn) error {
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
