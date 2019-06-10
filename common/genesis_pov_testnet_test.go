// +build testnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/trie"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("f0fce975e2d65396f7d34f36ba64053a0d02b741884070fba2c798dd3a34b336")
	expectSig, _ := types.NewSignature("c7597a1b4a9a4c781b61bb0890d0c4e559da4c757973df883ac44c82922393bb2688f884756f47d7cd3805cbadd291ff13ed73c92dcc727f33e297086b7fe60a")

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
