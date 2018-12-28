package ledger

import (
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/test/mock"
)

func TestProcess_BlockBasicInfoCheck(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	bs, err := mock.MockBlockChain()
	if err != nil {
		t.Fatal()
	}
	for i, b := range bs {
		logger.Info(i)
		if _, err := l.Process(b); err != nil {
			t.Fatal()
		}
	}
	checkInfo(t, l)
}

func checkInfo(t *testing.T, l *Ledger) {
	blocks, _ := l.GetBlocks()
	addrs := make(map[types.Address]int)
	fmt.Println("----blocks----")
	for i, b := range blocks {
		fmt.Println(b)
		if _, ok := addrs[b.GetAddress()]; !ok {
			addrs[b.GetAddress()] = i
		}
	}
	fmt.Println(addrs)
	fmt.Println("----frontiers----")
	fs, _ := l.GetFrontiers()
	for _, f := range fs {
		fmt.Println(f)
	}

	fmt.Println("----account----")
	for k, _ := range addrs {
		ac, _ := l.GetAccountMeta(k)
		fmt.Println("   account", ac.Address)
		for _, token := range ac.Tokens {
			fmt.Println("      token, ", token)
		}
	}

	fmt.Println("----representation----")
	for k, _ := range addrs {
		b, err := l.GetRepresentation(k)
		if err != nil {
			if err == ErrRepresentationNotFound {
			}
		} else {
			fmt.Println(k, b)
		}
	}
}
