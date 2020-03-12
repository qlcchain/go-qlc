package ledger

import (
	"testing"

	"github.com/qlcchain/go-qlc/mock"
)

func TestLedger_UncheckedSyncBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := mock.StateBlockWithoutWork()
	if err := l.AddUncheckedSyncBlock(block.Previous, block); err != nil {
		t.Fatal(err)
	}

	if r, err := l.HasUncheckedSyncBlock(block.Previous); err != nil || !r {
		t.Fatal(err)
	}

	blk, err := l.GetUncheckedSyncBlock(block.Previous)
	if err != nil {
		t.Fatal(err)
	}

	r, err := l.CountUncheckedSyncBlocks()
	if err != nil || r != 1 {
		t.Fatal(err)
	}

	if block.GetHash() != blk.GetHash() {
		t.Fatal("hash err")
	}

	if err := l.DeleteUncheckedSyncBlock(block.Previous); err != nil {
		t.Fatal(err)
	}

	_, err = l.GetUncheckedSyncBlock(block.Previous)
	if err == nil {
		t.Fatal(err)
	}
}

func TestLedger_UnconfirmedSyncBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := mock.StateBlockWithoutWork()
	hash := block.GetHash()
	if err := l.AddUnconfirmedSyncBlock(hash, block); err != nil {
		t.Fatal(err)
	}

	if r, err := l.CountUnconfirmedSyncBlocks(); err != nil || r != 1 {
		t.Fatal(err)
	}

	if has, err := l.HasUnconfirmedSyncBlock(hash); !has {
		t.Fatal(err)
	}

	blk, err := l.GetUnconfirmedSyncBlock(hash)
	if err != nil {
		t.Fatal(err)
	}

	if hash != blk.GetHash() {
		t.Fatal("hash err")
	}

	if err := l.DeleteUnconfirmedSyncBlock(hash); err != nil {
		t.Fatal(err)
	}

	if has, err := l.HasUnconfirmedSyncBlock(hash); has {
		t.Fatal(err)
	}

	_, err = l.GetUnconfirmedSyncBlock(hash)
	if err == nil {
		t.Fatal(err)
	}
}

func TestLedger_CleanSyncCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := mock.StateBlockWithoutWork()
	if err := l.AddUncheckedSyncBlock(block.Previous, block); err != nil {
		t.Fatal(err)
	}

	block2 := mock.StateBlockWithoutWork()
	hash2 := block2.GetHash()
	if err := l.AddUnconfirmedSyncBlock(hash2, block); err != nil {
		t.Fatal(err)
	}

	count := 0
	l.WalkSyncCache(func(kind byte, key []byte) {
		count = count + 1
	})
	if count != 2 {
		t.Fatal()
	}

	l.CleanSyncCache()
	if _, err := l.GetUncheckedSyncBlock(block.Previous); err == nil {
		t.Fatal(err)
	}

	if _, err := l.GetUnconfirmedSyncBlock(hash2); err == nil {
		t.Fatal(err)
	}
}
