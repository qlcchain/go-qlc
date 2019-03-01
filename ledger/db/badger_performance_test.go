package db

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dgraph-io/badger"
	badgerOpts "github.com/dgraph-io/badger/options"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
)

var (
	dir_checkblock = filepath.Join(config.QlcTestDataDir(), "checkblock")
	dir_testdb     = filepath.Join(config.QlcTestDataDir(), "badger_pref")
)

func generateBlock() *types.StateBlock {
	var blk types.StateBlock
	random.Bytes(blk.Previous[:])
	random.Bytes(blk.Representative[:])
	random.Bytes(blk.Address[:])
	random.Bytes(blk.Signature[:])
	random.Bytes(blk.Link[:])
	random.Bytes(blk.Signature[:])
	random.Bytes(blk.Token[:])
	return &blk
}

func NewBadger() (*badger.DB, error) {
	opts := badger.DefaultOptions
	opts.Dir = dir_testdb
	opts.ValueDir = dir_testdb
	opts.ValueLogLoadingMode = badgerOpts.FileIO

	_ = util.CreateDirIfNotExist(dir_testdb)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestBadgerWrite(t *testing.T) {
	db, err := NewBadger()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	fmt.Println(db.Size())
}

func TestBadgerPerformance_AddBlocks(t *testing.T) {
	db, err := NewBadger()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const m = 0
	const n = 10000

	start := time.Now()
	for k := 0; k < m; k++ {
		db.Update(func(txn *badger.Txn) error {
			for i := 0; i < n; i++ {
				blk := generateBlock()
				key := blk.GetHash()
				val, _ := blk.Serialize()
				if err := txn.Set(key[:], val); err != nil {
					t.Fatal(err)
				}
			}
			return nil
		})
	}

	end := time.Now()
	fmt.Printf("write benchmark: %d op/s ,time span,%f \n", int((m*n)/end.Sub(start).Seconds()), end.Sub(start).Seconds())
}

func TestBadgerPerformance_AddBlocksByGoroutine(t *testing.T) {
	db, err := NewBadger()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const m = 0
	const n = 10000
	var wg sync.WaitGroup
	start := time.Now()
	for k := 0; k < m; k++ {
		wg.Add(1)
		go func() {
			db.Update(func(txn *badger.Txn) error {
				for i := 0; i < n; i++ {
					blk := generateBlock()
					key := blk.GetHash()
					val, _ := blk.MarshalMsg(nil)
					if err := txn.Set(key[:], val); err != nil {
						t.Fatal(err)
					}
				}
				return nil
			})
			wg.Done()
		}()
	}
	wg.Wait()
	end := time.Now()
	fmt.Printf("write benchmark: %d op/s\n", int((m*n)/end.Sub(start).Seconds()))
}

func TestBadgerPerformance_AddBlocksByGoroutine2(t *testing.T) {
	var wg sync.WaitGroup

	const m = 0
	const n = 10000

	start := time.Now()

	for k := 0; k < m; k++ {
		wg.Add(1)
		go func() {
			db, err := NewBadger()
			if err != nil {
				t.Fatal(err)
			}
			db.Update(func(txn *badger.Txn) error {
				for i := 0; i < n; i++ {
					blk := generateBlock()
					key := blk.GetHash()
					val, _ := blk.MarshalMsg(nil)
					if err := txn.Set(key[:], val); err != nil {
						t.Fatal(err)
					}
				}
				return nil
			})
			db.Close()
			wg.Done()
		}()
	}
	wg.Wait()
	end := time.Now()
	fmt.Printf("write benchmark: %d op/s\n", int((m*n)/end.Sub(start).Seconds()))
}

func TestBadgerPerformance_CheckBlockMissing(t *testing.T) {
	db, err := NewBadger()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const m = 0
	const n = 10000

	if _, err := os.Stat(dir_checkblock); err != nil {
		if err := os.Mkdir(dir_checkblock, 0700); err != nil {
			t.Fatal(err)
		}
	}

	for k := 0; k < m; k++ {
		var buffer bytes.Buffer

		db.Update(func(txn *badger.Txn) error {
			for i := 0; i < n; i++ {
				blk := generateBlock()
				key := blk.GetHash()
				val, _ := blk.MarshalMsg(nil)
				if err := txn.Set(key[:], val); err != nil {
					t.Fatal(err)
				}
				buffer.WriteString(key.String())
				buffer.WriteString("\n")
			}
			return nil
		})

		if ioutil.WriteFile(fmt.Sprintf("%s/hash%s.txt", dir_checkblock, strconv.Itoa(k)), buffer.Bytes(), 0644) != nil {
			t.Fatal(err)
		}
	}

}

func TestBadgerPerformance_ReadBlock(t *testing.T) {
	db, err := NewBadger()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	dir, err := ioutil.ReadDir(dir_checkblock)
	if err != nil {
		t.Fatal(err)
	}
	index := 0
	var span time.Time

	for _, fi := range dir {
		if fileObj, err := os.Open(dir_checkblock + "/" + fi.Name()); err == nil {
			defer fileObj.Close()
			if contents, err := ioutil.ReadAll(fileObj); err == nil {
				hs := strings.Split(strings.TrimSpace(string(contents)), "\n")
				start := time.Now()

				db.View(func(txn *badger.Txn) error {
					for _, h := range hs {
						hash := types.Hash{}
						hash.Of(h)
						if item, err := txn.Get(hash[:]); err != nil {
							t.Fatal(err)
						} else {
							err := item.Value(func(val []byte) error {
								b := new(types.StateBlock)
								if _, err = b.UnmarshalMsg(val); err != nil {
									t.Fatal(err)
								}
								index++
								return nil
							})
							if err != nil {
								return nil
							}
						}
					}
					return nil
				})
				end := time.Now()
				d := end.Sub(start)
				span = span.Add(d)

				//fmt.Println("block num, ", index)
			} else {
				t.Fatal(err)
			}
		}
	}

	fmt.Println("total, ", index)
	fmt.Printf("read benchmark: %d s\n", span.Second())

}

func TestBadgerPerformance_ReadBlockByGoroutine(t *testing.T) {
	db, err := NewBadger()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	dir, err := ioutil.ReadDir(dir_checkblock)
	if err != nil {
		t.Log(err)
	}
	index := 0
	var wg sync.WaitGroup
	l := sync.Mutex{}
	start := time.Now()

	for _, fi := range dir {
		wg.Add(1)
		if fileObj, err := os.Open(dir_checkblock + "/" + fi.Name()); err == nil {
			defer fileObj.Close()
			if contents, err := ioutil.ReadAll(fileObj); err == nil {
				hs := strings.Split(strings.TrimSpace(string(contents)), "\n")
				go func(hs []string) {
					db.View(func(txn *badger.Txn) error {
						for _, h := range hs {
							hash := types.Hash{}
							hash.Of(h)
							if _, err := txn.Get(hash[:]); err != nil {
								t.Fatal(err)
							} else {
								l.Lock()
								index++
								l.Unlock()
							}
						}
						return nil
					})
					wg.Done()
				}(hs)

			} else {
				t.Fatal(err)
			}
		}
	}
	wg.Wait()
	fmt.Println(index)
	end := time.Now()
	fmt.Printf("read benchmark: %f s\n", end.Sub(start).Seconds())

}

func TestLedger_RemoveDB(t *testing.T) {
	if err := os.RemoveAll(dir_checkblock); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(dir_testdb); err != nil {
		t.Fatal(err)
	}
}
