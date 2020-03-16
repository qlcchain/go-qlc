package ledger

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/mock"
)

func addRepresentationWeight(t *testing.T, l *Ledger) *types.AccountMeta {
	address := mock.Address()
	ac := mock.AccountMeta(address)
	ac.CoinBalance = types.Balance{Int: big.NewInt(int64(1000))}
	ac.CoinVote = types.Balance{Int: big.NewInt(int64(1000))}
	benefit := &types.Benefit{
		Vote:    ac.CoinVote,
		Storage: ac.CoinStorage,
		Network: ac.CoinNetwork,
		Oracle:  ac.CoinOracle,
		Balance: ac.CoinBalance,
		Total:   ac.TotalBalance(),
	}

	err := l.AddRepresentation(address, benefit, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
	return ac
}

func TestLedger_AddRepresentationWeight(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ac := addRepresentationWeight(t, l)
	diff := &types.Benefit{
		Vote:    types.Balance{Int: big.NewInt(int64(10))},
		Storage: types.ZeroBalance,
		Network: types.ZeroBalance,
		Oracle:  types.ZeroBalance,
		Balance: types.Balance{Int: big.NewInt(int64(10))},
		Total:   types.Balance{Int: big.NewInt(int64(20))},
	}
	a, err := l.GetRepresentation(ac.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(a)
	err = l.AddRepresentation(ac.Address, diff, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
	a, err = l.GetRepresentation(ac.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(a)
	if !a.Total.Equal(ac.TotalBalance().Add(diff.Total)) {
		t.Fatal(a.Total, ac.TotalBalance(), diff.Total)
	}
}

func TestLedger_SubRepresentationWeight(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ac := addRepresentationWeight(t, l)
	diff := &types.Benefit{
		Vote:    types.Balance{Int: big.NewInt(int64(10))},
		Storage: types.ZeroBalance,
		Network: types.ZeroBalance,
		Oracle:  types.ZeroBalance,
		Balance: types.Balance{Int: big.NewInt(int64(10))},
		Total:   types.Balance{Int: big.NewInt(int64(20))},
	}

	err := l.SubRepresentation(ac.Address, diff, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
	a, err := l.GetRepresentation(ac.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(a)
	if !a.Total.Equal(ac.TotalBalance().Sub(diff.Total)) {
		t.Fatal(err)
	}
}

func TestLedger_GetRepresentations(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	am := addRepresentationWeight(t, l)
	addRepresentationWeight(t, l)

	time.Sleep(3 * time.Second)
	if _, err := l.GetRepresentation(am.Address); err != nil {
		t.Fatal(err)
	}

	err := l.GetRepresentations(func(address types.Address, benefit *types.Benefit) error {
		t.Log(address, benefit)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	count, err := l.CountRepresentations()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal("representation count error", count)
	}

	// deserialize error
	addr := mock.Address()
	k, err := storage.GetKeyOfParts(storage.KeyPrefixRepresentation, addr)
	if err != nil {
		t.Fatal()
	}
	d1 := make([]byte, 8)
	_ = random.Bytes(d1)
	if err := l.store.Put(k, d1); err != nil {
		t.Fatal(err)
	}
	if r, err := l.GetRepresentation(addr); err == nil {
		t.Fatal(err, r)
	}

	if err := l.GetRepresentations(func(addresses types.Address, benefit *types.Benefit) error {
		return nil
	}); err == nil {
		t.Fatal(err)
	}
}

func TestLedger_OnlineRepresentations(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	var addr []*types.Address
	for i := 0; i < 10; i++ {
		a1 := mock.Address()
		addr = append(addr, &a1)
	}

	err := l.SetOnlineRepresentations(addr)
	if err != nil {
		t.Fatal(err)
	}

	if addr2, err := l.GetOnlineRepresentations(); err == nil {
		if len(addr) != len(addr2) || len(addr2) != 10 {
			t.Fatal("invalid online rep size")
		}
		for i, v := range addr {
			if v.String() != addr2[i].String() {
				t.Fatal("invalid ")
			} else {
				t.Log(v.String())
			}
		}
	} else {
		t.Fatal(err)
	}
}

func TestLedger_SetOnlineRepresentations(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	var addr []*types.Address
	err := l.SetOnlineRepresentations(addr)
	if err != nil {
		t.Fatal(err)
	}
	if addr2, err := l.GetOnlineRepresentations(); err == nil {
		if len(addr2) != 0 {
			t.Fatal("invalid online rep")
		}
	} else {
		t.Fatal(err)
	}
}

func TestLedger_UpdateRepresentation(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addr := mock.Address()
	acc := mock.AccountMeta(addr)
	tm := mock.TokenMeta(addr)
	tm.Type = config.ChainToken()
	acc.Tokens = append(acc.Tokens, tm)
	if err := l.AddAccountMeta(acc, l.cache.GetCache()); err != nil {
		t.Fatal()
	}
	time.Sleep(2 * time.Second)
	if err := l.updateRepresentation(); err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddRepresentation(t *testing.T) {
	am := new(types.AccountMeta)
	fmt.Printf("=======%p \n", am)

	am = am.Clone()
	fmt.Printf("=======%p \n", am)

}
