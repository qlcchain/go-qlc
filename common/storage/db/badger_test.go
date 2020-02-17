package db

import (
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func TestNewBadgerStore(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "badger", uuid.New().String())
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Put([]byte{1, 2, 3}, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	r, err := db.Get([]byte{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestNewBadgerStore_Batch(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "badger", uuid.New().String())
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	batch := db.Batch(true)

	for i := 0; i < 100000; i++ {
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

	count := 0
	err = db.Iterator([]byte{byte(1)}, nil, func(k, v []byte) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(count)
}
