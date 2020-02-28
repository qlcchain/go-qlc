package ledger

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *Ledger) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cfg, _ := cm.Load()
	l := NewLedger(cm.ConfigFile)
	for _, v := range cfg.Genesis.GenesisBlocks {
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: v.Mintage,
			GenesisBlock:        v.Genesis,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
	}
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

//var bc, _ = mock.BlockChain()

func TestLedger_Instance1(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	cm.Load()
	l1 := NewLedger(cm.ConfigFile)
	l2 := NewLedger(cm.ConfigFile)
	t.Logf("l1:%v,l2:%v", l1, l2)
	defer func() {
		l1.Close()
		//l2.Close()
		_ = os.RemoveAll(dir)
	}()
	b := reflect.DeepEqual(l1, l2)
	if l1 == nil || l2 == nil || !b {
		t.Fatal("error")
	}
	//_ = os.RemoveAll(dir)
}

func TestLedger_Instance2(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	dir2 := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	cm.Load()
	cm2 := config.NewCfgManager(dir2)
	cm2.Load()
	l1 := NewLedger(cm.ConfigFile)
	l2 := NewLedger(cm2.ConfigFile)
	defer func() {
		l1.Close()
		l2.Close()
		_ = os.RemoveAll(dir)
		_ = os.RemoveAll(dir2)
	}()
	if l1 == nil || l2 == nil || reflect.DeepEqual(l1, l2) {
		t.Fatal("error")
	}
}

func TestGetTxn(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	txn := l.store.Batch(true)
	fmt.Println(txn)
	txn2, flag := l.getBatch(true, txn)
	if flag {
		t.Fatal("get txn flag error")
	}
	if txn != txn2 {
		t.Fatal("txn!=tnx2")
	}
}

func TestLedgerSession_BatchUpdate(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	genesis := common.GenesisBlock()
	if err := l.AddStateBlock(&genesis); err != nil {
		t.Fatal()
	}
	blk := mock.StateBlockWithoutWork()
	blk.Link = genesis.GetHash()
	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal()
	}
	blk2 := mock.StateBlockWithoutWork()
	blk2.Link = genesis.GetHash()
	if err := l.AddStateBlock(blk2); err != nil {
		t.Fatal()
	}
	if ok, _ := l.HasStateBlock(blk.GetHash()); !ok {
		t.Fatal()
	}
}

func TestReleaseLedger(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger1")
	dir2 := filepath.Join(config.QlcTestDataDir(), "ledger2")

	cm := config.NewCfgManager(dir)
	cm.Load()
	cm2 := config.NewCfgManager(dir2)
	cm2.Load()
	l1 := NewLedger(cm.ConfigFile)
	_ = NewLedger(cm2.ConfigFile)
	defer func() {
		//only release ledger1
		l1.Close()
		CloseLedger()
		_ = os.RemoveAll(dir)
		_ = os.RemoveAll(dir2)
	}()
}
