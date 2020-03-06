package ledger

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func addSmartContract(t *testing.T, l *Ledger) *types.SmartContractBlock {
	scb := &types.SmartContractBlock{
		Address: mock.Address(),
		Owner:   mock.Address(),
	}
	if err := l.AddSmartContractBlock(scb); err != nil {
		t.Fatal(err)
	}
	return scb
}

func TestLedger_AddSmartContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addSmartContract(t, l)
}

func TestLedger_GetSmartContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	scb := addSmartContract(t, l)
	value, err := l.GetSmartContractBlock(scb.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	if scb.GetHash() != value.GetHash() {
		t.Fatal(err)
	}
	t.Log(value)
}

func TestLedger_GetSmartContracts(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addSmartContract(t, l)
	addSmartContract(t, l)
	err := l.GetSmartContractBlocks(func(block *types.SmartContractBlock) error {
		t.Log(block)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_CountSmartContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addSmartContract(t, l)
	addSmartContract(t, l)
	num, err := l.CountSmartContractBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if num != 2 {
		t.Fatal("error count")
	}
	t.Log("account,", num)
}

func TestLedger_HasSmartContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	scb := addSmartContract(t, l)
	b, err := l.HasSmartContractBlock(scb.GetHash())
	if err != nil || !b {
		t.Fatal(err)
	}
}
