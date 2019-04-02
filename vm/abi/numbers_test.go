package abi

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/qlcchain/go-qlc/common/util"
)

func TestNumberTypes(t *testing.T) {
	ubytes := make([]byte, util.WordSize)
	ubytes[util.WordSize-1] = 1

	unsigned := U256(big.NewInt(1))
	if !bytes.Equal(unsigned, ubytes) {
		t.Errorf("expected %x got %x", ubytes, unsigned)
	}
}
