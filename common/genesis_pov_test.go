// +build mainnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("f9b322a0b87057122ef30c780245614c90be780b68f0f1ae355b0b0e00000000")

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

	powHash := genesisPovBlock.Header.ComputePowHash()
	powInt := powHash.ToBigInt()
	algoInt := genesisPovBlock.Header.GetAlgoTargetInt()
	normInt := genesisPovBlock.Header.GetNormTargetInt()
	t.Logf("powInt %064x algoInt %064x", powInt, algoInt)
	t.Logf("normInt %064x paramInt %064x", normInt, PovGenesisPowInt)
	if powInt.Cmp(algoInt) > 0 {
		t.Fatal("invalid main net genesis pov pow hash", powHash.String())
	}
}
