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

	blk, err := l.GetUncheckedSyncBlock(block.Previous)
	if err != nil {
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
