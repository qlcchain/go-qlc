// +build testnet

package common

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `
{
    "hash":"f0fce975e2d65396f7d34f36ba64053a0d02b741884070fba2c798dd3a34b336",
    "previous":"0000000000000000000000000000000000000000000000000000000000000000",
    "merkleRoot":"0000000000000000000000000000000000000000000000000000000000000000",
    "nonce":399610,
    "voteSignature":"d48a500a9de13d47c88377f29dfb24148e1dbabba8e865674280d5f9734af8573498b954e7614d0c7726348ca46f5e62b5419e649343946e3b1c9e175b000000",
    "height":0,
    "timestamp":1559801730,
    "target":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f000000",
    "coinbase":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
    "txNum":0,
    "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
    "signature":"c7597a1b4a9a4c781b61bb0890d0c4e559da4c757973df883ac44c82922393bb2688f884756f47d7cd3805cbadd291ff13ed73c92dcc727f33e297086b7fe60a",
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
