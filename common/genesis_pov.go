// +build !testnet

package common

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `
{
    "hash":"4e833e2f18f2da77a488a1be13e9db9131d0a2ca521bf23121204c52b84e9af1",
    "previous":"0000000000000000000000000000000000000000000000000000000000000000",
    "merkleRoot":"0000000000000000000000000000000000000000000000000000000000000000",
    "nonce":375666,
    "voteSignature":"608a355ef7cdf4a253d5ee31576dc57aca97d8921c3c69d435f1182013a1c1f3f8b6c09ee39b9f07e89ccf90de85546c4303a282d65a4bd03f1d9a8fd3000000",
    "height":0,
    "timestamp":1558138088,
    "target":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000",
    "coinbase":"qlc_1afnoj8acwikgsazz1ocgakss3hck6htgxcm4wafuit7439izg8kzdu6twou",
    "txNum":0,
    "stateHash":"0000000000000000000000000000000000000000000000000000000000000000",
    "signature":"1f9016e168efa77923ca8571561ca28339a32373f234890db3f4bc09b7f2f1c39bfe6b820d973a6ece25b6cafdaa73fd322c31dda720dbe08ed6e6f268f05909",
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
