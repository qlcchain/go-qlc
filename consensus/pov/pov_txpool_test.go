package pov

import (
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/trie"
	"os"
	"path/filepath"
	"testing"
)

type povTxPoolMockData struct {
	eb     event.EventBus
	ledger ledger.Store
	chain  PovTxChainReader
}

type povTxChainReaderMockChain struct {
}

func (mc *povTxChainReaderMockChain) GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState {
	return nil
}

func (mc *povTxChainReaderMockChain) RegisterListener(listener EventListener) {}

func (mc *povTxChainReaderMockChain) UnRegisterListener(listener EventListener) {}

func setupPovTxPoolTestCase(t *testing.T) (func(t *testing.T), *povTxPoolMockData) {
	t.Parallel()

	md := &povTxPoolMockData{}

	uid := uuid.New().String()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uid)
	_ = os.RemoveAll(dir)
	md.ledger = ledger.NewLedger(dir)

	md.eb = event.GetEventBus(uid)

	md.chain = new(povTxChainReaderMockChain)

	return func(t *testing.T) {
		err := md.ledger.DBStore().Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}

		err = md.eb.Close()
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestPovTxPool_AddDelTx(t *testing.T) {
	teardownTestCase, md := setupPovTxPoolTestCase(t)
	defer teardownTestCase(t)

	txPool := NewPovTxPool(md.eb, md.ledger, md.chain)
	if txPool == nil {
		t.Fatal("NewPovTxPool is nil")
	}

	txPool.Init()

	txPool.Start()

	txBlk1 := mock.StateBlockWithoutWork()
	txHash1 := txBlk1.GetHash()
	txPool.addTx(txHash1, txBlk1)

	retTxBlk1 := txPool.getTx(txHash1)
	if retTxBlk1 == nil {
		t.Fatalf("failed to add tx %s", txHash1)
	}

	txPool.delTx(txHash1)

	retTxBlk1 = txPool.getTx(txHash1)
	if retTxBlk1 != nil {
		t.Fatalf("failed to delete tx %s", txHash1)
	}

	txPool.Stop()
}

func TestPovTxPool_SelectTx(t *testing.T) {
	teardownTestCase, md := setupPovTxPoolTestCase(t)
	defer teardownTestCase(t)

	txPool := NewPovTxPool(md.eb, md.ledger, md.chain)
	if txPool == nil {
		t.Fatal("NewPovTxPool is nil")
	}

	txPool.Init()

	txPool.Start()

	txBlk1 := mock.StateBlockWithoutWork()
	txHash1 := txBlk1.GetHash()
	txPool.addTx(txHash1, txBlk1)

	txBlk2 := mock.StateBlockWithoutWork()
	txHash2 := txBlk2.GetHash()
	txPool.addTx(txHash2, txBlk2)

	statTrie := trie.NewTrie(md.ledger.DBStore(), nil, trie.NewSimpleTrieNodePool())
	retTxs := txPool.SelectPendingTxs(statTrie, 10)
	//t.Logf("retTxs %d", len(retTxs))

	tx1Exist := false
	for _, retTx := range retTxs {
		if retTx.GetHash() == txHash1 {
			tx1Exist = true
			break
		}
	}
	if !tx1Exist {
		t.Fatalf("failed to select tx1 %s", txHash1)
	}

	tx2Exist := false
	for _, retTx := range retTxs {
		if retTx.GetHash() == txHash2 {
			tx2Exist = true
			break
		}
	}
	if !tx2Exist {
		t.Fatalf("failed to select tx2 %s", txHash2)
	}

	txPool.Stop()
}
