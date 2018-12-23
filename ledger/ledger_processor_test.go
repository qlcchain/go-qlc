package ledger

import (
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func TestProcess_BlockBasicInfoCheck(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	blks := parseBlocks(t, "testdata/blocks_ledger_process.json")

	logger.Info("------ genesis, addr(1c47) ------")
	err := l.BatchUpdate(func(txn db.StoreTxn) error {
		state := blks[0].(*types.StateBlock)
		return l.addBasicInfo(state, false, txn)
	})
	if err != nil {
		t.Fatal(err)
	}

	logger.Info("------ addr(1c47) send to addr(1zbo) 6000 ------")
	r := l.Process(blks[1])
	if r == Other {
		t.Fatal(r)
	}

	logger.Info("------ addr(1zbo) open  ------")
	r = l.Process(blks[2])
	if r == Other {
		t.Fatal(r)
	}

	logger.Info("------ addr(1zbo) change rep to (1zbo) ------")
	r = l.Process(blks[3])
	if r == Other {
		t.Fatal(r)
	}

	logger.Info("------ addr(1zbo) send to addr(1c47) 4000 ------")
	r = l.Process(blks[4])
	if r == Other {
		t.Fatal(r)
	}

	logger.Info("------ addr(1c47) receive ------")
	r = l.Process(blks[5])
	if r == Other {
		t.Fatal(r)
	}

	logger.Info("------ add token ------")
	err = l.BatchUpdate(func(txn db.StoreTxn) error {
		state := blks[6].(*types.StateBlock)
		return l.addBasicInfo(state, false, txn)
	})
	if err != nil {
		t.Fatal(err)
	}
	checkInfo(t, l)
}

func checkInfo(t *testing.T, l *Ledger) {
	blocks, _ := l.GetBlocks()
	fmt.Println("----blocks: ")
	for _, b := range blocks {
		fmt.Println(*b)
	}

	fmt.Println("----frontiers:")
	fs, _ := l.GetFrontiers()
	for _, f := range fs {
		fmt.Println(f)
	}
	fmt.Println("----account: ")
	addr1, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	addr2, _ := types.HexToAddress("qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i")
	ac, _ := l.GetAccountMeta(addr1)
	fmt.Println(" account1,", ac.Address)
	for _, t := range ac.Tokens {
		fmt.Println("  token, ", t)
	}
	ac, err := l.GetAccountMeta(addr2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(" account2,", ac.Address)
	for _, t := range ac.Tokens {
		fmt.Println("  token, ", t)
	}

	fmt.Println("----representation:")
	b, err := l.GetRepresentation(addr1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(addr1, b)
	b, err = l.GetRepresentation(addr2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(addr2, b)
}
