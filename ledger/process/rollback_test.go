package process

import (
	"fmt"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/trie"
)

func TestRollback_Block(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)
	if err := lv.BlockProcess(bc[0]); err != nil {
		t.Fatal(err)
	}
	t.Log("bc hash", bc[0].GetHash())
	for i, b := range bc[1:] {
		fmt.Println(i + 1)
		fmt.Println("bc.previous", b.GetPrevious())
		if p, err := lv.Process(b); err != nil || p != Progress {
			t.Fatal(p, err)
		}
	}

	rb := bc[2]
	fmt.Println("rollback")
	fmt.Println("rollback hash: ", rb.GetHash(), rb.GetType(), rb.GetPrevious().String())
	if err := lv.Rollback(rb.GetHash()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	checkInfo(t, l)
}

func checkInfo(t *testing.T, l *ledger.Ledger) {
	addrs := make(map[types.Address]int)
	fmt.Println("----blocks----")
	err := l.GetStateBlocksConfirmed(func(block *types.StateBlock) error {
		fmt.Println(block)
		if block.GetHash() != config.GenesisBlockHash() {
			if _, ok := addrs[block.GetAddress()]; !ok {
				addrs[block.GetAddress()] = 0
			}
			return nil
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(addrs)
	fmt.Println("----frontiers----")
	fs, _ := l.GetFrontiers()
	for _, f := range fs {
		fmt.Println(f)
	}

	fmt.Println("----account----")
	for k := range addrs {
		ac, err := l.GetAccountMeta(k)
		if err != nil {
			t.Fatal(err, k)
		}
		fmt.Println("   account", ac.Address)
		for _, token := range ac.Tokens {
			fmt.Println("      token, ", token)
		}
	}

	fmt.Println("----representation----")
	for k := range addrs {
		b, err := l.GetRepresentation(k)
		if err != nil {
			if err == ledger.ErrRepresentationNotFound {
			}
		} else {
			fmt.Println(k, b)
		}
	}

	fmt.Println("----pending----")
	err = l.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		fmt.Println("      key:", pendingKey)
		fmt.Println("      info:", pendingInfo)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

//func TestRollback_BlockCache(t *testing.T) {
//	teardownTestCase, _, lv := setupTestCase(t)
//	defer teardownTestCase(t)
//	// account, only one token
//	addr := mock.Address()
//	ac := mock.AccountMeta(addr)
//	tm := mock.TokenMeta(addr)
//	tm.Type = config.ChainToken()
//	ac.Tokens = []*types.TokenMeta{tm}
//
//	// set block1
//	block1 := mock.StateBlockWithoutWork()
//	block1.Address = ac.Address
//	block1.Token = ac.Tokens[0].Type
//	if err := lv.l.AddBlockCache(block1); err != nil {
//		t.Fatal(err)
//	}
//	tm.BlockCount = 2
//	if err := lv.l.AddAccountMetaCache(ac); err != nil {
//		t.Fatal(err)
//	}
//
//	// set block2
//	block2 := mock.StateBlockWithoutWork()
//	block2.Previous = block1.GetHash()
//	block2.Address = ac.Address
//	block2.Token = ac.Tokens[0].Type
//
//	if err := lv.l.AddBlockCache(block2); err != nil {
//		t.Fatal(err)
//	}
//
//	if err := lv.Rollback(block1.GetHash()); err != nil {
//		t.Fatal(err)
//	}
//
//	if b, _ := lv.l.HasBlockCache(block1.GetHash()); b {
//		t.Fatal()
//	}
//	ac, err := lv.l.GetAccountMeteCache(addr)
//	if err == nil {
//		t.Fatal(err)
//	}
//}

func TestRollback_BlockCache2(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	if err := lv.BlockProcess(bc[0]); err != nil {
		t.Fatal(Other, err)
	}
	if err := lv.BlockCacheProcess(bc[1]); err != nil {
		t.Fatal(Other, err)
	}
	if err := lv.BlockProcess(bc[1]); err != nil {
		t.Fatal(Other, err)
	}
	for _, b := range bc[2:5] {
		if r, err := lv.BlockCacheCheck(b); r != Progress || err != nil {
			t.Fatal(r, err)
		}
		if err := lv.BlockCacheProcess(b); err != nil {
			t.Fatal(Other, err)
		}
	}

	// rollbackCache - rollbackCacheBlocks(cache para is true)
	if err := lv.Rollback(bc[3].GetHash()); err != nil {
		t.Fatal(err)
	}
}

func TestRollback_checkBlockCache(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)

	if err := lv.BlockProcess(bc[0]); err != nil {
		t.Fatal(Other, err)
	}
	if err := lv.BlockCacheProcess(bc[1]); err != nil {
		t.Fatal(Other, err)
	}
	if err := lv.BlockProcess(bc[1]); err != nil {
		t.Fatal(Other, err)
	}
	for _, b := range bc[2:5] {
		if r, err := lv.BlockCacheCheck(b); r != Progress || err != nil {
			t.Fatal(r, err)
		}
		if err := lv.BlockCacheProcess(b); err != nil {
			t.Fatal(Other, err)
		}
	}

	batch := l.DBStore().Batch(true)
	defer func() {
		if err := l.DBStore().PutBatch(batch); err != nil {
			t.Fatal(err)
		}
	}()

	//  checkBlockCache - rollbackCacheBlocks(cache para is false)
	if err := lv.checkBlockCache(bc[3], batch); err != nil {
		t.Fatal(err)
	}

	if err := lv.checkBlockCache(mock.StateBlockWithoutWork(), batch); err != nil {
		t.Fatal(err)
	}

	sendBlk := mock.StateBlockWithoutWork()
	sendBlk.Type = types.Send
	receBlk := mock.StateBlockWithoutWork()
	receBlk.Type = types.Receive
	receBlk.Link = sendBlk.GetHash()
	if err := lv.l.AddBlockCache(receBlk); err != nil {
		t.Fatal(err)
	}
	if err := lv.checkBlockCache(sendBlk, batch); err != nil {
		t.Fatal(err)
	}
}

func TestRollback_UncheckedBlock(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)

	h1 := mock.Hash()
	preBlock := mock.StateBlockWithoutWork()
	preBlock.Previous = h1
	linkBlock := mock.StateBlockWithoutWork()
	linkBlock.Link = h1
	if err := l.AddUncheckedBlock(h1, preBlock, types.UncheckedKindPrevious, types.Synchronized); err != nil {
		t.Fatal(err)
	}
	if err := l.AddUncheckedBlock(h1, linkBlock, types.UncheckedKindLink, types.Synchronized); err != nil {
		t.Fatal(err)
	}
	lv.RollbackUnchecked(h1)

	// UncheckedKindTokenInfo

}

func TestRollback_ContractData(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)

	bs := mock.ContractBlocks()
	if err := lv.BlockProcess(bs[0]); err != nil {
		t.Fatal(err)
	}
	for i, b := range bs[1:] {
		fmt.Println("index，", i+1)
		fmt.Println("bc.previous", b.GetPrevious())
		if p, err := lv.Process(b); err != nil || p != Progress {
			t.Fatal(p, err)
		}
	}
	time.Sleep(2 * time.Second)

	nodeCount := nodesCount(lv.l.DBStore(), bs[2].GetExtra())
	if nodeCount == 0 {
		t.Fatal("failed to add nodes", nodeCount)
	}

	time.Sleep(1 * time.Second)
	c, err := l.CountStateBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if int(c) != len(bs) {
		t.Fatal("block count error")
	}
	fmt.Println("block count === ", c)

	nodeCount = nodesCount(lv.l.DBStore(), bs[2].GetExtra())
	if nodeCount == 0 {
		t.Fatal("failed to add nodes", nodeCount)
	}
	fmt.Println(" node count ", nodeCount)

	if err := lv.Rollback(bs[2].GetHash()); err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)
	nodeCount = nodesCount(lv.l.DBStore(), bs[2].GetExtra())
	t.Log(nodeCount)
	if nodeCount > 0 {
		t.Fatal("failed to remove nodes", nodeCount)
	}

	fmt.Println("rollback again")
	if p, err := lv.Process(bs[2]); err != nil || p != Progress {
		t.Fatal(p, err)
	}
	time.Sleep(2 * time.Second)
	nodeCount = nodesCount(lv.l.DBStore(), bs[2].GetExtra())
	t.Log(nodeCount)
	if nodeCount == 0 {
		t.Fatal("failed to add nodes", nodeCount)
	}

	if err := lv.Rollback(bs[2].GetHash()); err != nil {
		t.Fatal(err)
	}
	fmt.Println("process again")

	time.Sleep(2 * time.Second)
	if p, err := lv.Process(bs[2]); err != nil || p != Progress {
		t.Fatal(p, err)
	}
	//time.Sleep(4 * time.Second)
	//nodeCount = nodesCount(lv.l.DBStore(), bs[2].GetExtra())
	//t.Log(nodeCount)
	//if nodeCount == 0 {
	//	t.Fatal("failed to add nodes", nodeCount)
	//}
}

