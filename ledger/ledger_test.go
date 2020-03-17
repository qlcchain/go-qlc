package ledger

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/crypto/random"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *Ledger) {
	//t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
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

//var bc, _ = mock.BlockChain()

func TestLedger_Instance1(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
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
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	dir2 := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
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

func TestLedger_BatchUpdate(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	genesis := config.GenesisBlock()
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

func TestLedger_Release(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	dir2 := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())

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

func TestLedger_Close(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())

	cm := config.NewCfgManager(dir)
	cm.Load()
	l1 := NewLedger(cm.ConfigFile)
	blk := mock.StateBlockWithoutWork()
	if len(lcache) != 1 {
		t.Fatal(len(lcache))
	}
	if err := l1.Close(); err != nil {
		t.Fatal(err)
	}
	if len(lcache) != 0 {
		t.Fatal(len(lcache))
	}

	l2 := NewLedger(cm.ConfigFile)
	if _, err := l2.GetStateBlockConfirmed(blk.GetHash()); err == nil {
		t.Fatal(err)
	}
	if len(lcache) != 1 {
		t.Fatal(len(lcache))
	}
	if err := l2.Close(); err != nil {
		t.Fatal(err)
	}
	if len(lcache) != 0 {
		t.Fatal(len(lcache))
	}
}

func TestLedger_GenerateSendBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	// init ac1
	ac1 := mock.Account()
	balance := types.Balance{Int: big.NewInt(int64(100000000000))}
	blk := new(types.StateBlock)
	blk.Type = types.Open
	blk.Address = ac1.Address()
	blk.Previous = types.ZeroHash
	blk.Token = config.ChainToken()
	blk.Balance = balance
	blk.Timestamp = common.TimeNow().Unix()
	blk.Link = mock.Hash()
	blk.Representative = ac1.Address()

	am := mock.AccountMeta(ac1.Address())
	tm := &types.TokenMeta{
		Type:           config.ChainToken(),
		Header:         blk.GetHash(),
		OpenBlock:      types.ZeroHash,
		Representative: ac1.Address(),
		Balance:        balance,
		BelongTo:       ac1.Address(),
	}
	am.Tokens = append(am.Tokens, tm)

	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal(err)
	}
	if err := l.AddAccountMeta(am, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	block, td := generatePovBlock(nil)
	err := l.AddPovBlock(block, td)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovBestHash(block.GetHeight(), block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.SetPovLatestHeight(block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	// GenerateSendBlock
	ac2 := mock.Account()
	ac2Addr := ac2.Address()
	balance2 := types.Balance{Int: big.NewInt(int64(100000))}
	sendBlk1, err := l.GenerateSendBlock(&types.StateBlock{
		Address: ac1.Address(),
		Token:   config.ChainToken(),
		Link:    ac2Addr.ToHash(),
	}, balance2, ac1.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(sendBlk1)

	if err := l.AddStateBlock(sendBlk1); err != nil {
		t.Fatal(err)
	}

	tm.Header = sendBlk1.GetHash()
	tm.Balance = tm.Balance.Sub(balance2)
	if err := l.UpdateAccountMeta(am, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	pendingKey := &types.PendingKey{
		Address: ac2.Address(),
		Hash:    sendBlk1.GetHash(),
	}
	pendingInfo := &types.PendingInfo{
		Source: ac1.Address(),
		Type:   config.ChainToken(),
		Amount: balance2,
	}
	if err := l.AddPending(pendingKey, pendingInfo, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	// GenerateReceiveBlock if ac2 has no tokenmeta
	receBlk1, err := l.GenerateReceiveBlock(sendBlk1, ac2.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(receBlk1)

	// GenerateReceiveBlock if ac2 has tokenmeta
	if err := l.AddStateBlock(receBlk1); err != nil {
		t.Fatal(err)
	}
	am2 := mock.AccountMeta(ac2.Address())
	tm2 := &types.TokenMeta{
		Type:           config.ChainToken(),
		Header:         receBlk1.GetHash(),
		OpenBlock:      receBlk1.GetHash(),
		Representative: ac1.Address(),
		Balance:        balance2,
		BelongTo:       ac2.Address(),
	}
	am2.Tokens = append(am2.Tokens, tm2)
	if err := l.AddAccountMeta(am2, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	sendBlk2, err := l.GenerateSendBlock(&types.StateBlock{
		Address: ac1.Address(),
		Token:   config.ChainToken(),
		Link:    ac2Addr.ToHash(),
	}, balance2, ac1.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(sendBlk2)

	if err := l.AddStateBlock(sendBlk2); err != nil {
		t.Fatal(err)
	}

	pendingKey2 := &types.PendingKey{
		Address: ac2.Address(),
		Hash:    sendBlk2.GetHash(),
	}
	pendingInfo2 := &types.PendingInfo{
		Source: ac1.Address(),
		Type:   config.ChainToken(),
		Amount: balance2,
	}
	if err := l.AddPending(pendingKey2, pendingInfo2, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	receBlk2, err := l.GenerateReceiveBlock(sendBlk2, ac2.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(receBlk2)

	// GenerateChangeBlock
	ac3 := mock.AccountMeta(mock.Address())
	if err := l.AddAccountMeta(ac3, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	changeBlk, err := l.GenerateChangeBlock(ac1.Address(), ac3.Address, ac1.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(changeBlk)

	// GenerateOnlineBlock
	onlineBlk, err := l.GenerateOnlineBlock(ac1.Address(), ac1.PrivateKey(), 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(onlineBlk)

	// CalculateAmount
	b1, err := l.CalculateAmount(sendBlk1)
	if err != nil {
		t.Fatal(err)
	}
	if !b1.Equal(balance2) {
		t.Fatal()
	}
	b2, err := l.CalculateAmount(receBlk2)
	if err != nil {
		t.Fatal(err)
	}
	if !b2.Equal(balance2) {
		t.Fatal()
	}
	b3, err := l.CalculateAmount(changeBlk)
	if err != nil {
		t.Fatal(err)
	}
	if !b3.Equal(types.ZeroBalance) {
		t.Fatal()
	}
}

func TestLedger_TestStore(t *testing.T) {
	teardownTestCase, l := NewTestLedger()
	defer teardownTestCase()
	if l == nil {
		t.Fatal()
	}
	r := DefaultStore()
	if r != nil {
		t.Fatal()
	}
}

func TestLedger_Action(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	r, err := l.Action(storage.Size, 0)
	if err != nil {
		t.Fatal()
	}
	t.Log(r)
}

func TestLedger_Cache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	// get from a specific
	c := l.cache.GetCache()
	key := []byte{1, 2, 3}
	if err := c.Put(key, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if i, r, err := l.Get(key, c); err != nil || i == nil || r != nil {
		t.Fatal(err, i, r)
	}
	if err := c.Delete(key); err != nil {
		t.Fatal(err)
	}
	if i, r, err := l.Get(key, c); err != storage.KeyNotFound || i != nil || r != nil {
		t.Fatal(err, i, r)
	}
	if i, r, err := l.Get(key); err != storage.KeyNotFound || i != nil || r != nil {
		t.Fatal(err, i, r)
	}
	// get from ledger cache
	key2 := []byte{1, 2, 4}
	if err := l.store.Put(key2, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if i, r, err := l.Get(key2); err != nil || i != nil || r == nil {
		t.Fatal(err, i, r)
	}

	// ledger put
	key3 := []byte{1, 2, 5}
	if err := l.Put(key3, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if err := l.Delete(key3); err != nil {
		t.Fatal(err)
	}
	if i, r, err := l.Get(key3); err != storage.KeyNotFound || i != nil || r != nil {
		t.Fatal(err, i, r)
	}

	// ledger get from store
	key4 := []byte{1, 2, 6}
	b := l.store.Batch(true)
	if err := b.Put(key4, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if r, err := l.getFromStore(key4, b); err != nil || r == nil {
		t.Fatal(err)
	}
	if err := l.store.PutBatch(b); err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)
	t.Log(l.GetCacheStatue())
	t.Log(l.GetCacheStat())
}

func TestLedger_Iterator(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	prefix := []byte{10, 20, 30, 40}
	for i := 0; i < 100; i++ {
		cache := l.Cache().GetCache()
		d1 := make([]byte, 10)
		_ = random.Bytes(d1)
		if err := cache.Put(append(prefix, d1...), d1); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(1 * time.Second)
	for i := 0; i < 100; i++ {
		cache := l.Cache().GetCache()
		d1 := make([]byte, 12)
		_ = random.Bytes(d1)
		if err := cache.Put(append(prefix, d1...), d1); err != nil {
			t.Fatal(err)
		}
	}
	count := 0
	err := l.Iterator(prefix, nil, func(k []byte, v []byte) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(count)
	if count != 200 {
		t.Fatal()
	}
}
