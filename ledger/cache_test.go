package ledger

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bluele/gcache"
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
	a.Purge()
	rm := a.GetALL(false)
	for k, v := range rm {
		t.Log(k, v)
	}
	t.Log(a.Len(false))
}

func TestCache_TimeSpan(t *testing.T) {
	cs := new(CacheStat)
	cs.Start = time.Now().UnixNano()
	time.Sleep(30 * time.Millisecond)
	cs.End = time.Now().UnixNano()
	fmt.Println(time.Now().Unix())
	fmt.Println(time.Now().UnixNano())

	span := cs.End - cs.Start
	fmt.Println(span / 1000000)

	fmt.Println(cs)
	fmt.Println(time.Unix(cs.Start/1000000000, 0).Format("2006-01-02 15:04:05"))

}

func TestCache_Iterator(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	l := NewLedger(cm.ConfigFile)

	defer func() {
		if err := l.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	mc := l.Cache()
	mc.Put([]byte{1, 2, 3, 4}, []byte{1, 4})
	mc.Put([]byte{1, 2, 3, 5}, []byte{1})
	mc.Put([]byte{1, 2, 3, 6}, []byte{4})
	kvs := mc.prefixIterator([]byte{1, 2, 3})
	for _, kv := range kvs {
		t.Log(kv.key, kv.value)
	}

}

//func TestCache_Put(t *testing.T) {
//	teardownTestCase, l := setupTestCase(t)
//	defer teardownTestCase(t)
//
//	timer := time.
//
//}
