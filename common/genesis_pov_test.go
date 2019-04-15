/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"encoding/hex"
	"fmt"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func makeGenesisPovBlock(t *testing.T, genesisBlock *types.PovBlock) {
	addr, priKey, _ := types.GenerateAddress()
	genesisBlock.Coinbase = addr

	genesisTarget := genesisBlock.Target.ToBigInt()

	fmt.Println("address", addr, "priKey", hex.EncodeToString(priKey[:]))

	fmt.Println("genesisTarget", genesisTarget.Text(16), "BitLen", genesisTarget.BitLen())

	for i := uint32(0); i <= ^uint32(0); i++ {
		genesisBlock.Nonce = uint64(i)
		voteHash := genesisBlock.ComputeVoteHash()
		voteSigBytes := ed25519.Sign(priKey, voteHash[:])
		copy(genesisBlock.VoteSignature[:], voteSigBytes)

		chkTarget := genesisBlock.VoteSignature.ToBigInt()

		//fmt.Println("voteHash", voteHash, "voteSig", genesisPovBlock.VoteSignature, "diff", target.Text(16), "bitLen", target.BitLen())

		if chkTarget.Cmp(genesisTarget) > 0 {
			continue
		}

		break
	}

	genesisBlock.Hash = genesisBlock.ComputeHash()
	blockSigBytes := ed25519.Sign(priKey, genesisBlock.Hash[:])
	copy(genesisBlock.Signature[:], blockSigBytes)

	fmt.Println("genesisBlock", util.ToString(genesisBlock))
}

// TestMakeGenesisPovBlocks just for make genesis pov blocks
func _TestMakeGenesisPovBlocks(t *testing.T) {
	fmt.Println("==== make main net genesisPovBlock ====")
	makeGenesisPovBlock(t, &genesisPovBlock)

	fmt.Println("==== make test net GenesisPovBlock ====")
	makeGenesisPovBlock(t, &testGenesisPovBlock)
}

func TestGenesisPovBlock1(t *testing.T) {
	h, _ := types.NewHash("87db7ff44ff7a2acb4332a7f4fb9b429aa8128bb47bc4edc5f71a53b218188ef")

	h2 := genesisPovBlock.ComputeHash()
	if h2 != h {
		t.Log(util.ToString(genesisPovBlock))
		t.Fatal("invalid main net genesis pov block hash", h2.String(), h.String())
	}

	voteSigInt := genesisPovBlock.VoteSignature.ToBigInt()
	targetInt := genesisPovBlock.Target.ToBigInt()
	if voteSigInt.Cmp(targetInt) > 0 {
		t.Fatal("invalid main net genesis pov block target", voteSigInt, targetInt)
	}
}

func TestGenesisPovBlock2(t *testing.T) {
	h, _ := types.NewHash("e1b24548f5a692c39b8ee024143ec531b8c06c121da122438eb4a3315441679a")

	h2 := testGenesisPovBlock.ComputeHash()
	if h2 != h {
		t.Log(util.ToString(testGenesisPovBlock))
		t.Fatal("invalid test net genesis pov block", h2.String(), h.String())
	}

	voteSigInt := testGenesisPovBlock.VoteSignature.ToBigInt()
	targetInt := testGenesisPovBlock.Target.ToBigInt()
	if voteSigInt.Cmp(targetInt) > 0 {
		t.Fatal("invalid main net genesis pov block target", voteSigInt, targetInt)
	}
}
