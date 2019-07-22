// +build !testnet

package common

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `{
    "hash":"a70c5e5d46e350480ece7b4f651fa98a33c0198de0ef17c05546bf8351f9b032",
    "previous":"0000000000000000000000000000000000000000000000000000000000000000",
    "merkleRoot":"0000000000000000000000000000000000000000000000000000000000000000",
    "nonce":261629,
    "voteSignature":"cdf4643477c69fc6fca32fc6e0c10963bb9adc259020d943216e61e558ee2c0c27ab555ba79944b43e871acf22660f9a69388d4a0bb9a0490bcd01fa45000000",
    "height":0,
    "timestamp":1565913600,
    "target":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f000000",
    "coinbase":"qlc_1afnoj8acwikgsazz1ocgakss3hck6htgxcm4wafuit7439izg8kzdu6twou",
    "txNum":0,
    "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
    "signature":"58bf9c2622c1be3d1de69fbc4f7764186d8cd15083474aa361addacf203f03be18a783709452607a9ef330e879e7a2f31a65c47efc0484e2a78b989c3810ab04",
    "transactions":[]
}`

	genesisPovBlock types.PovBlock
)

func init() {
	err := json.Unmarshal([]byte(jsonPovGenesis), &genesisPovBlock)
	if err != nil {
		panic(err)
	}
}

func GenesisPovBlock() types.PovBlock {
	return genesisPovBlock
}

func IsGenesisPovBlock(block *types.PovBlock) bool {
	h := block.GetHash()
	return h == genesisPovBlock.GetHash()
}

func GenesisPovStateKVs() (keys [][]byte, values [][]byte) {
	keys = append(keys, []byte("qlc"))
	values = append(values, []byte("create something wonderful"))

	return keys, values
}
