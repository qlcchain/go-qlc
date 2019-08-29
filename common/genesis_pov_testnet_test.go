// +build testnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("a01179dc04100cb497859a5c5f3c2399ff0a9d0cbfb2a073084b7f19af0d0000")
	expectSig, _ := types.NewSignature("cd158adccc61b28a1e59ab234a3b370a676f7d5f9f642357aaf020cfd3a254075fd6aa35c536963724a66a2ec7d13703a9736024816fabd98dd4b9d077579706")

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
