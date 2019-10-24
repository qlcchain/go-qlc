package process

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

func TestProcess_Rollback(t *testing.T) {
	teardownTestCase, l, lv := setupTestCase(t)
	defer teardownTestCase(t)
	var bc, _ = mock.BlockChain(false)
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

func TestLedgerVerifier_BlockCacheCheck(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)
	addr := mock.Address()
	ac := mock.AccountMeta(addr)
	token := ac.Tokens[0].Type
	block := mock.StateBlock()
	block.Address = addr
	block.Token = token

	if err := lv.l.AddBlockCache(block); err != nil {
		t.Fatal(err)
	}
	if err := lv.l.AddAccountMetaCache(ac); err != nil {
		t.Fatal(err)
	}
	if err := lv.RollbackBlock(block.GetHash()); err != nil {
		t.Fatal(err)
	}

	if b, err := lv.l.HasBlockCache(block.GetHash()); b || err != nil {
		t.Fatal(err)
	}
	ac, err := lv.l.GetAccountMetaCache(addr)
	if err != nil {
		t.Fatal(err)
	}
	if tm := ac.Token(token); tm != nil {
		t.Fatal(err)
	}
}

func TestLedger_Rollback_ContractData(t *testing.T) {
	t.Skip()
	dir := filepath.Join(config.DefaultDataDir(), "ledger")
	t.Log(dir)
	cm := config.NewCfgManager(dir)
	cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)
	lv := NewLedgerVerifier(l)

	defer func() {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	hash := new(types.Hash)
	hash.Of("06910e3bfea136a891709f2bace609f12135d8766325fca328fc519e9f638157")
	block, err := lv.l.GetStateBlock(*hash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(block)
}
