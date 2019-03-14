package ledger

import (
	"bytes"
	"testing"

	"github.com/qlcchain/go-qlc/test/mock"
)

func TestLedger_Storage(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	prefix := mock.Hash()
	key := []byte{10, 20, 30}
	value := []byte{10, 20, 30, 40}
	if err := l.SetStorage(prefix[:], key, value); err != nil {
		t.Fatal(err)
	}
	v, err := l.GetStorage(prefix[:], key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(value, v) {
		t.Fatal("err store")
	}

}
