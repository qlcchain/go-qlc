package ledger

import (
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/mock"
)

func addPending(t *testing.T, l *Ledger) (pendingkey types.PendingKey, pendinginfo types.PendingInfo) {
	address := mock.Address()
	hash := mock.Hash()

	i, _ := random.Intn(math.MaxUint32)
	balance := types.Balance{Int: big.NewInt(int64(i))}
	pendinginfo = types.PendingInfo{
		Source: address,
		Amount: balance,
		Type:   mock.Hash(),
	}
	pendingkey = types.PendingKey{Address: address, Hash: hash}
	t.Log(pendinginfo)
	err := l.AddPending(&pendingkey, &pendinginfo, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestLedger_AddPending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pk, pv := addPending(t, l)
	if _, err := l.rcache.GetAccountPending(pk.Address, pv.Type); err == nil {
		t.Fatal(err)
	}
	a, err := l.PendingAmount(pk.Address, pv.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(a)
	b, err := l.rcache.GetAccountPending(pk.Address, pv.Type)
	if err != nil {
		t.Fatal(err)
	}
	if !a.Equal(b) {
		t.Fatal("balance not equal")
	}
	// add pending again
	pk.Hash = mock.Hash()
	err = l.AddPending(&pk, &pv, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_GetPending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addPending(t, l)
	pendingkey, _ := addPending(t, l)
	p, err := l.GetPending(&pendingkey)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("pending,", p)
	time.Sleep(2 * time.Second)
	if _, err := l.GetPending(&pendingkey); err != nil {
		t.Fatal(err)
	}
	pendingkey2 := types.PendingKey{Address: mock.Address(), Hash: mock.Hash()}
	if _, err := l.GetPending(&pendingkey2); err == nil {
		t.Fatal()
	}
	count := 0
	err = l.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		t.Log(pendingKey, pendingInfo)
		count = count + 1
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal("pending count error", count)
	}

	// Deserialize error
	key := &types.PendingKey{
		Address: mock.Address(),
		Hash:    mock.Hash(),
	}
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPending, key)
	if err != nil {
		t.Fatal()
	}
	d1 := make([]byte, 10)
	_ = random.Bytes(d1)
	if err := l.store.Put(k, d1); err != nil {
		t.Fatal(err)
	}
	if _, err := l.GetPending(key); err == nil {
		t.Fatal(err)
	}

	if err := l.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		return nil
	}); err == nil {
		t.Fatal(err)
	}
}

func TestLedger_DeletePending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pendingkey, _ := addPending(t, l)

	err := l.DeletePending(&pendingkey, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.GetPending(&pendingkey); err != nil && err != ErrPendingNotFound {
		t.Fatal(err)
	}
	t.Log("delete pending success")
}

func TestLedger_SearchPending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	address := mock.Address()

	pendingkeys := make([]*types.PendingKey, 0)
	for idx := 0; idx < 10; idx++ {
		hash := mock.Hash()
		i, _ := random.Intn(math.MaxUint32)
		balance := types.Balance{Int: big.NewInt(int64(i))}
		v := &types.PendingInfo{
			Source: address,
			Amount: balance,
			Type:   mock.Hash(),
		}
		k := &types.PendingKey{Address: address, Hash: hash}
		err := l.AddPending(k, v, l.cache.GetCache())
		if err != nil {
			t.Fatal(err)
		}
		pendingkeys = append(pendingkeys, k)
		//t.Log(idx, util.ToString(k), util.ToString(v))
	}
	//t.Log("build cache done")

	time.Sleep(3 * time.Second)
	counter := 0
	err := l.GetPendingsByAddress(address, func(key *types.PendingKey, value *types.PendingInfo) error {
		t.Log(counter, util.ToString(key), util.ToString(value))
		counter++
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if counter != 10 {
		t.Fatal("invalid", counter)
	}
	count1 := 0
	err = l.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		count1++
		return nil
	})
	if count1 != 10 {
		t.Fatal("invalid", count1)
	}

}

func TestLedger_PendingAmount(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pendingkey, pendinginfo := addPending(t, l)
	time.Sleep(2 * time.Second)
	if _, err := l.rcache.GetAccountPending(pendingkey.Address, pendinginfo.Type); err == nil {
		t.Fatal("pending should not found")
	}
	a, err := l.PendingAmount(pendingkey.Address, pendinginfo.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(a)
	b, err := l.rcache.GetAccountPending(pendingkey.Address, pendinginfo.Type)
	if err != nil {
		t.Fatal()
	}
	if !a.Equal(b) {
		t.Fatal("balance should equal")
	}
}
