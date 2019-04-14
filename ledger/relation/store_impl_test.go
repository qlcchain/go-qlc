package relation

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/test/mock"
)

func TestRelation_CreateData(t *testing.T) {
	blk := mock.StateBlockWithoutWork()
	blk.Type = types.Send
	blk.Sender = []byte("1580000")
	blk.Receiver = []byte("1851111")
	dir := filepath.Join(config.QlcTestDataDir())
	cfg, err := config.DefaultConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRelation(cfg, nil)
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
	dir := filepath.Join(config.QlcTestDataDir())
	cfg, err := config.DefaultConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	eb := event.New()
	r1, err := NewRelation(cfg, eb)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := NewRelation(cfg, eb)
	if err != nil {
		t.Fatal(err)
	}
	r3, err := NewRelation(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r1.Close(); err != nil {
			t.Fatal(err)
		}
		if err := r3.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	t.Logf("r1, %p", r1)
	t.Logf("r2, %p", r2)
	t.Logf("r3, %p", r3)

	b := reflect.DeepEqual(r1, r2)
	if r1 == nil || r2 == nil || !b {
		t.Fatal("error")
	}
	b2 := reflect.DeepEqual(r1, r3)
	if r1 == nil || r3 == nil || b2 {
		t.Fatal("error")
	}
}
