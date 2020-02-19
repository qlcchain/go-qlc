package ledger

import (
	"fmt"
	"github.com/bluele/gcache"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func setupCacheTestCase(t *testing.T) (func(t *testing.T), *Ledger) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	l := NewLedger(cm.ConfigFile)
	return func(t *testing.T) {
		//err := l.DBStore.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l
}

func TestNewCache(t *testing.T) {
	//cac := gcache.New(1000).LFU().Build()
	//arr := []byte{1, 2, 3}
	//cac.Set(arr, "123")

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	l := NewLedger(cm.ConfigFile)

	defer func() {
		if err := l.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	blk := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	blk2 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk2, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	fmt.Println("result: ")
	fmt.Println(l.GetStateBlockConfirmed(blk.GetHash()))
	fmt.Println(l.GetStateBlockConfirmed(blk2.GetHash()))

	blk3 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk3, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	blk4 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk4, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	blk5 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk5, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	blk6 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk6, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if err := l.DeleteStateBlock(blk6.GetHash(), l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	blk7 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk7, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	fmt.Println(l.relation.BlocksCount())
	fmt.Println(l.relation.BlocksCountByType())
	blk8 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk8, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
}

//func TestCache_Get(t *testing.T) {
//	teardownTestCase, l := setupTestCase(t)
//	defer teardownTestCase(t)
//
//}

func TestGcCache(t *testing.T) {
	amap := map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
	a := gcache.New(5).Build()
	for k, v := range amap {
		if err := a.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if err := a.Set(6, 6); err != nil {
		t.Fatal(err)
	}
	rm := a.GetALL(false)
	for k, v := range rm {
		t.Log(k, v)
	}
	t.Log(a.Len(false))
}
