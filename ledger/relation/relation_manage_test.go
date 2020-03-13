package relation

import (
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/mock"
)

func TestRelation_Relation(t *testing.T) {
	teardownTestCase, r := setupTestCase(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()
	r.Add(blk1)
	r.Add(blk2)
	r.Add(mock.StateBlockWithoutWork())
	r.Delete(blk1)
	r.Delete(blk2)
	time.Sleep(1 * time.Second)
	c, err := r.BlocksCount()
	if err != nil || c != 1 {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if err := r.EmptyStore(); err != nil {
		t.Fatal(err)
	}
	c, err = r.BlocksCount()
	if err != nil || c != 0 {
		t.Fatal(err)
	}
}

func TestRelation_RelationClose(t *testing.T) {
	teardownTestCase, r := setupTestCase(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()
	r.Add(blk1)
	r.Add(blk2)

	r.addChan <- mock.StateBlockWithoutWork()
	r.deleteChan <- blk1
}
