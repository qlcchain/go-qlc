// +build integrate

package test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/rpc/api"
)

const (
	privateKey = "504ef6b87d98ec0a9b882f1053d8903090cd4955f444f239fea294be38db6cc9123a54070e4ab8007093536def7092596da77d280074bc133c2ab96a2ceb8f76"
	address    = "qlc_16jtci5iwkor13rb8nufxxrb6pdfnxyki15nqibmrcosfapgq5up3d167isq"
)

func TestProcess_Open(t *testing.T) {
	teardownTestCase, _, ls, err := generateChain()
	defer func() {
		if err := teardownTestCase(); err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	l := ls.Ledger
	lv := process.NewLedgerVerifier(l)
	addr, err := types.HexToAddress(address)
	if err != nil {
		t.Fatal(err)
	}
	sb := types.StateBlock{
		Address: testReceiveBlock.Address,
		Token:   testReceiveBlock.Token,
		Link:    addr.ToHash(),
	}

	prk, err := hex.DecodeString(testPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	am := types.Balance{Int: big.NewInt(int64(100000000000))}
	b, err := l.GenerateSendBlock(&sb, am, prk)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b.String())
	p, _ := lv.Process(b)
	if p != process.Progress {
		t.Fatal("process send block error")
	}
	receivePrk, err := hex.DecodeString(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	r, err := l.GenerateReceiveBlock(b, receivePrk)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.String())
	err = l.DeleteStateBlock(b.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	// test GapSource
	p, err = lv.BlockCheck(r)
	if p != process.GapSource {
		t.Fatal("result should GapSource")
	}
	err = l.AddStateBlock(b)
	if err != nil {
		t.Fatal(err)
	}

	//test Progress
	p, err = lv.BlockCheck(r)
	if p != process.Progress {
		t.Fatal("process open block error")
	}

	temp := r.Clone()
	r.Vote = types.Balance{Int: big.NewInt(int64(1))}
	r.Work = 0

	// test BadWork
	p, err = lv.BlockCheck(r)
	if p != process.BadWork {
		t.Fatal("result should BadWork")
	}
	r.Work = temp.Work

	// test BadSignature
	p, err = lv.BlockCheck(r)
	if p != process.BadSignature {
		t.Fatal("result should BadSignature")
	}
	acc := types.NewAccount(receivePrk)
	r.Signature = acc.Sign(r.GetHash())

	// test BalanceMismatch
	p, err = lv.BlockCheck(r)
	if p != process.BalanceMismatch {
		t.Fatal("result should BalanceMismatch")
	}

	r.Vote = types.Balance{Int: big.NewInt(int64(0))}
	r.Signature = temp.Signature
	p, err = lv.BlockCheck(r)
	p, _ = lv.Process(r)
	if p != process.Progress {
		t.Fatal("result should progress")
	}

	// test Old
	p, _ = lv.Process(r)
	if p != process.Old {
		t.Fatal("result should old")
	}

	// test fork
	temp.Timestamp = 1558854606
	temp.Signature = acc.Sign(temp.GetHash())
	t.Log(temp.String())
	p, err = lv.BlockCheck(temp)
	if p != process.Fork {
		t.Fatal("result should Fork")
	}
}

func TestProcess_Send(t *testing.T) {
	teardownTestCase, _, ls, err := generateChain()
	defer func() {
		if err := teardownTestCase(); err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	l := ls.Ledger
	lv := process.NewLedgerVerifier(l)
	addr, err := types.HexToAddress(address)
	if err != nil {
		t.Fatal(err)
	}
	sb := types.StateBlock{
		Address: testReceiveBlock.Address,
		Token:   testReceiveBlock.Token,
		Link:    addr.ToHash(),
	}

	prk, err := hex.DecodeString(testPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	am := types.Balance{Int: big.NewInt(int64(100000000000))}
	b, err := l.GenerateSendBlock(&sb, am, prk)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b.String())

	//test Progress
	p, _ := lv.BlockCheck(b)
	if p != process.Progress {
		t.Fatal("process send block error")
	}
	temp := b.Clone()
	err = l.DeleteStateBlock(testReceiveBlock.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	//test GapPrevious
	p, _ = lv.BlockCheck(b)
	if p != process.GapPrevious {
		t.Fatal("result should be GapPrevious")
	}

	err = l.AddStateBlock(&testReceiveBlock)
	if err != nil {
		t.Fatal(err)
	}
	b.Balance = b.Balance.Add(types.Balance{Int: big.NewInt(int64(200000000000))})
	acc := types.NewAccount(prk)
	b.Signature = acc.Sign(b.GetHash())

	//test BalanceMismatch
	p, _ = lv.BlockCheck(b)
	if p != process.BalanceMismatch {
		t.Fatal("result should be BalanceMismatch")
	}

	//test fork
	p, _ = lv.Process(temp)
	if p != process.Progress {
		t.Fatal("process send block error")
	}
	b.Balance = b.Balance.Add(types.Balance{Int: big.NewInt(int64(1000000000))})
	b.Signature = acc.Sign(b.GetHash())
	p, _ = lv.Process(b)
	if p != process.Fork {
		t.Fatal("result should be fork")
	}

}

func TestProcess_Change(t *testing.T) {
	teardownTestCase, _, ls, err := generateChain()
	defer func() {
		if err := teardownTestCase(); err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	l := ls.Ledger
	lv := process.NewLedgerVerifier(l)

	prk, err := hex.DecodeString(testPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	cb, err := l.GenerateChangeBlock(testReceiveBlock.Address, testReceiveBlock.Address, prk)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := lv.BlockCheck(cb)

	//test progress
	if p != process.Progress {
		t.Fatal("process change block error")
	}
	temp := cb.Clone()
	err = l.DeleteStateBlock(testReceiveBlock.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	//test GapPrevious
	p, _ = lv.BlockCheck(cb)
	if p != process.GapPrevious {
		t.Fatal("result should be GapPrevious")
	}

	err = l.AddStateBlock(&testReceiveBlock)
	if err != nil {
		t.Fatal(err)
	}
	cb.Balance = cb.Balance.Add(types.Balance{Int: big.NewInt(int64(200000000000))})
	acc := types.NewAccount(prk)
	cb.Signature = acc.Sign(cb.GetHash())

	//test BalanceMismatch
	p, _ = lv.BlockCheck(cb)
	if p != process.BalanceMismatch {
		t.Fatal("result should be BalanceMismatch")
	}

	//test fork
	p, _ = lv.Process(temp)
	if p != process.Progress {
		t.Fatal("process send block error")
	}
	cb.Balance = cb.Balance.Add(types.Balance{Int: big.NewInt(int64(1000000000))})
	cb.Signature = acc.Sign(cb.GetHash())
	p, _ = lv.Process(cb)
	if p != process.Fork {
		t.Fatal("result should be fork")
	}
}

func TestProcess_ContractSend(t *testing.T) {
	teardownTestCase, client, ls, err := generateChain()
	defer func() {
		if err := teardownTestCase(); err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	l := ls.Ledger
	lv := process.NewLedgerVerifier(l)

	//test genesis contractSend block
	genesis := common.GenesisMintageBlock()
	p, _ := lv.BlockCheck(&genesis)
	if p != process.Progress {
		t.Fatal("process genesis contractSend block error")
	}

	selfBytes, err := hex.DecodeString(testPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	s := types.NewAccount(selfBytes)
	beneficialBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	b := types.NewAccount(beneficialBytes)
	id := mock.Hash()
	param := &api.RewardsParam{
		Id:     hex.EncodeToString(id[:]),
		Amount: types.Balance{Int: big.NewInt(10e7)},
		Self:   s.Address(),
		To:     b.Address(),
	}
	var hash types.Hash
	err = client.Call(&hash, "rewards_getUnsignedRewardData", param)
	if err != nil {
		t.Fatal(err)
	}

	sign := s.Sign(hash)
	var blk types.StateBlock
	err = client.Call(&blk, "rewards_getSendRewardBlock", param, &sign)
	if err != nil {
		t.Fatal(err)
	}
	temp := blk.Clone()
	// test progress
	p, _ = lv.BlockCheck(&blk)
	if p != process.Progress {
		t.Fatal("process result should be progress")
	}
	//test BalanceMismatch
	blk.Vote = blk.Vote.Add(types.Balance{Int: big.NewInt(int64(100000000000000))})
	p, _ = lv.BlockCheck(&blk)
	if p != process.BalanceMismatch {
		t.Fatal("process result should be BalanceMismatch")
	}
	blk.Vote = temp.Vote
	//test data
	blk.Data[0] = 0x01

	p, _ = lv.BlockCheck(&blk)
	if p != process.Other {
		t.Fatal("process result should be InvalidData")
	}
	blk.Data = temp.Data
	blk.Work = 1
	blk.Signature = testReceiveBlock.Signature

	//test Work,Signature checck,if contractSend block is generate by rewards contract,will not check Work and Signature
	p, _ = lv.Process(&blk)
	if p != process.Progress {
		t.Fatal("should not check work and signature")
	}

	//test Fork
	blk.Timestamp = 1558854606
	p, err = lv.BlockCheck(&blk)
	if p != process.Fork {
		t.Fatal("result should Fork")
	}
}

func TestProcess_ContractReceive(t *testing.T) {
	teardownTestCase, client, ls, err := generateChain()
	defer func() {
		if err := teardownTestCase(); err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	l := ls.Ledger
	lv := process.NewLedgerVerifier(l)

	//test genesis contractSend block
	genesis := common.GenesisBlock()
	p, _ := lv.BlockCheck(&genesis)
	if p != process.Progress {
		t.Fatal("process genesis contractReceive block error")
	}

	selfBytes, err := hex.DecodeString(testPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	s := types.NewAccount(selfBytes)
	beneficialBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	b := types.NewAccount(beneficialBytes)
	id := mock.Hash()
	param := &api.RewardsParam{
		Id:     hex.EncodeToString(id[:]),
		Amount: types.Balance{Int: big.NewInt(10e7)},
		Self:   s.Address(),
		To:     b.Address(),
	}
	var hash types.Hash
	err = client.Call(&hash, "rewards_getUnsignedRewardData", param)
	if err != nil {
		t.Fatal(err)
	}

	sign := s.Sign(hash)
	var send types.StateBlock
	err = client.Call(&send, "rewards_getSendRewardBlock", param, &sign)
	if err != nil {
		t.Fatal(err)
	}
	p, _ = lv.Process(&send)
	if p != process.Progress {
		t.Fatal("process contractSend block error")
	}
	var receive types.StateBlock
	err = client.Call(&receive, "rewards_getReceiveRewardBlock", send.GetHash())
	if err != nil {
		fmt.Println(err)
	}
	//test progress
	p, _ = lv.BlockCheck(&receive)
	if p != process.Progress {
		t.Fatal("process contractReceive block error")
	}
	err = l.DeleteStateBlock(send.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)

	//test GapSource
	p, _ = lv.BlockCheck(&receive)
	if p != process.GapSource {
		t.Fatal("process reslut should be GapSource")
	}
	err = l.AddStateBlock(&send)
	if err != nil {
		t.Fatal(err)
	}
	p, _ = lv.Process(&receive)
	if p != process.Progress {
		t.Fatal("process reslut should be Progress")
	}
	//test Fork
	receive.Timestamp = 1558854606
	p, err = lv.BlockCheck(&receive)
	if p != process.Fork {
		t.Fatal("result should Fork")
	}
}
