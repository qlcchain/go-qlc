package ledger

import (
	"math"
	"math/big"
	"testing"

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
	err := l.AddPending(&pendingkey, &pendinginfo)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestLedger_AddPending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addPending(t, l)
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
		t.Fatal("pending count error")
	}
}

func TestLedger_DeletePending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pendingkey, _ := addPending(t, l)

	err := l.DeletePending(&pendingkey)
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
		err := l.AddPending(k, v)
		if err != nil {
			t.Fatal(err)
		}
		pendingkeys = append(pendingkeys, k)
		//t.Log(idx, util.ToString(k), util.ToString(v))
	}
	//t.Log("build cache done")

	counter := 0
	err := l.SearchPending(address, func(key *types.PendingKey, value *types.PendingInfo) error {
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

	if err := l.UpdatePending(pendingkeys[0], types.PendingUsed); err != nil {
		t.Fatal(err)
	}

	count2 := 0
	err = l.SearchPending(address, func(key *types.PendingKey, value *types.PendingInfo) error {
		t.Log(count2, util.ToString(key), util.ToString(value))
		count2++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count2 != 9 {
		t.Fatal("invalid", count2)
	}
	count3 := 0
	err = l.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		count3++
		return nil
	})
	if count3 != 9 {
		t.Fatal("invalid", count3)
	}
	count4 := 0
	err = l.SearchAllKindPending(address, func(key *types.PendingKey, value *types.PendingInfo, kind types.PendingKind) error {
		t.Log(count4, util.ToString(key), util.ToString(value), kind)
		count4++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count4 != 10 {
		t.Fatal("invalid", count4)
	}

}

func TestLedger_Pending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pendingkey, _ := addPending(t, l)
	pending, err := l.Pending(pendingkey.Address)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range pending {
		t.Log(k, v)
	}
}

func TestLedger_TokenPending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pendingkey, pendinginfo := addPending(t, l)
	pending, err := l.TokenPending(pendingkey.Address, pendinginfo.Type)
	if err != nil && err != ErrPendingNotFound {
		t.Fatal(err)
	}
	for k, v := range pending {
		t.Log(k, v)
	}
}
