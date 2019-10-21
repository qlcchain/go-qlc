package relation

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), string) {
	t.Log("setup test case")
	configDir := filepath.Join(config.QlcTestDataDir(), "relation", uuid.New().String())

	return func(t *testing.T) {
		t.Log("teardown test case")
		err := os.RemoveAll(configDir)
		if err != nil {
			t.Fatal(err)
		}
	}, configDir
}

func TestRelation_CreateData(t *testing.T) {
	teardownTestCase, dir := setupTestCase(t)
	defer teardownTestCase(t)

	blk := mock.StateBlockWithoutWork()
	blk.Type = types.Send
	blk.Sender = []byte("1580000")
	blk.Receiver = []byte("1851111")
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	r, err := NewRelation(cm.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	err = r.AddBlock(blk)
	if err != nil {
		t.Fatal(err)
	}
	b, err := r.AccountBlocks(blk.Address, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b)
	b, err = r.Blocks(10, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b)
	count, err := r.BlocksCount()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(count)
	g, err := r.BlocksCountByType()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(g)
	err = r.DeleteBlock(blk.GetHash())
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewRelation(t *testing.T) {
	teardownTestCase, dir := setupTestCase(t)
	defer teardownTestCase(t)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()

	r1, err := NewRelation(cm.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := NewRelation(cm.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r1.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	t.Logf("r1, %p", r1)
	t.Logf("r2, %p", r2)

	b := reflect.DeepEqual(r1, r2)
	if r1 == nil || r2 == nil || !b {
		t.Fatal("error")
	}
}
