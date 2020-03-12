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
	scb := addSmartContract(t, l)
	if err := l.AddSmartContractBlock(scb); err != ErrBlockExists {
		t.Fatal(err)
	}
}

func TestLedger_GetSmartContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	scb := addSmartContract(t, l)
	tests := []struct {
		name       string
		args       types.Hash
		wantReturn *types.SmartContractBlock
		wantErr    error
	}{
		{
			name:       "f1",
			args:       scb.GetHash(),
			wantReturn: scb,
			wantErr:    nil,
		},
		{
			name:       "f2",
			args:       mock.Hash(),
			wantReturn: nil,
			wantErr:    ErrBlockNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := l.GetSmartContractBlock(tt.args)
			if err != tt.wantErr {
				t.Errorf("GetSmartContractBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantReturn != nil {
				if r.GetHash() != tt.wantReturn.GetHash() {
					t.Errorf("GetSmartContractBlock() value = %v, want %v", r, tt.wantReturn)
				}
			}
		})
	}
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
