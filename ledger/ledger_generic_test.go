package ledger

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func addGenericType(t *testing.T, l *Ledger) (types.GenericKey, *types.GenericType) {
	t.Skip()
	key := types.GenericKey{}
	value := &types.GenericType{}
	if err := l.AddGenericType(key, value); err != nil {
		t.Fatal(err)
	}
	return key, value
}

func TestLedger_AddGenericType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addGenericType(t, l)
}

func TestLedger_GetGenericType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	k, v := addGenericType(t, l)
	value, err := l.GetGenericType(k)
	if err != nil {
		t.Fatal(err)
	}
	if v != value {
		t.Fatal(err)
	}
}

func TestLedger_DeleteGenericType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	k, _ := addGenericType(t, l)
	err := l.DeleteGenericType(k)
	if err != nil {
		t.Fatal(err)
	}
	_, err = l.GetGenericType(k)
	if err == nil {
		t.Fatal(err)
	}
}

func TestLedger_AddOrUpdateGenericType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addGenericType(t, l)

	k := types.GenericKey{}
	v := &types.GenericType{}
	err := l.AddOrUpdateGenericType(k, v)
	if err != nil {
		t.Fatal(err)
	}

}

func TestLedger_UpdateGenericType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	k := types.GenericKey{}
	v := &types.GenericType{}
	err := l.UpdateGenericType(k, v)
	if err == nil {
		t.Fatal(err)
	}

	k2, _ := addGenericType(t, l)
	v2 := &types.GenericType{}
	err = l.UpdateGenericType(k2, v2)
	if err == nil {
		t.Fatal(err)
	}
}

func TestLedger_GetGenericTypes(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addGenericType(t, l)
	addGenericType(t, l)
	err := l.GetGenericTypes(func(key types.GenericKey, value *types.GenericType) error {
		t.Log(key)
		t.Log(value)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_CountGenericType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addGenericType(t, l)
	addGenericType(t, l)
	num, err := l.CountGenericTypes()
	if err != nil {
		t.Fatal(err)
	}
	if num != 2 {
		t.Fatal("error count")
	}
	t.Log("account,", num)
}

func TestLedger_HasGenericType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	k, _ := addGenericType(t, l)
	b, err := l.HasGenericType(k)
	if err != nil || !b {
		t.Fatal(err)
	}
}
