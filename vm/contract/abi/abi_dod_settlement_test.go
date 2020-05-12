package abi

import (
	"strings"
	"testing"

	"github.com/qlcchain/go-qlc/vm/abi"
)

func TestDoDSettlementABI(t *testing.T) {
	_, err := abi.JSONToABIContract(strings.NewReader(JsonDoDSettlement))
	if err != nil {
		t.Fatal(err)
	}
}
