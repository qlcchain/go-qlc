// +build mainnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/trie"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("13825fa8945001434d9c034a31edc130b04c3a154ef147b30b5814cd236a6b40")
	expectSig, _ := types.NewSignature("9b1bc37bb2e4c94da3479cc3b281222a8666d37962f797f0d43793b1ac7e76e19cb333caf563bb88becd6c9203b28de92f07e8fd922deedd3ce19fb079d32a08")

	stateTrie := trie.NewTrie(nil, nil, nil)
	keys, values := GenesisPovStateKVs()
	for i := range keys {
		stateTrie.SetValue(keys[i], values[i])
	}

	expectStateHash := stateTrie.Hash()

	checkStateHash := genesisPovBlock.StateHash
	if *expectStateHash != checkStateHash {
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

	voteSigInt := genesisPovBlock.VoteSignature.ToBigInt()
	targetInt := genesisPovBlock.Target.ToBigInt()
	if voteSigInt.Cmp(targetInt) > 0 {
		t.Fatal("invalid main net genesis pov block target", voteSigInt, targetInt)
	}
}
