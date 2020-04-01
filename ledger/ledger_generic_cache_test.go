package ledger

import (
	"testing"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/types"
)

func addGenericTypeC(t *testing.T, l *Ledger) (*types.GenericKeyC, *types.GenericTypeC) {
	key := &types.GenericKeyC{Key: uuid.New().String()}
	value := &types.GenericTypeC{Value: uuid.New().String()}
	if err := l.AddGenericTypeC(key, value, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	return key, value
}

func TestLedger_AddGenericTypeC(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addGenericTypeC(t, l)
}

func TestLedger_GetGenericTypeC(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	k, v := addGenericTypeC(t, l)
	value, err := l.GetGenericTypeC(k)
	if err != nil {
		t.Fatal(err)
	}
	if v.Value != value.Value {
		t.Fatal(err, v, value)
	}
	t.Log(value)
}

func TestLedger_DeleteGenericTypeC(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	k, _ := addGenericTypeC(t, l)
	err := l.DeleteGenericTypeC(k, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
	_, err = l.GetGenericTypeC(k)
	if err == nil {
		t.Fatal(err)
	}
}

func TestLedger_AddOrUpdateGenericTypeC(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addGenericTypeC(t, l)

	k := &types.GenericKeyC{}
	v := &types.GenericTypeC{}
	err := l.AddOrUpdateGenericTypeC(k, v, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_UpdateGenericTypeC(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	k := &types.GenericKeyC{}
	v := &types.GenericTypeC{}
	err := l.UpdateGenericTypeC(k, v, l.cache.GetCache())
	if err == nil {
		t.Fatal(err)
	}

	k2, _ := addGenericTypeC(t, l)
	v2 := &types.GenericTypeC{}
	err = l.UpdateGenericTypeC(k2, v2, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_GetGenericTypeCs(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addGenericTypeC(t, l)
	addGenericTypeC(t, l)
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	err := l.GetGenericTypeCs(func(key *types.GenericKeyC, value *types.GenericTypeC) error {
		t.Log(key)
		t.Log(value)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_CountGenericTypeC(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addGenericTypeC(t, l)
	addGenericTypeC(t, l)
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	num, err := l.CountGenericTypeCs()
	if err != nil {
		t.Fatal(err)
	}
	if num != 2 {
		t.Fatal("error count", num)
	}
	t.Log("account,", num)
}

func TestLedger_HasGenericTypeC(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	k, _ := addGenericTypeC(t, l)
	b, err := l.HasGenericTypeC(k)
	if err != nil || !b {
		t.Fatal(err)
	}
}
