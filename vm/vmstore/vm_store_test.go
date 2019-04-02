/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package vmstore

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/ledger"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/test/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *VMContext) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "vm", uuid.New().String())
	_ = os.RemoveAll(dir)
	l := ledger.NewLedger(dir)

	v := NewVMContext(l)

	return func(t *testing.T) {
		//err := v.db.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, v
}

func TestLedger_Storage(t *testing.T) {
	teardownTestCase, context := setupTestCase(t)
	defer teardownTestCase(t)

	prefix := mock.Hash()
	key := []byte{10, 20, 30}
	value := []byte{10, 20, 30, 40}
	if err := context.SetStorage(prefix[:], key, value); err != nil {
		t.Fatal(err)
	}
	v, err := context.GetStorage(prefix[:], key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(value, v) {
		t.Fatal("err store")
	}

}
