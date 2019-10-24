package ledger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func addStateBlock(t *testing.T, l *Ledger) *types.StateBlock {
	blk := mock.StateBlockWithoutWork()
	com := common.GenesisBlock()
	blk.Link = com.GetHash()
	if err := l.AddStateBlock(&com); err != nil {
		t.Fatal(err)
	}
	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal(err)
	}
	return blk
}

func addSmartContractBlock(t *testing.T, l *Ledger) *types.SmartContractBlock {
	jsonBlock := `{
    "internalAccount": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "contract": {
      "abi": "mcvnzY+zF5mVDjsvknvPfFgRToMQAVI4wivQGRZBwerbUIvfrKD6/suZJWiFVOI5sbTa98bpY9+1cUhE2T9yidxSCpvZ4kkBVBMfcL3OJIqG",
      "abiLength": 81,
      "abiHash": "79dab43dcc97205918b297c3aba6259e3ab1ed7d0779dc78eec6f57e5d6307ce"
    },
    "owner": "qlc_1nawsw4yatupd47p3scd5x5i3s9szbsggxbxmfy56f8jroyu945i5seu1cdd",
	"isUseStorage": false,
    "type": "SmartContract",
    "address": "qlc_3watpnwym9i43kbkt35yfp8xnqo7c9ujp3b6udajza71mspjfzpnpdgoydzn",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000007bb1fe",
    "signature": "d9d71c82eccdca0324e102c089b28c1430b0ae61f2af809e6134b289d5186b16cbcb6fcd4bfc1424fd34aa40e9bdd05069bc56d05fecf833470d80d047048a05"
  }`
	blk := new(types.SmartContractBlock)
	_ = json.Unmarshal([]byte(jsonBlock), blk)
	if err := l.AddSmartContractBlock(blk); err != nil {
		t.Log(err)
	}
	return blk
}

func TestLedger_AddBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addStateBlock(t, l)
}

func TestLedger_GetBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addStateBlock(t, l)
	blk, err := l.GetStateBlock(block.GetHash())
	t.Log("blk,", blk)
	if err != nil || blk == nil {
		t.Fatal(err)
	}
}

func TestLedger_GetSmartContrantBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addSmartContractBlock(t, l)
	blk, err := l.GetSmartContractBlock(block.GetHash())
	t.Log("blk,", blk)
	if err != nil || blk == nil {
		t.Fatal(err)
	}
}

func TestLedger_HasSmartContrantBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addSmartContractBlock(t, l)
	b, err := l.HasSmartContractBlock(block.GetHash())
	t.Log(b)
	if err != nil || !b {
		t.Fatal(err)
	}
}

func TestLedger_GetSmartContractBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addSmartContractBlock(t, l)
	blk, err := l.GetSmartContractBlock(block.GetHash())
	t.Log("blk,", blk)
	if err != nil || blk == nil {
		t.Fatal(err)
	}
	n, err := l.CountSmartContractBlocks()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(n)
	err = l.GetSmartContractBlocks(func(block *types.SmartContractBlock) error {
		fmt.Println(block)
		return nil
	})
	fmt.Println(err)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_GetAllBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	genesis := common.GenesisBlock()
	if err := l.AddStateBlock(&genesis); err != nil {
		t.Fatal(err)
	}
	blk := mock.StateBlockWithoutWork()
	blk.Link = genesis.GetHash()
	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal(err)
	}
	blk2 := mock.StateBlockWithoutWork()
	blk2.Link = genesis.GetHash()
	if err := l.AddStateBlock(blk2); err != nil {
		t.Fatal(err)
	}
	err := l.GetStateBlocks(func(block *types.StateBlock) error {
		t.Log(block)
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_DeleteBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	block := addStateBlock(t, l)
	if err := l.DeleteStateBlock(block.GetHash()); err != nil {
		t.Fatal(err)
	}
}

func TestLedger_HasBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addStateBlock(t, l)
	r, err := l.HasStateBlock(block.GetHash())
	if err != nil || !r {
		t.Fatal(err)
	}
	t.Log("hasblock,", r)

	r, err = l.HasStateBlockConfirmed(block.GetHash())
	if err != nil || !r {
		t.Fatal(err)
	}
}

