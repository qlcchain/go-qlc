package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
)

func TestMigrationTo20(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "store", uuid.New().String())
	opts := badger.DefaultOptions(dir)

	opts.MaxTableSize = common.BadgerMaxTableSize
	opts.Logger = nil
	opts.ValueLogLoadingMode = options.FileIO
	_ = util.CreateDirIfNotExist(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tool := filepath.Join(filepath.Dir(dir), toolName)
	fmt.Println(tool)
	f, err := os.Create(tool)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	fmt.Println(err)
	_ = MigrationTo20(dir)
}
