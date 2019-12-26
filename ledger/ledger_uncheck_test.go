package ledger

import (
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func addUncheckedBlock(t *testing.T, l *Ledger) (hash types.Hash, block *types.StateBlock, kind types.UncheckedKind) {
	block = mock.StateBlockWithoutWork()
	hash = block.GetLink()
	kind = types.UncheckedKindLink
	fmt.Println(hash)
	if err := l.AddUncheckedBlock(hash, block, kind, types.UnSynchronized); err != nil {
		t.Fatal(err)
	}
	return
}

func TestLedger_AddUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addUncheckedBlock(t, l)
}

func TestLedger_GetUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	parentHash, _, kind := addUncheckedBlock(t, l)

	if b, s, err := l.GetUncheckedBlock(parentHash, kind); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("unchecked,%s", b)
		t.Log(s)
	}
}

func TestLedger_CountUncheckedBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addUncheckedBlock(t, l)
	addUncheckedBlock(t, l)

	c, err := l.CountUncheckedBlocks()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("unchecked count,", c)
}

func TestLedger_HasUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	parentHash, _, kind := addUncheckedBlock(t, l)
	r, err := l.HasUncheckedBlock(parentHash, kind)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has unchecked,", r)
}

func TestLedger_GetUncheckedBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addUncheckedBlock(t, l)
	addUncheckedBlock(t, l)

	err := l.WalkUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
		t.Log(block)
		t.Log(link, unCheckType, sync)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_DeleteUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	parentHash, _, kind := addUncheckedBlock(t, l)
	err := l.DeleteUncheckedBlock(parentHash, kind)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddGapPovBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()

	err := l.AddGapPovBlock(10, blk1, types.Synchronized)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddGapPovBlock(10, blk2, types.UnSynchronized)
	if err != nil {
		t.Fatal(err)
	}

	blks, kinds, err := l.GetGapPovBlock(10)
	if err != nil {
		t.Fatal(err)
	}

	if len(blks) != 2 || len(kinds) != 2 {
		t.Fatal("len err", len(blks), len(kinds))
	}

	if blks[0].GetHash() != blk1.GetHash() || blks[1].GetHash() != blk2.GetHash() {
		t.Fatal("block err", blks[0].GetHash(), blks[1].GetHash())
	}

	if kinds[0] != types.Synchronized || kinds[1] != types.UnSynchronized {
		t.Fatal("kind err", kinds)
	}

	err = l.DeleteGapPovBlock(10)
	if err != nil {
		t.Fatal(err)
	}

	blks, kinds, err = l.GetGapPovBlock(10)
	if len(blks) != 0 || len(kinds) != 0 {
		t.Fatal(err)
	}
}

func TestLedger_WalkGapPovBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()

	err := l.AddGapPovBlock(10, blk1, types.Synchronized)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddGapPovBlock(100, blk2, types.UnSynchronized)
	if err != nil {
		t.Fatal(err)
	}

	err = l.WalkGapPovBlocks(func(blocks types.StateBlockList, height uint64, sync types.SynchronizedKind) error {
		switch height {
		case 10:
			if blocks[0].GetHash() != blk1.GetHash() {
				t.Fatal("block err")
			}

			if sync != types.Synchronized {
				t.Fatal("sync kind err")
			}
		case 100:
			if blocks[0].GetHash() != blk2.GetHash() {
				t.Fatal("block err")
			}

			if sync != types.UnSynchronized {
				t.Fatal("sync kind err")
			}
		default:
			t.Fatal("height err", height)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddGapPublishBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()
	hash := mock.Hash()

	err := l.AddGapPublishBlock(hash, blk1, types.Synchronized)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddGapPublishBlock(hash, blk2, types.Synchronized)
	if err != nil {
		t.Fatal(err)
	}

	b1Match := false
	b2Match := false
	err = l.WalkGapPublishBlock(hash, func(block *types.StateBlock, sync types.SynchronizedKind) error {
		bHash := block.GetHash()
		if bHash == blk1.GetHash() {
			b1Match = true
		}
		if bHash == blk2.GetHash() {
			b2Match = true
		}
		return nil
	})

	if !b1Match || !b2Match {
		t.Fatal()
	}
}

func TestLedger_DeleteGapPublishBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()
	hash := mock.Hash()

	err := l.AddGapPublishBlock(hash, blk1, types.Synchronized)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddGapPublishBlock(hash, blk2, types.Synchronized)
	if err != nil {
		t.Fatal(err)
	}

	b1Match := false
	b2Match := false
	err = l.WalkGapPublishBlock(hash, func(block *types.StateBlock, sync types.SynchronizedKind) error {
		bHash := block.GetHash()
		if bHash == blk1.GetHash() {
			b1Match = true
		}
		if bHash == blk2.GetHash() {
			b2Match = true
		}
		return nil
	})

	if !b1Match || !b2Match {
		t.Fatal()
	}

	err = l.WalkGapPublishBlock(hash, func(block *types.StateBlock, sync types.SynchronizedKind) error {
		err := l.DeleteGapPublishBlock(hash, block.GetHash())
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})

	b1Match = false
	b2Match = false
	err = l.WalkGapPublishBlock(hash, func(block *types.StateBlock, sync types.SynchronizedKind) error {
		bHash := block.GetHash()
		if bHash == blk1.GetHash() {
			b1Match = true
		}
		if bHash == blk2.GetHash() {
			b2Match = true
		}
		return nil
	})

	if b1Match || b2Match {
		t.Fatal()
	}
}
