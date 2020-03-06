package relation

import (
	"github.com/qlcchain/go-qlc/mock"
	"testing"
	"time"
)

func TestRelation_Relation(t *testing.T) {
	teardownTestCase, r := setupTestCase(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()
	r.Add(blk1)
	r.Add(blk2)
	r.Delete(blk1)

	time.Sleep(1 * time.Second)
	if err := r.EmptyStore(); err != nil {
		t.Fatal(err)
	}

}
