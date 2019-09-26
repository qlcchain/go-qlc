package ledger

import (
	"math/big"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestRepresentationCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	address := mock.Address()

	resp := types.Benefit{
		Balance: types.Balance{Int: big.NewInt(int64(500000000000))},
		Vote:    types.Balance{Int: big.NewInt(int64(500000000000))},
		Network: types.ZeroBalance,
		Oracle:  types.ZeroBalance,
		Storage: types.ZeroBalance,
		Total:   types.Balance{Int: big.NewInt(int64(1000000000000))},
	}

	diff := types.Benefit{
		Balance: types.Balance{Int: big.NewInt(int64(300000000000))},
		Vote:    types.Balance{Int: big.NewInt(int64(300000000000))},
		Network: types.ZeroBalance,
		Oracle:  types.ZeroBalance,
		Storage: types.ZeroBalance,
		Total:   types.Balance{Int: big.NewInt(int64(800000000000))},
	}

	if err := l.AddRepresentation(address, &resp); err != nil {
		t.Fatal(err)
	}
	ro, b := l.representCache.getFromMemory(address.String())
	if !b {
		t.Fatal("representation not found")
	}
	r1 := ro.(*types.Benefit)

	if !r1.Vote.Equal(resp.Vote) || !r1.Balance.Equal(resp.Balance) || !r1.Total.Equal(resp.Total) {
		t.Fatal("benefit not equal")
	}

	if err := l.SubRepresentation(address, &diff); err != nil {
		t.Fatal(err)
	}
	ro, b = l.representCache.getFromMemory(address.String())
	if !b {
		t.Fatal("representation not found")
	}
	r2 := ro.(*types.Benefit)

	if !r2.Vote.Equal(resp.Vote.Sub(diff.Vote)) || !r2.Balance.Equal(resp.Balance.Sub(diff.Balance)) || !r2.Total.Equal(resp.Total.Sub(diff.Total)) {
		t.Fatal("benefit not equal")
	}

	_, err := l.getRepresentation(address)
	if err != ErrRepresentationNotFound {
		t.Fatal(err)
	}

	txn := l.Store.NewTransaction(true)
	l.representCache.cacheToConfirmed(txn)
	txn.Commit(nil)

	rd, err := l.getRepresentation(address)
	if err != nil {
		t.Fatal(err)
	}

	if !r2.Vote.Equal(rd.Vote) || !r2.Balance.Equal(rd.Balance) || !r2.Total.Equal(rd.Total) {
		t.Fatal("benefit not equal")
	}

}
