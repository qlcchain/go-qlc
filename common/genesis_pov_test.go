// +build mainnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("b3459203c8e69d2125932163496de55661a24611b85e2f14a800cd548c040000")
	expectSig, _ := types.NewSignature("d2e273d62d7d04f2407af95f5444f25afbd840b9514f9d034c5fff41d7d0daf2fc9ae388349b16c16d0e769975179ef20a938891517e2cbffe98bc1f921e5e02")

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
