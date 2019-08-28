// +build testnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("7be84840ec28602b4602d242b4e034c641ab09e2b4ccf05707034722bc0c0000")
	expectSig, _ := types.NewSignature("c438e7ce2735f5bbdbe18af6c44f96c4cd8eb0c9b45b12d72d95e390b2ceaea67c21a52ec48b33d355c8ec524f6d746c959fc8f83b9cc518ac2a406bf1073f08")

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
