package relation

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	chaincontext "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/relation/db"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/mock"
)

func TestRelation_Relation(t *testing.T) {
	teardownTestCase, r := setupTestCase(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()
	v1, _ := blk1.ConvertToSchema()
	r.Add(v1)
	v2, _ := blk2.ConvertToSchema()
	r.Add(v2)
	for i := 0; i < batchMaxCount+10; i++ {
		v, _ := mock.StateBlockWithoutWork().ConvertToSchema()
		r.Add(v)
	}
	r.Delete(&types.BlockHash{Hash: blk1.GetHash().String()})
	r.Delete(&types.BlockHash{Hash: blk2.GetHash().String()})
	time.Sleep(3 * time.Second)
	c, err := r.BlocksCount()
	if err != nil || c != batchMaxCount+10 {
		t.Fatal(err, c, batchMaxCount+10)
	}
	if err := r.EmptyStore(); err != nil {
		t.Fatal(err)
	}
	c, err = r.BlocksCount()
	if err != nil || c != 0 {
		t.Fatal(err)
	}
}

func TestRelation_flush(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "relation", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cfgFile := cm.ConfigFile

	cc := chaincontext.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	store, err := db.NewDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	r := &Relation{
		db:         store,
		eb:         cc.EventBus(),
		dir:        cfgFile,
		deleteChan: make(chan types.Schema, 10240),
		addChan:    make(chan types.Schema, 10240),
		closedChan: make(chan bool),
		tables:     make(map[string]schema),
		logger:     log.NewLogger("relation"),
	}
	tables := new(types.BlockHash)
	if err := r.Register(tables); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	for i := 0; i < batchMaxCount+10; i++ {
		v1, _ := mock.StateBlockWithoutWork().ConvertToSchema()
		r.Add(v1)
	}
	for i := 0; i < batchMaxCount+10; i++ {
		r.Delete(&types.BlockHash{Hash: mock.StateBlockWithoutWork().GetHash().String()})
	}
	r.flush()
	if len(r.addChan) > 0 || len(r.deleteChan) > 0 {
		t.Fatal(len(r.addChan), len(r.deleteChan))
	}
	r.DB()
}

func TestRelation_Close(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "relation", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()

	store, err := NewRelation(cm.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(cache) != 1 {
		t.Fatal(len(cache))
	}
	v1, _ := mock.StateBlockWithoutWork().ConvertToSchema()
	store.Add(v1)
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if len(cache) != 0 {
		t.Fatal(len(cache))
	}

	store2, err := NewRelation(cm.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(cache) != 1 {
		t.Fatal(len(cache))
	}
	if err := store2.Close(); err != nil {
		t.Fatal(err)
	}
	if len(cache) != 0 {
		t.Fatal(len(cache))
	}
}
