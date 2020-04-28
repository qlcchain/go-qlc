package ledger

import (
	"github.com/qlcchain/go-qlc/mock"
	"testing"
)

func TestLedger_SetStorage(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	v := map[string]interface{}{string(mock.Hash().Bytes()): mock.Hash().Bytes()}
	if err := l.SetStorage(v); err != nil {
		t.Fatal(err)
	}
}

func TestLedger_SaveStorage(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	v := map[string]interface{}{string(mock.Hash().Bytes()): mock.Hash().Bytes()}
	if err := l.SaveStorage(v); err != nil {
		t.Fatal(err)
	}
	v2 := map[string]interface{}{string(mock.Hash().Bytes()): mock.Hash().Bytes()}
	if err := l.SaveStorage(v2, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	v3 := map[string]interface{}{string(mock.Hash().Bytes()): mock.StateBlock()}
	if err := l.SaveStorage(v3, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	if err := l.SaveStorageByConvert(mock.Hash().Bytes(), mock.Hash().Bytes(), l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
}
