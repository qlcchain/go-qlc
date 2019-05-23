// +build testnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/trie"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("70172acee1be7ba2ac3678363a01faeedb18c9dc60306a6085836dc184b42938")
	expectSig, _ := types.NewSignature("ace0c998b630843d557047af1f2a033cd048a6336a92044550a4de7988fffd534a6376ad3815dde91f446dfa18d393ca74af987044db58848a89c9d967e4ec01")

	stateTrie := trie.NewTrie(nil, nil, nil)
	keys, values := GenesisPovStateKVs()
	for i := range keys {
		stateTrie.SetValue(keys[i], values[i])
	}

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
