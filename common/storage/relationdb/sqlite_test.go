package relationdb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), RelationDB) {
	t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "relation", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cfg, _ := cm.Load()
	db, err := NewSqliteDB(cfg)
	if err != nil {
		t.Fatal(err)
	}

	return func(t *testing.T) {
		//err := l.DBStore.Erase()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, db
}

func TestSqliteDB_Set(t *testing.T) {
	teardownTestCase, db := setupTestCase(t)
	defer teardownTestCase(t)

	block := mock.StateBlockWithoutWork()
	schame, key := block.TableSchema()
	sql := db.CreateTable(block.TableName(), schame, key)
	if _, err := db.Store().Exec(sql); err != nil {
		t.Fatal(err)
	}

	s, i := db.Set(block.TableName(), block.SetRelation())
	if _, err := db.Store().Exec(s, i...); err != nil {
		t.Fatal(err)
	}

	r := db.Delete(block.TableName(), block.RemoveRelation())
	if _, err := db.Store().Exec(r); err != nil {
		t.Fatal(err)
	}
}
