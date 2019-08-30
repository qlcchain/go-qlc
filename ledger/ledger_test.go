package ledger

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *Ledger) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	l := NewLedger(dir)

	return func(t *testing.T) {
		//err := l.Store.Erase()
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
	l1 := NewLedger(dir)
	l2 := NewLedger(dir)
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
	l1 := NewLedger(dir)
	l2 := NewLedger(dir2)
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
	txn := l.Store.NewTransaction(false)
	fmt.Println(txn)
	txn2, flag := l.getTxn(false, txn)
	if flag {
		t.Fatal("get txn flag error")
	}
	if txn != txn2 {
		t.Fatal("txn!=tnx2")
	}
}

func TestLedger_Empty(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	e, err := l.Empty()

	if err != nil {
		t.Fatal(err)
	}
	t.Logf("is empty %s", strconv.FormatBool(e))
}

func TestLedgerSession_BatchUpdate(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	err := l.BatchUpdate(func(txn db.StoreTxn) error {
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
		if ok, err := l.HasStateBlock(blk.GetHash()); err != nil || !ok {
			t.Fatal()
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestReleaseLedger(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger1")
	dir2 := filepath.Join(config.QlcTestDataDir(), "ledger2")

	l1 := NewLedger(dir)
	_ = NewLedger(dir2)
	defer func() {
		//only release ledger1
		l1.Close()
		CloseLedger()
		_ = os.RemoveAll(dir)
		_ = os.RemoveAll(dir2)
	}()
}
