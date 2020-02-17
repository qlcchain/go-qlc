package ledger

import (
	"fmt"
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

	blk := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)
	blk2 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk2, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	fmt.Println("result: ")
	fmt.Println(l.GetStateBlockConfirmed(blk.GetHash()))
	fmt.Println(l.GetStateBlockConfirmed(blk2.GetHash()))

	time.Sleep(1 * time.Second)
	blk3 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk3, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	blk4 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk4, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

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

	blk7 := mock.StateBlockWithoutWork()
	if err := l.UpdateStateBlock(blk7, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	if err := l.Close(); err != nil {
		t.Fatal(err)
	}

	fmt.Println(l.relation.BlocksCount())
	fmt.Println(l.relation.BlocksCountByType())

}

//func TestCache_Get(t *testing.T) {
//	teardownTestCase, l := setupTestCase(t)
//	defer teardownTestCase(t)
//
//}
