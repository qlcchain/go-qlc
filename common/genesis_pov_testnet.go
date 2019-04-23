// +build testnet

package common

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `
{
    "hash":"93adfc01a9c1a3ad08270166741b82dccfa0a9251f9a84dfe6ca16225a113516",
    "previous":"0000000000000000000000000000000000000000000000000000000000000000",
    "merkleRoot":"0000000000000000000000000000000000000000000000000000000000000000",
    "nonce":1481494,
    "voteSignature":"105a4609daecb5fd63230eff18b57ec200351f9a28d5cd29c9bbca3e405a917760da0f867ba9922f2008ec04a359b52390ca7521650f3e172698e57ec1000000",
    "height":0,
    "timestamp":1555891688,
    "target":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000",
    "coinbase":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
    "txNum":0,
    "stateHash":"0000000000000000000000000000000000000000000000000000000000000000",
    "signature":"368e406e9fd1684010814eb3244f7e9bfae45f80ee38ab6914176257093380795a9aaa9b0f1c95109c48e7ffa9b48a47d32dc597148c486a5473dedcb5195108",
    "transactions":[]
}
`

	genesisPovBlock types.PovBlock
)

func init() {
	err := json.Unmarshal([]byte(jsonPovGenesis), &genesisPovBlock)
	if err != nil {
		panic(err)
	}
}

func GenesisPovBlock() *types.PovBlock {
	return &genesisPovBlock
}

func IsGenesisPovBlock(block *types.PovBlock) bool {
	h := block.GetHash()
	return h == genesisPovBlock.GetHash()
}
