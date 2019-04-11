package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/test/mock"
)

func TestDBSQL_Create(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "sqlite3")
	d, err := NewSQLDB(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := d.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	condition := make(map[Column]interface{})
	condition[ColumnHash] = mock.Hash().String()
	condition[ColumnTimestamp] = time.Now().Unix()
	condition[ColumnType] = types.Open.String()
	condition[ColumnAddress] = mock.Address().String()
	if err := d.Create(TableBlockHash, condition); err != nil {
		t.Fatal(err)
	}
	var b []blocksHash
	if err := d.Read(TableBlockHash, condition, -1, -1, ColumnNoNeed, &b); err != nil {
		t.Fatal(err)
	}
	var i int
	if err := d.Count(TableBlockHash, &i); err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Fatal(err)
	}
	if err := d.Delete(TableBlockHash, condition); err != nil {
		t.Fatal(err)
	}
	if err := d.Count(TableBlockHash, &i); err != nil {
		t.Fatal(err)
	}
	if i != 0 {
		t.Fatal(err)
	}
}

type blocksHash struct {
	Id        int64
	Hash      string
	Type      string
	Address   string
	Timestamp int64
}
