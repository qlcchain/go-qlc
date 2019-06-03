// +build testnet

package common

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `
{
    "hash":"1f60e2bd07072901f17304f8b9fb85f008d8c1d6dc72407e78c9c9c4ffc63334",
    "previous":"0000000000000000000000000000000000000000000000000000000000000000",
    "merkleRoot":"0000000000000000000000000000000000000000000000000000000000000000",
    "nonce":399610,
    "voteSignature":"d48a500a9de13d47c88377f29dfb24148e1dbabba8e865674280d5f9734af8573498b954e7614d0c7726348ca46f5e62b5419e649343946e3b1c9e175b000000",
    "height":0,
    "timestamp":1559024130,
    "target":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000",
    "coinbase":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
    "txNum":0,
    "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
    "signature":"353bc16f6de796e630c130f35e2c379c307a09152e1c69a30c498721876ccc8a9c1fd1b176454962bb0fc9ab96967f21def565ef567667e20c458ee614dcbf07",
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