func TestLedger_GetRandomBlock_Empty(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	b, err := l.GetRandomStateBlock()

	if err != ErrStoreEmpty {
		t.Fatal(err)
	}
	t.Log("block ,", b)
}

func TestLedger_GetRandomBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk := mock.StateBlockWithoutWork()
	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal(err)
	}
	b, err := l.GetRandomStateBlock()

	if err != nil {
		t.Fatal(err)
	}

	if blk.GetHash() != b.GetHash() {
		t.Fatal("block not equal")
	}
	t.Log("block ,", b)
}

func addBlockCache(t *testing.T, l *Ledger) *types.StateBlock {
	blk := mock.StateBlockWithoutWork()
	if err := l.AddBlockCache(blk); err != nil {
		t.Fatal(err)
	}
	return blk
}

func TestLedger_AddBlockCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addBlockCache(t, l)
}

func TestLedger_HasBlockCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk := addBlockCache(t, l)
	b, err := l.HasBlockCache(blk.GetHash())
	if err != nil || !b {
		t.Fatal(err)
	}
}

func TestLedger_DeleteBlockCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	blk := addBlockCache(t, l)
	if err := l.DeleteBlockCache(blk.GetHash()); err != nil {
		t.Fatal(err)
	}
}

func TestLedger_CountBlockCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	blk := addBlockCache(t, l)
	if c, err := l.CountBlockCache(); err != nil || c != 1 {
		t.Fatal("CountBlockCache error,should be 1")
	}
	if err := l.DeleteBlockCache(blk.GetHash()); err != nil {
		t.Fatal(err)
	}
	if c, err := l.CountBlockCache(); err != nil || c != 0 {
		t.Fatal("CountBlockCache error,should be 0")
	}
}

func TestLedger_GetBlockCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	blk := addBlockCache(t, l)
	hash := blk.GetHash()
	blk1, err := l.GetBlockCache(hash)
	if err != nil {
		t.Fatal(err)
	}
	hash1 := blk1.GetHash()
	if hash != hash1 {
		t.Fatal("hash not match")
	}

	err = l.GetBlockCaches(func(block *types.StateBlock) error {
		fmt.Println(block)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_BlockChild(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addr1 := mock.Address()
	b1 := mock.StateBlockWithoutWork()
	b1.Address = addr1

	b2 := mock.StateBlockWithoutWork()
	b2.Type = types.Send
	b2.Previous = b1.GetHash()

	b3 := mock.StateBlockWithoutWork()
	b3.Type = types.Send
	b3.Previous = b1.GetHash()

	if err := l.AddStateBlock(b1); err != nil {
		t.Fatal(err)
	}
	if err := l.AddStateBlock(b2); err != nil {
		t.Fatal(err)
	}
	h, err := l.GetChild(b1.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	if h != b2.GetHash() {
		t.Fatal()
	}

	err = l.AddStateBlock(b3)
	if err != nil {
		t.Log(err)
	} else {
		t.Fatal()
	}

	if err := l.DeleteStateBlock(b2.GetHash()); err != nil {
		t.Fatal(err)
	}

	h, err = l.GetChild(b1.GetHash())
	if err != nil {
		t.Log(err)
	} else {
		t.Fatal()
	}

	if err := l.AddStateBlock(b3); err != nil {
		t.Fatal(err)
	}

	h, err = l.GetChild(b1.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	if h != b3.GetHash() {
		t.Fatal()
	}
}

func TestLedger_MessageInfo(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	h := mock.Hash()
	m := []byte{1, 2, 3}
	if err := l.AddMessageInfo(h, m); err != nil {
		t.Fatal(err)
	}
	m2, err := l.GetMessageInfo(h)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(m, m2) {
		t.Fatal("wrong result")
	}
}

func TestLedger_SyncBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	block := mock.StateBlockWithoutWork()
	if err := l.AddSyncCacheBlock(block); err != nil {
		t.Fatal(err)
	}
	count := 0
	err := l.GetSyncCacheBlocks(func(block *types.StateBlock) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("sync count error")
	}
	if err := l.DeleteSyncCacheBlock(block.GetHash()); err != nil {
		t.Fatal(err)
	}
	count2 := 0
	err = l.GetSyncCacheBlocks(func(block *types.StateBlock) error {
		count2++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count2 != 0 {
		t.Fatal("sync count error")
	}
}
