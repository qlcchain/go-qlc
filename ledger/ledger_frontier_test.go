package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
	"testing"
)

func addFrontier(t *testing.T, l *Ledger) *types.Frontier {
	frontier := new(types.Frontier)
	frontier.HeaderBlock = mock.Hash()
	frontier.OpenBlock = mock.Hash()
	if err := l.AddFrontier(frontier, l.cache.GetCache()); err != nil {
		t.Fatal()
	}
	return frontier
}

func TestLedger_AddFrontier(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addFrontier(t, l)
}

func TestLedger_GetFrontier(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	f := addFrontier(t, l)
	f, err := l.GetFrontier(f.HeaderBlock)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("frontier,", f)
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	f2, err := l.GetFrontier(f.HeaderBlock)
	if err != nil {
		t.Fatal(err)
	}
	if f.OpenBlock != f2.OpenBlock || f.HeaderBlock != f2.HeaderBlock {
		t.Fatal()
	}
	if _, err := l.GetFrontier(mock.Hash()); err == nil {
		t.Fatal(err)
	}
}

func TestLedger_GetAllFrontiers(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addFrontier(t, l)
	addFrontier(t, l)
	addFrontier(t, l)
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	c, err := l.CountFrontiers()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("frontier count,", c)

	fs, err := l.GetFrontiers()
	if err != nil {
		t.Fatal(err)
	}
	for index, f := range fs {
		t.Log("frontier", index, f)
	}
}

func TestLedger_DeleteFrontier(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	f := addFrontier(t, l)
	err := l.DeleteFrontier(f.HeaderBlock, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.GetFrontier(f.HeaderBlock); err != nil && err != ErrFrontierNotFound {
		t.Fatal(err)
	}
	t.Log("delete frontier success")
}
