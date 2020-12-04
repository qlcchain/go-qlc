/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package process

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

var bc []*types.StateBlock

func setupTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger, *LedgerVerifier) {
	//t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	//dir := filepath.Join(config.DefaultDataDir()) // if want to test rollback contract and remove time sleep

	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cfg, _ := cm.Load()
	cfg.TrieClean.Enable = false
	cm.Save()

	l := ledger.NewLedger(cm.ConfigFile)
	bc, _ = mock.BlockChain(false)
	fmt.Println(t.Name())
	setPovStatus(l, t)
	return func(t *testing.T) {
		//err := l.db.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l, NewLedgerVerifier(l)
}

func setPovStatus(l *ledger.Ledger, t *testing.T) {
	block, td := mock.GeneratePovBlock(nil, 0)
	block.Header.BasHdr.Height = 0
	if err := l.AddPovBlock(block, td); err != nil {
		t.Fatal(err)
	}
	if err := l.AddPovBestHash(block.GetHeight(), block.GetHash()); err != nil {
		t.Fatal(err)
	}
	if err := l.SetPovLatestHeight(block.GetHeight()); err != nil {
		t.Fatal(err)
	}
}

func TestProcess_BlockProcess(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
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
}

func TestProcess_ContractBlockProcess(t *testing.T) {
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
}

func TestProcess_Exception(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	checkResults := []ProcessResult{
		Progress,
		BadWork,
		BadSignature,
		BadHash,
		BadMerkleRoot,
		BadTarget,
		BadStateHash,
		BadCoinbase,
		Old,
		Fork,
		GapPrevious,
		GapSource,
		GapSmartContract,
		GapTransaction,
		GapTokenInfo,
		GapPovHeight,
		GapPublish,
		BalanceMismatch,
		UnReceivable,
		InvalidData,
		InvalidTime,
		InvalidTxNum,
		InvalidHeight,
		InvalidTxOrder,
		BadConsensus,
		ReceiveRepeated,
		BadAuxHeader,
	}

	for _, r := range checkResults {
		if r.String() == "<invalid>" {
			t.Fatal(r.String())
		}
	}

	genesisBlk := config.GenesisBlock()
	if r, err := lv.BlockCheck(&genesisBlk); err != nil || r != Progress {
		t.Fatal(r, err)
	}

	// invalid type
	if r, err := lv.BlockCheck(&types.StateBlock{}); err == nil {
		t.Fatal(r, err)
	}

	// open
	bc[0].Signature, _ = types.NewSignature("5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600")
	if r, err := lv.BlockCheck(bc[0]); r != BadSignature {
		t.Fatal(r, err)
	}
	if r, err := lv.BlockCheck(bc[1]); r != GapPrevious {
		t.Fatal(r, err)
	}

	if err := lv.BlockProcess(bc[0]); err != nil {
		t.Fatal(err)
	}
	if r, err := lv.BlockCheck(bc[0]); r != Old {
		t.Fatal(r, err)
	}

	// open gapSource
	if r, err := lv.BlockCheck(bc[2]); r != GapSource {
		t.Fatal(r, err)
	}

	// receive unReceivable
	if err := lv.BlockProcess(bc[1]); err != nil {
		t.Fatal(err)
	}
	if err := lv.l.DeletePending(&types.PendingKey{
		Address: bc[2].Address,
		Hash:    bc[1].GetHash(),
	}, lv.l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	if r, err := lv.BlockCheck(bc[2]); r != UnReceivable {
		t.Fatal(r, err)
	}

	// send
	if r, err := lv.BlockCheck(bc[4]); r != GapPrevious {
		t.Fatal(r, err)
	}

	// receive
	if r, err := lv.BlockCheck(bc[5]); r != GapSource {
		t.Fatal(r, err)
	}

	// contract block
	bc := mock.StateBlockWithoutWork()
	bc.Type = types.ContractReward
	if r, err := lv.BlockCheck(bc); r != GapSource {
		t.Fatal(r, err)
	}
	bc.Type = types.ContractSend
	bc.Link = contractaddress.NEP5PledgeAddress.ToHash()
	if r, err := lv.BlockCheck(bc); r == Progress {
		t.Fatal(r, err)
	}

	bs := mock.ContractBlocks()
	if r, err := lv.BlockCheck(bs[1]); r != GapPrevious {
		t.Fatal(r, err)
	}

	if err := lv.BlockProcess(bs[1]); err == nil {
		t.Fatal(err)
	}
}

func TestBlock_fork(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)

	// fork check
	forkCheck := blockForkCheck{}
	am := mock.AccountMeta(mock.Address())
	if err := l.AddAccountMeta(am, l.Cache().GetCache()); err != nil {
		t.Fatal()
	}
	pre := mock.StateBlockWithoutWork()
	preKey, _ := storage.GetKeyOfParts(storage.KeyPrefixBlock, pre.GetHash())
	preVal, _ := pre.Serialize()
	if err := l.DBStore().Put(preKey, preVal); err != nil {
		t.Fatal(err)
	}
	blk := mock.StateBlockWithoutWork()
	blk.Previous = pre.GetHash()
	blk.Address = am.Address
	blk.Token = am.Tokens[0].Type
	r, err := forkCheck.fork(lv, blk)
	if r != Fork {
		t.Fatal(r, err)
	}
	blk.Type = types.Send
	r, err = forkCheck.fork(lv, blk)
	if r != Fork {
		t.Fatal(r, err)
	}
	blk.Type = types.ContractReward
	r, err = forkCheck.fork(lv, blk)
	if r != Fork {
		t.Fatal(r, err)
	}
	blk.Previous = types.ZeroHash
	r, err = forkCheck.fork(lv, blk)
	if r != Fork {
		t.Fatal(r, err)
	}
	r, err = forkCheck.fork(lv, &types.StateBlock{})
	if err == nil {
		t.Fatal(r, err)
	}

	// cache fork  check
	cForkCheck := cacheBlockForkCheck{}
	blk.Type = types.Open
	r, err = cForkCheck.fork(lv, blk)
	if r != Fork {
		t.Fatal(r, err)
	}
	blk.Previous = pre.GetHash()
	blk.Type = types.Send
	r, err = cForkCheck.fork(lv, blk)
	if r != Fork {
		t.Fatal(r, err)
	}
	blk.Type = types.ContractReward
	r, err = cForkCheck.fork(lv, blk)
	if r != Fork {
		t.Fatal(r, err)
	}
	blk.Previous = types.ZeroHash
	r, err = cForkCheck.fork(lv, blk)
	if r != Fork {
		t.Fatal(r, err)
	}
	r, err = cForkCheck.fork(lv, &types.StateBlock{})
	if err == nil {
		t.Fatal(r, err)
	}
}
