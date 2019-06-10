// +build mainnet

package common

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/trie"
)

func TestGenesisPovBlock1(t *testing.T) {
	expectHash, _ := types.NewHash("0ea9f07173118f7e57d8e9eec7edff4ef54a3c12554b0485ac1ce50ce5b6148f")
	expectSig, _ := types.NewSignature("3b827b4cad3671557822f97999d453a76cd9ca01f22e3cc0b8f51fe5fe77dcca909f379537585f3ef733deeb620caeeebfcd86333c37fb14f6773e7045c40d04")

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
