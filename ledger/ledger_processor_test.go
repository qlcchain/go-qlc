package ledger

import (
	"fmt"
	"github.com/qlcchain/go-qlc/test/mock"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestProcess_BlockBasicInfoCheck(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	bs, err := mock.MockBlockChain()
	if err != nil {
		t.Fatal()
	}
	for _, b := range bs {
		processBlock(t, l, b)
	}

	//checkInfo(t, l)
}

func processBlock(t *testing.T, l *Ledger, block types.Block) {
	p, err := l.BlockCheck(block)
	if err != nil {
		t.Fatal()
	}
	if p != Progress {
		t.Fatal(p)
	}
	r := l.BlockProcess(block)
	if r != nil {
		t.Fatal(r)
	}
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
