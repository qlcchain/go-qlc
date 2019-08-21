// +build mainnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("778dd8f1bdedc851e8c4177474aaa9e4fbf138a9be8611cd00cc50e80d0c0000")
	expectSig, _ := types.NewSignature("ddb47aa6bcc8bdb71e04434f96fa29a8ae5e307970185391df78e203d884d17565028a25222a64b4425f7ea53d9aa418e73d599b0f1cbcc4b1837300fb2a1809")

	expectStateHash, _ := types.NewHash("1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49")

	checkStateHash := genesisPovBlock.GetStateHash()
	if expectStateHash != checkStateHash {
		t.Log(util.ToString(genesisPovBlock))
		t.Fatal("invalid main net genesis pov state hash", checkStateHash.String(), expectStateHash.String())
	}

	checkHash := genesisPovBlock.ComputeHash()
	if checkHash != expectHash {
		t.Log(util.ToString(genesisPovBlock))
		t.Fatal("invalid main net genesis pov block hash", checkHash.String(), expectHash.String())
	}

	checkSig := genesisPovBlock.GetSignature()
	if checkSig != expectSig {
		t.Log(util.ToString(genesisPovBlock))
		t.Fatal("invalid main net genesis pov block signature", checkSig.String(), expectSig.String())
	}
}
