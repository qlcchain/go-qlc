package ledger

import (
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"

	"github.com/qlcchain/go-qlc/mock"
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

func TestLedger_NewVMIterator(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	key := make([]byte, 0)
	key = append(key, storage.KeyPrefixVMStorage)
	key = append(key, contractaddress.NEP5PledgeAddress.Bytes()...)
	key = append(key, []byte{1, 2, 3, 4}...)

	if err := l.Put(key, []byte{10, 20, 30, 40}); err != nil {
		t.Fatal(err)
	}

	key2 := make([]byte, 0)
	key2 = append(key2, storage.KeyPrefixVMStorage)
	key2 = append(key2, contractaddress.NEP5PledgeAddress.Bytes()...)
	key2 = append(key2, []byte{1, 5, 6, 7}...)

	if err := l.Put(key2, []byte{11, 21, 31, 41}); err != nil {
		t.Fatal(err)
	}

	iter := l.NewVMIterator(&contractaddress.NEP5PledgeAddress)
	count := 0
	if err := iter.Next([]byte{1}, func(key []byte, value []byte) error {
		count++
		fmt.Println(key)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal()
	}

}
