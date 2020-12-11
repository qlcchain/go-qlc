// +build testnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("635232d442d0a9e54e38fb3a8f35697cac8360d5cece2d975e4834e646000000")
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

	algoInt := genesisPovBlock.Header.GetAlgoTargetInt()
	normInt := genesisPovBlock.Header.GetNormTargetInt()

	// for nonce := uint32(0); nonce < uint32(0xFFFFFFFF); nonce += uint32(1) {
	// 	genesisPovBlock.Header.BasHdr.Nonce = nonce
	// 	chash := genesisPovBlock.Header.ComputePowHash()
	// 	cint := chash.ToBigInt()
	// 	if cint.Cmp(algoInt) < 0 {
	// 		t.Log("get nonce: ", nonce)
	// 		break
	// 	}
	// }

	powHash := genesisPovBlock.Header.ComputePowHash()
	powInt := powHash.ToBigInt()
	t.Log(PovGenesisPowBits)
	t.Logf("powInt %064x algoInt %064x", powInt, algoInt)
	t.Logf("normInt %064x paramInt %064x", normInt, PovGenesisPowInt)
	if powInt.Cmp(algoInt) > 0 {
		t.Fatal("invalid test net genesis pov pow hash", powHash.String())
	}
}
