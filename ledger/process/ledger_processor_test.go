/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package process

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/qlcchain/go-qlc/test"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

func storeGenesisBlock(l *ledger.Ledger, verifier *LedgerVerifier) error {
	genesis := common.GenesisBlock()
	ctx := vmstore.NewVMContext(l)
	err := ctx.SetStorage(types.MintageAddress[:], genesis.Token[:], genesis.Data)
	if err != nil {
		return err
	}
	mintageHash := common.GenesisMintageHash()
	if b, err := l.HasStateBlock(mintageHash); !b && err == nil {
		mintage := common.GenesisMintageBlock()
		if err := l.AddStateBlock(&mintage); err != nil {
			return err
		}
	} else {
		if err != nil {
			return err
		}
	}

	genesisHash := common.GenesisBlockHash()
	if b, err := l.HasStateBlock(genesisHash); !b && err == nil {
		if err := verifier.BlockProcess(&genesis); err != nil {
			return err
		}
	} else {
		if err != nil {
			return err
		}
	}

	//gas block storage
	gas := common.GasBlock()
	err = ctx.SetStorage(types.MintageAddress[:], gas.Token[:], gas.Data)
	if err != nil {
		return err
	}

	_ = ctx.SaveStorage()

	gasMintageHash := common.GasMintageHash()
	if b, err := l.HasStateBlock(gasMintageHash); !b && err == nil {
		gasMintage := common.GasMintageBlock()
		if err := l.AddStateBlock(&gasMintage); err != nil {
			return err
		}
	} else {
		if err != nil {
			return err
		}
	}

	gasHash := common.GasBlockHash()
	if b, err := l.HasStateBlock(gasHash); !b && err == nil {
		if err := verifier.BlockProcess(&gas); err != nil {
			return err
		}
	} else {
		if err != nil {
			return err
		}
	}
	return nil
}

func setupTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger, *LedgerVerifier) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	l := ledger.NewLedger(dir)

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

//func TestProcess_BlockCheck(t *testing.T) {
//	teardownTestCase, l, lv := setupTestCase(t)
//	defer teardownTestCase(t)
//	blockCheck(t, lv)
//	checkInfo(t, l)
//}

func TestLedger_Rollback(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)
	genesis := common.GenesisBlock()
	var bc, _ = mock.BlockChain()
	if err := lv.l.AddStateBlock(&genesis); err != nil {
		t.Fatal(err)
	}
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

	h := bc[5].GetHash()
	if err := l.Rollback(h); err != nil {
		t.Fatal(err)
	}
	checkInfo(t, l)
}

