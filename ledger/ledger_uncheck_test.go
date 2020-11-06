package ledger

import (
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
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

	// deserialize error
	key := mock.Hash()
	k, err := storage.GetKeyOfParts(l.uncheckedKindToPrefix(types.UncheckedKindPrevious), key)
	if err != nil {
		t.Fatal()
	}
	d1 := make([]byte, 0)
	_ = random.Bytes(d1)
	if err := l.store.Put(k, d1); err != nil {
		t.Fatal(err)
	}
	if r, u, err := l.GetUncheckedBlock(key, types.UncheckedKindPrevious); err == nil {
		t.Fatal(err, r, u)
	}

	if err := l.GetUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
		return nil
	}); err == nil {
		t.Fatal(err)
	}

	blk := mock.StateBlockWithoutWork()
	key2 := mock.Hash()
	k2, err := storage.GetKeyOfParts(storage.KeyPrefixGapPublish, key2, blk.GetHash())
	if err != nil {
		t.Fatal()
	}
	d2 := make([]byte, 10)
	_ = random.Bytes(d2)
	if err := l.store.Put(k2, d2); err != nil {
		t.Fatal(err)
	}
	if err := l.GetGapPublishBlock(key2, func(block *types.StateBlock, sync types.SynchronizedKind) error {
		return nil
	}); err == nil {
		t.Fatal(err)
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
	r, _ := l.HasUncheckedBlock(parentHash, kind)
	if !r {
		t.Fatal()
	}
	t.Log("has unchecked,", r)
}

func TestLedger_GetUncheckedBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addUncheckedBlock(t, l)
	addUncheckedBlock(t, l)

	err := l.GetUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
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

	if err := l.FlushU(); err != nil {
		t.Fatal(err)
	}

	b1m := false
	b2m := false
	err = l.WalkGapPovBlocksWithHeight(10, func(block *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		if height != 10 {
			t.Fatal()
		}

		hash := block.GetHash()
		if hash == blk1.GetHash() && sync == types.Synchronized {
			b1m = true
		}
		if hash == blk2.GetHash() && sync == types.UnSynchronized {
			b2m = true
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if b1m == false || b2m == false {
		t.Fatal()
	}
}

func TestLedger_DeleteGapPovBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk := mock.StateBlockWithoutWork()

	err := l.AddGapPovBlock(10, blk, types.Synchronized)
	if err != nil {
		t.Fatal(err)
	}

	err = l.WalkGapPovBlocksWithHeight(10, func(block *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		if height != 10 {
			t.Fatal()
		}

		hash := block.GetHash()
		if hash != blk.GetHash() || sync != types.Synchronized {
			t.Fatal()
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	err = l.DeleteGapPovBlock(10, blk.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.WalkGapPovBlocksWithHeight(10, func(block *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		t.Fatal()
		return nil
	})

	if err != nil {
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
	if err := l.FlushU(); err != nil {
		t.Fatal(err)
	}

	b1m := false
	b2m := false
	err = l.WalkGapPovBlocks(func(block *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		switch height {
		case 10:
			if block.GetHash() == blk1.GetHash() && sync == types.Synchronized {
				b1m = true
			}
		case 100:
			if block.GetHash() == blk2.GetHash() && sync == types.UnSynchronized {
				b2m = true
			}
		default:
			t.Fatal()
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if b1m == false || b2m == false {
		t.Fatal()
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
	if err := l.FlushU(); err != nil {
		t.Fatal(err)
	}

	b1Match := false
	b2Match := false
	err = l.GetGapPublishBlock(hash, func(block *types.StateBlock, sync types.SynchronizedKind) error {
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
	if err := l.FlushU(); err != nil {
		t.Fatal(err)
	}

	b1Match := false
	b2Match := false
	err = l.GetGapPublishBlock(hash, func(block *types.StateBlock, sync types.SynchronizedKind) error {
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

	err = l.GetGapPublishBlock(hash, func(block *types.StateBlock, sync types.SynchronizedKind) error {
		err := l.DeleteGapPublishBlock(hash, block.GetHash())
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err := l.FlushU(); err != nil {
		t.Fatal(err)
	}

	b1Match = false
	b2Match = false
	err = l.GetGapPublishBlock(hash, func(block *types.StateBlock, sync types.SynchronizedKind) error {
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

func TestLedger_GetGapPublishBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk := mock.StateBlockWithoutWork()
	hash := mock.Hash()

	err := l.AddGapPublishBlock(hash, blk, types.Synchronized)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.FlushU(); err != nil {
		t.Fatal(err)
	}

	get := false
	err = l.GetGapPublishBlock(hash, func(block *types.StateBlock, sync types.SynchronizedKind) error {
		if blk.GetHash() == block.GetHash() {
			get = true
		}
		return nil
	})

	if !get {
		t.Fatal()
	}
}

func TestLedger_AddGapDoDSettleStateBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	internalId := mock.Hash()
	block := mock.StateBlockWithoutWork()
	err := l.AddGapDoDSettleStateBlock(internalId, block, types.UnSynchronized)
	if err != nil {
		t.Fatal(err)
	}

	err = l.GetGapDoDSettleStateBlock(internalId, func(blk *types.StateBlock, sync types.SynchronizedKind) error {
		if blk.GetHash() != block.GetHash() {
			t.Fatal()
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeleteGapDoDSettleStateBlock(internalId, block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.GetGapDoDSettleStateBlock(internalId, func(blk *types.StateBlock, sync types.SynchronizedKind) error {
		t.Fatal()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_PovHeightAddGap(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	height := uint64(100)
	err := l.PovHeightAddGap(height)
	if err != nil {
		t.Fatal(err)
	}

	has, err := l.PovHeightHasGap(height)
	if err != nil || !has {
		t.Fatal(err)
	}

	err = l.PovHeightDeleteGap(height)
	if err != nil {
		t.Fatal(err)
	}

	has, err = l.PovHeightHasGap(height)
	if err != nil || has {
		t.Fatal(err)
	}
}
