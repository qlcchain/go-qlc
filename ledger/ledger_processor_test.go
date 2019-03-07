package ledger

import (
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestProcess_BlockBasicInfoCheck(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	l.BlockProcess(bc[0])
	for i, b := range bc[1:] {
		t.Log(i)
		if _, err := l.Process(b); err != nil {
			t.Fatal()
		}
	}
	checkInfo(t, l)
}

func checkInfo(t *testing.T, l *Ledger) {
	addrs := make(map[types.Address]int)
	fmt.Println("----blocks----")
	err := l.GetStateBlocks(func(block *types.StateBlock) error {
		fmt.Println(block)
		if _, ok := addrs[block.GetAddress()]; !ok {
			addrs[block.GetAddress()] = 0
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