func checkInfo(t *testing.T, l *ledger.Ledger) {
	addrs := make(map[types.Address]int)
	fmt.Println("----blocks----")
	err := l.GetStateBlocks(func(block *types.StateBlock) error {
		fmt.Println(block)
		if block.GetHash() != common.GenesisBlockHash() {
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
	for k, _ := range addrs {
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
	for k, _ := range addrs {
		b, err := l.GetRepresentation(k)
		if err != nil {
			if err == ledger.ErrRepresentationNotFound {
			}
		} else {
			fmt.Println(k, b)
		}
	}
}

var send types.StateBlock
var receive types.StateBlock
var privateKey = "504ef6b87d98ec0a9b882f1053d8903090cd4955f444f239fea294be38db6cc9123a54070e4ab8007093536def7092596da77d280074bc133c2ab96a2ceb8f76"
var address = "qlc_16jtci5iwkor13rb8nufxxrb6pdfnxyki15nqibmrcosfapgq5up3d167isq"

func TestProcess_Open(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)
	err := storeGenesisBlock(l, lv)
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal([]byte(test.JsonTestSend), &send)
	_ = json.Unmarshal([]byte(test.JsonTestReceive), &receive)

	p, _ := lv.Process(&send)
	if p != Progress {
		t.Fatal("process send block error")
	}
	p, _ = lv.Process(&receive)
	if p != Progress {
		t.Fatal("process receive block error")
	}
	addr, err := types.HexToAddress(address)
	if err != nil {
		t.Fatal(err)
	}
	sb := types.StateBlock{
		Address: receive.Address,
		Token:   receive.Token,
		Link:    addr.ToHash(),
	}

	prk, err := hex.DecodeString(test.TestPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	am := types.Balance{Int: big.NewInt(int64(100000000000))}
	b, err := l.GenerateSendBlock(&sb, am, prk)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b.String())
	p, _ = lv.Process(b)
	if p != Progress {
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

	// test GapSource
	p, err = lv.BlockCheck(r)
	if p != GapSource {
		t.Fatal("result should GapSource")
	}
	err = l.AddStateBlock(b)
	if err != nil {
		t.Fatal(err)
	}

	//test Progress
	p, err = lv.BlockCheck(r)
	if p != Progress {
		t.Fatal("process open block error")
	}

	temp := r.Clone()
	r.Vote = types.Balance{Int: big.NewInt(int64(1))}
	r.Work = 0

	// test BadWork
	p, err = lv.BlockCheck(r)
	if p != BadWork {
		t.Fatal("result should BadWork")
	}
	r.Work = temp.Work

	// test BadSignature
	p, err = lv.BlockCheck(r)
	if p != BadSignature {
		t.Fatal("result should BadSignature")
	}
	acc := types.NewAccount(receivePrk)
	r.Signature = acc.Sign(r.GetHash())

	// test BalanceMismatch
	p, err = lv.BlockCheck(r)
	if p != BalanceMismatch {
		t.Fatal("result should BalanceMismatch")
	}

	r.Vote = types.Balance{Int: big.NewInt(int64(0))}
	r.Signature = temp.Signature
	p, err = lv.BlockCheck(r)
	p, _ = lv.Process(r)
	if p != Progress {
		t.Fatal("result should progress")
	}

	// test Old
	p, _ = lv.Process(r)
	if p != Old {
		t.Fatal("result should old")
	}

	// test fork
	temp.Timestamp = 1558854606
	temp.Signature = acc.Sign(temp.GetHash())
	t.Log(temp.String())
	p, err = lv.BlockCheck(temp)
	if p != Fork {
		t.Fatal("result should Fork")
	}
}

func TestProcess_Send(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)
	err := storeGenesisBlock(l, lv)
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal([]byte(test.JsonTestSend), &send)
	_ = json.Unmarshal([]byte(test.JsonTestReceive), &receive)

	p, _ := lv.Process(&send)
	if p != Progress {
		t.Fatal("process send block error")
	}
	p, _ = lv.Process(&receive)
	if p != Progress {
		t.Fatal("process receive block error")
	}
	addr, err := types.HexToAddress(address)
	if err != nil {
		t.Fatal(err)
	}
	sb := types.StateBlock{
		Address: receive.Address,
		Token:   receive.Token,
		Link:    addr.ToHash(),
	}

	prk, err := hex.DecodeString(test.TestPrivateKey)
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
	p, _ = lv.BlockCheck(b)
	if p != Progress {
		t.Fatal("process send block error")
	}
	temp := b.Clone()
	err = l.DeleteStateBlock(receive.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	//test GapPrevious
	p, _ = lv.BlockCheck(b)
	if p != GapPrevious {
		t.Fatal("result should be GapPrevious")
	}

	err = l.AddStateBlock(&receive)
	if err != nil {
		t.Fatal(err)
	}
	b.Balance = b.Balance.Add(types.Balance{Int: big.NewInt(int64(200000000000))})
	acc := types.NewAccount(prk)
	b.Signature = acc.Sign(b.GetHash())

	//test BalanceMismatch
	p, _ = lv.BlockCheck(b)
	if p != BalanceMismatch {
		t.Fatal("result should be BalanceMismatch")
	}

	//test fork
	p, _ = lv.Process(temp)
	if p != Progress {
		t.Fatal("process send block error")
	}
	b.Balance = b.Balance.Add(types.Balance{Int: big.NewInt(int64(1000000000))})
	b.Signature = acc.Sign(b.GetHash())
	p, _ = lv.Process(b)
	if p != Fork {
		t.Fatal("result should be fork")
	}

}

func TestProcess_Change(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)
	err := storeGenesisBlock(l, lv)
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal([]byte(test.JsonTestSend), &send)
	_ = json.Unmarshal([]byte(test.JsonTestReceive), &receive)

	p, _ := lv.Process(&send)
	if p != Progress {
		t.Fatal("process send block error")
	}
	p, _ = lv.Process(&receive)
	if p != Progress {
		t.Fatal("process receive block error")
	}

	prk, err := hex.DecodeString(test.TestPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	cb, err := l.GenerateChangeBlock(receive.Address, receive.Address, prk)
	if err != nil {
		t.Fatal(err)
	}
	p, _ = lv.BlockCheck(cb)

	//test progress
	if p != Progress {
		t.Fatal("process change block error")
	}
	temp := cb.Clone()
	err = l.DeleteStateBlock(receive.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	//test GapPrevious
	p, _ = lv.BlockCheck(cb)
	if p != GapPrevious {
		t.Fatal("result should be GapPrevious")
	}

	err = l.AddStateBlock(&receive)
	if err != nil {
		t.Fatal(err)
	}
	cb.Balance = cb.Balance.Add(types.Balance{Int: big.NewInt(int64(200000000000))})
	acc := types.NewAccount(prk)
	cb.Signature = acc.Sign(cb.GetHash())

	//test BalanceMismatch
	p, _ = lv.BlockCheck(cb)
	if p != BalanceMismatch {
		t.Fatal("result should be BalanceMismatch")
	}

	//test fork
	p, _ = lv.Process(temp)
	if p != Progress {
		t.Fatal("process send block error")
	}
	cb.Balance = cb.Balance.Add(types.Balance{Int: big.NewInt(int64(1000000000))})
	cb.Signature = acc.Sign(cb.GetHash())
	p, _ = lv.Process(cb)
	if p != Fork {
		t.Fatal("result should be fork")
	}
}
