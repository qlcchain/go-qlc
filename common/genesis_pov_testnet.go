// +build testnet

package common

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `
{
    "hash":"e25685cec8eba35fa42fbe62f607858d40f1dd4fc76db1679c8c8d9a10ab5a41",
    "previous":"0000000000000000000000000000000000000000000000000000000000000000",
    "merkleRoot":"0000000000000000000000000000000000000000000000000000000000000000",
    "nonce":399610,
    "voteSignature":"d48a500a9de13d47c88377f29dfb24148e1dbabba8e865674280d5f9734af8573498b954e7614d0c7726348ca46f5e62b5419e649343946e3b1c9e175b000000",
    "height":0,
    "timestamp":1560816000,
    "target":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f000000",
    "coinbase":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
    "txNum":0,
    "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
    "signature":"f4502fd2644d6b1d7fc6314bf04930d955dcd206c3dcfe72f6959bc5df274dec8f82efad1339da8e29f1675eb69493c1d0d9809ab9c051f1b1a23933ea154d0f",
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
