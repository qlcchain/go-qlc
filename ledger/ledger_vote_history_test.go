package ledger

import (
	"github.com/qlcchain/go-qlc/mock"
	"testing"
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

func TestLedger_CleanVoteHistoryTimeout(t *testing.T) {

}