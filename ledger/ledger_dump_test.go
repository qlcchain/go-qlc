package ledger

import (
	"fmt"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestLedger_dump(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlockWithoutWork()
	acc1 := mock.AccountMeta(blk1.Address)
	blk1.Address = acc1.Address
	blk1.Token = acc1.Tokens[0].Type
	acc1.Tokens[0].OpenBlock = blk1.GetHash()

	blk2 := mock.StateBlockWithoutWork()
	blk2.Previous = blk1.GetHash()
	blk2.Address = acc1.Address
	blk2.Token = acc1.Tokens[0].Type
	acc1.Tokens[0].Header = blk2.GetHash()

	blk3 := mock.StateBlockWithoutWork()
	blk3.Address = acc1.Address
	blk3.Token = acc1.Tokens[0].Type

	if err := l.AddStateBlock(blk1); err != nil {
		t.Fatal(err)
	}
	if err := l.AddStateBlock(blk2); err != nil {
		t.Fatal(err)
	}
	if err := l.AddStateBlock(blk3); err != nil {
		t.Fatal(err)
	}
	if err := l.AddAccountMeta(acc1, l.cache.GetCache()); err != nil {
		t.Fatal(err)
	}

	blkCache1 := mock.StateBlockWithoutWork()
	accCache1 := mock.AccountMeta(blkCache1.Address)
	blkCache1.Token = accCache1.Tokens[0].Type
	blkCache2 := mock.StateBlockWithoutWork()
	accCache2 := mock.AccountMeta(blkCache2.Address)
	blkCache2.Token = accCache2.Tokens[0].Type
	if err := l.AddBlockCache(blkCache1); err != nil {
		t.Fatal(err)
	}
	if err := l.AddBlockCache(blkCache2); err != nil {
		t.Fatal(err)
	}
	if err := l.AddAccountMetaCache(accCache1); err != nil {
		t.Fatal(err)
	}
	if err := l.AddUncheckedBlock(mock.Hash(), mock.StateBlockWithoutWork(), types.UncheckedKindLink, types.UnSynchronized); err != nil {
		fmt.Println(err)
		return
	}
	if _, err := l.Dump(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)
}
