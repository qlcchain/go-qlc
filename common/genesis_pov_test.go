// +build mainnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/trie"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("288a70e4fd0a738f6e963c7dad253e89a1eda85e57b8dd0079ce11f68246cb75")
	expectSig, _ := types.NewSignature("03fc8dee4ceaca10a6f33ccb5500332482e3a70f51d201f17f9e3fffed2a5fd42e605d3b3ad8ddb81ad1e3e3b86d5d9540fba42fc68fbd71921e739d6815bd05")

	stateTrie := trie.NewTrie(nil, nil, nil)
	stateTrie.SetValue([]byte("qlc"), []byte("qlc is wonderful"))
	expectStateHash := stateTrie.Hash()

	checkStateHash := genesisPovBlock.StateHash
	if *expectStateHash != checkStateHash {
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

	voteSigInt := genesisPovBlock.VoteSignature.ToBigInt()
	targetInt := genesisPovBlock.Target.ToBigInt()
	if voteSigInt.Cmp(targetInt) > 0 {
		t.Fatal("invalid main net genesis pov block target", voteSigInt, targetInt)
	}
}
