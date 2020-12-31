package ledger

import (
	"bytes"
	"testing"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/mock"
)

func TestLedger_BlockPrivatePayload(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk := mock.StateBlockWithoutWork()
	blkHash := blk.GetHash()
	pl := util.RandomFixedBytes(256)

	err := l.AddBlockPrivatePayload(blkHash, pl)
	if err != nil {
		t.Fatal(err)
	}

	plRet, err := l.GetBlockPrivatePayload(blkHash)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(plRet, pl) != 0 {
		t.Fatal("payload not equal")
	}

	err = l.DeleteBlockPrivatePayload(blkHash)
	if err != nil {
		t.Fatal(err)
	}

	plRet, err = l.GetBlockPrivatePayload(blkHash)
	if err == nil {
		t.Fatal("deleted payload exist")
	}
}

func TestLedger_BlockPrivatePayload_Cache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blk := mock.StateBlockWithoutWork()
	blkHash := blk.GetHash()
	pl := util.RandomFixedBytes(256)

	err := l.AddBlockPrivatePayload(blkHash, pl, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}

	plRet, err := l.GetBlockPrivatePayload(blkHash, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(plRet, pl) != 0 {
		t.Fatal("payload not equal")
	}

	err = l.DeleteBlockPrivatePayload(blkHash, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}

	plRet, err = l.GetBlockPrivatePayload(blkHash, l.cache.GetCache())
	if err == nil {
		t.Fatal("deleted payload exist")
	}
}
