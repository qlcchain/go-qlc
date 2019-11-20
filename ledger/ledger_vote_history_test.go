package ledger

import (
	"testing"

	"github.com/qlcchain/go-qlc/mock"
)

func TestLedger_AddVoteHistory(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	hash := mock.Hash()
	address := mock.Address()
	err := l.AddVoteHistory(hash, address)
	if err != nil {
		t.Fatal(err)
	}

	has := l.HasVoteHistory(hash, address)
	if !has {
		t.Fatal()
	}
}

func TestLedger_CleanBlockVoteHistory(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	hash1 := mock.Hash()
	address1 := mock.Address()
	err := l.AddVoteHistory(hash1, address1)
	if err != nil {
		t.Fatal(err)
	}

	address2 := mock.Address()
	err = l.AddVoteHistory(hash1, address2)
	if err != nil {
		t.Fatal(err)
	}

	hash2 := mock.Hash()
	address3 := mock.Address()
	err = l.AddVoteHistory(hash2, address3)
	if err != nil {
		t.Fatal(err)
	}

	has1 := l.HasVoteHistory(hash1, address1)
	has2 := l.HasVoteHistory(hash1, address2)
	has3 := l.HasVoteHistory(hash2, address3)
	if !has1 || !has2 || !has3 {
		t.Fatal()
	}

	err = l.CleanBlockVoteHistory(hash1)
	if err != nil {
		t.Fatal()
	}

	has1 = l.HasVoteHistory(hash1, address1)
	has2 = l.HasVoteHistory(hash1, address2)
	if has1 || has2 {
		t.Fatal()
	}

	has3 = l.HasVoteHistory(hash2, address3)
	if !has3 {
		t.Fatal()
	}
}

func TestLedger_CleanAllVoteHistory(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	hash1 := mock.Hash()
	address1 := mock.Address()
	err := l.AddVoteHistory(hash1, address1)
	if err != nil {
		t.Fatal(err)
	}

	address2 := mock.Address()
	err = l.AddVoteHistory(hash1, address2)
	if err != nil {
		t.Fatal(err)
	}

	hash2 := mock.Hash()
	address3 := mock.Address()
	err = l.AddVoteHistory(hash2, address3)
	if err != nil {
		t.Fatal(err)
	}

	has1 := l.HasVoteHistory(hash1, address1)
	has2 := l.HasVoteHistory(hash1, address2)
	has3 := l.HasVoteHistory(hash2, address3)
	if !has1 || !has2 || !has3 {
		t.Fatal()
	}

	err = l.CleanAllVoteHistory()
	if err != nil {
		t.Fatal()
	}

	has1 = l.HasVoteHistory(hash1, address1)
	has2 = l.HasVoteHistory(hash1, address2)
	has3 = l.HasVoteHistory(hash2, address3)
	if has1 || has2 || has3 {
		t.Fatal()
	}
}
