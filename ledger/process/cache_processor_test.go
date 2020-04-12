package process

import (
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func TestProcess_CacheBlockProcess(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	if err := lv.BlockCacheProcess(bc[0]); err != nil {
		t.Fatal(err)
	}
	if err := lv.BlockProcess(bc[0]); err != nil {
		t.Fatal(Other, err)
	}
	t.Log("bc hash", bc[0].GetHash())
	for i, b := range bc[1:] {
		fmt.Println(i + 1)
		fmt.Println("bc.previous", b.GetPrevious())
		if r, err := lv.BlockCacheCheck(b); r != Progress || err != nil {
			t.Fatal(r, err)
		}
		if err := lv.BlockCacheProcess(b); err != nil {
			t.Fatal(Other, err)
		}
		if err := lv.BlockProcess(b); err != nil {
			t.Fatal(Other, err)
		}
	}
}

func TestProcess_CacheContractBlockProcess(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	bs := mock.ContractBlocks()
	if err := lv.BlockCacheProcess(bs[0]); err != nil {
		t.Fatal(err)
	}
	if err := lv.BlockProcess(bs[0]); err != nil {
		t.Fatal(Other, err)
	}
	for i, b := range bs[1:] {
		fmt.Println(i)
		if r, err := lv.BlockCacheCheck(b); r != Progress || err != nil {
			t.Fatal(r, err)
		}
		if err := lv.BlockCacheProcess(b); err != nil {
			t.Fatal(Other, err)
		}
		if err := lv.BlockProcess(b); err != nil {
			t.Fatal(Other, err)
		}
	}
}

func TestProcess_BlockRepeated(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)

	sendblk := mock.StateBlockWithoutWork()
	r1 := mock.StateBlockWithoutWork()
	r1.Link = sendblk.GetHash()
	r2 := mock.StateBlockWithoutWork()
	r2.Link = sendblk.GetHash()
	if err := l.AddBlockCache(r1); err != nil {
		t.Fatal(err)
	}
	if err := checkReceiveBlockRepeat(lv, r2); err != ReceiveRepeated {
		t.Fatal(err)
	}
}

func TestProcess_CacheException(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	genesisBlk := config.GenesisBlock()
	if r, err := lv.BlockCacheCheck(&genesisBlk); err != nil || r != Progress {
		t.Fatal(r, err)
	}

	// invalid type
	if r, err := lv.BlockCacheCheck(&types.StateBlock{}); err == nil {
		t.Fatal(r, err)
	}

	// open
	bc[0].Signature, _ = types.NewSignature("5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600")
	if r, err := lv.BlockCacheCheck(bc[0]); r != BadSignature {
		t.Fatal(r, err)
	}
	if r, err := lv.BlockCacheCheck(bc[1]); r != GapPrevious {
		t.Fatal(r, err)
	}

	if err := lv.BlockCacheProcess(bc[0]); err != nil {
		t.Fatal(err)
	}
	if r, err := lv.BlockCacheCheck(bc[0]); r != Old {
		t.Fatal(r, err)
	}

	// open gapSource
	if r, err := lv.BlockCacheCheck(bc[2]); r != GapSource {
		t.Fatal(r, err)
	}

	// send
	if r, err := lv.BlockCacheCheck(bc[4]); r != GapPrevious {
		t.Fatal(r, err)
	}

	// receive
	if r, err := lv.BlockCacheCheck(bc[5]); r != GapPrevious {
		t.Fatal(r, err)
	}

	// contract block
	bc := mock.StateBlockWithoutWork()
	bc.Type = types.ContractReward
	if r, err := lv.BlockCacheCheck(bc); r != GapSource {
		t.Fatal(r, err)
	}
	bc.Type = types.ContractSend
	bc.Link = contractaddress.NEP5PledgeAddress.ToHash()
	if r, err := lv.BlockCacheCheck(bc); r == Progress {
		t.Fatal(r, err)
	}

	bs := mock.ContractBlocks()
	if r, err := lv.BlockCacheCheck(bs[1]); r != GapPrevious {
		t.Fatal(r, err)
	}
}
