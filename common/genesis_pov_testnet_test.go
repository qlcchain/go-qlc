// +build testnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("acf6860543409cbbf09dbf0aa3c9b1100ab7aa35f66b648501c734abe4080000")
	expectSig, _ := types.NewSignature("1ea87b68ebf2bc596a8b8f19df000ac9850f9a7bb6ff5f37ee3944988c60e8cc0b058d0b2eac0839f11e4e1aeee26474ee23fcc1376b359a8bef88def002000d")

	expectStateHash, _ := types.NewHash("1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49")

	checkStateHash := genesisPovBlock.GetStateHash()
	if expectStateHash != checkStateHash {
		t.Log(util.ToString(genesisPovBlock))
		t.Fatal("invalid test net genesis pov state hash", checkStateHash.String(), expectStateHash.String())
	}

	checkHash := genesisPovBlock.ComputeHash()
	if checkHash != expectHash {
		t.Log(util.ToString(genesisPovBlock))
		t.Fatal("invalid test net genesis pov block hash", checkHash.String(), expectHash.String())
	}

	checkSig := genesisPovBlock.GetSignature()
	if checkSig != expectSig {
		t.Log(util.ToString(genesisPovBlock))
		t.Fatal("invalid test net genesis pov block signature", checkSig.String(), expectSig.String())
	}
}