func nodesCount(db storage.Store, rootHash types.Hash) int {
	tr := trie.NewTrie(db, &rootHash, trie.NewSimpleTrieNodePool())
	iterator := tr.NewIterator(nil)
	counter := 0
	for {
		if _, _, ok := iterator.Next(); !ok {
			break
		} else {
			counter++
		}
	}
	return counter
}

func TestRollback_ContractBlocks(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	bs := mock.ContractBlocks()
	if err := lv.BlockProcess(bs[0]); err != nil {
		t.Fatal(err)
	}
	for i, b := range bs[1:] {
		fmt.Println(i)
		if r, err := lv.BlockCheck(b); r != Progress || err != nil {
			t.Fatal(r, err)
		}
		if err := lv.BlockProcess(b); err != nil {
			t.Fatal(Other, err)
		}
	}
	if err := lv.Rollback(bs[1].GetHash()); err != nil {
		t.Fatal(err)
	}
}

func TestRollback_blockOrderCompare(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	if err := lv.BlockProcess(bc[0]); err != nil {
		t.Fatal(err)
	}
	for _, b := range bc[1:] {
		if p, err := lv.Process(b); err != nil || p != Progress {
			t.Fatal(p, err)
		}
	}

	if b, err := lv.blockOrderCompare(bc[2], bc[4]); err != nil || !b {
		t.Fatal(b, err)
	}
	if b, err := lv.blockOrderCompare(bc[4], bc[2]); err != nil || b {
		t.Fatal(b, err)
	}
	if b, err := lv.blockOrderCompare(bc[1], bc[2]); err == nil {
		t.Fatal(b, err)
	}
}

//func TestRollback_checkBlockCache(t *testing.T) {
//	teardownTestCase, _, lv := setupTestCase(t)
//	defer teardownTestCase(t)
//
//}
