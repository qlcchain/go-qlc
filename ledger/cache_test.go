package ledger

import (
	"fmt"
	"github.com/qlcchain/go-qlc/common/storage"
	"os"
	"path/filepath"
	"strconv"
	"sync"
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

func TestCache_Put(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	c := time.NewTicker(20 * time.Second)
	defer c.Stop()
InfLoop:
	for {
		select {
		case <-c.C:
			break InfLoop
		default:
			cache := l.Cache().GetCache()
			block := mock.StateBlockWithoutWork()
			k, _ := storage.GetKeyOfParts(storage.KeyPrefixBlock, block.GetHash())
			if err := cache.Put(k, block); err != nil {
				t.Fatal(err)
			}
		}
	}

	for _, cs := range l.cacheStats {
		span := strconv.FormatInt((cs.End-cs.Start)/1000000, 10) + "ms"
		fmt.Printf("index: %d, key: %d, span: %s  \n", cs.Index, cs.Key, span)
	}
}

func TestCache_PutConcurrency(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := time.NewTicker(10 * time.Second)
			defer c.Stop()
		InfLoop:
			for {
				select {
				case <-c.C:
					break InfLoop
				default:
					cache := l.Cache().GetCache()
					block := mock.StateBlockWithoutWork()
					k, _ := storage.GetKeyOfParts(storage.KeyPrefixBlock, block.GetHash())
					if err := cache.Put(k, block); err != nil {
						t.Fatal(err)
					}
				}
			}
		}()
	}
	wg.Wait()
	for _, cs := range l.cacheStats {
		span := strconv.FormatInt((cs.End-cs.Start)/1000000, 10) + "ms"
		fmt.Printf("index: %d, key: %d, span: %s  \n", cs.Index, cs.Key, span)
	}
}

func TestCache_Get(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	c := time.NewTicker(20 * time.Second)
	defer c.Stop()
InfLoop:
	for {
		select {
		case <-c.C:
			break InfLoop
		default:
			cache := l.Cache().GetCache()
			block := mock.StateBlockWithoutWork()
			k, _ := storage.GetKeyOfParts(storage.KeyPrefixBlock, block.GetHash())
			if err := cache.Put(k, block); err != nil {
				t.Fatal(err)
			}
			if _, _, err := l.Get(k); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, cs := range l.cacheStats {
		span := strconv.FormatInt((cs.End-cs.Start)/1000000, 10) + "ms"
		fmt.Printf("index: %d, key: %d, span: %s  \n", cs.Index, cs.Key, span)
	}
}

func TestCache_Get2(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	c := time.NewTicker(20 * time.Second)
	defer c.Stop()
	count := 0
InfLoop:
	for {
		select {
		case <-c.C:
			break InfLoop
		default:
			cache := l.Cache().GetCache()
			block := mock.StateBlockWithoutWork()
			k, _ := storage.GetKeyOfParts(storage.KeyPrefixBlock, block.GetHash())
			if err := cache.Put(k, block); err != nil {
				t.Fatal(err)
			}
			count++
			if count == 1000 {
				time.Sleep(20 * time.Millisecond)
				count = 0
			}
			if _, _, err := l.Get(k); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, cs := range l.cacheStats {
		span := strconv.FormatInt((cs.End-cs.Start)/1000000, 10) + "ms"
		fmt.Printf("index: %d, key: %d, span: %s  \n", cs.Index, cs.Key, span)
	}
}
