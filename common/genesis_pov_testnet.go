// +build testnet

package common

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `
{
    "hash":"70172acee1be7ba2ac3678363a01faeedb18c9dc60306a6085836dc184b42938",
    "previous":"0000000000000000000000000000000000000000000000000000000000000000",
    "merkleRoot":"0000000000000000000000000000000000000000000000000000000000000000",
    "nonce":399610,
    "voteSignature":"d48a500a9de13d47c88377f29dfb24148e1dbabba8e865674280d5f9734af8573498b954e7614d0c7726348ca46f5e62b5419e649343946e3b1c9e175b000000",
    "height":0,
    "timestamp":1558483688,
    "target":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000",
    "coinbase":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
    "txNum":0,
    "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
    "signature":"ace0c998b630843d557047af1f2a033cd048a6336a92044550a4de7988fffd534a6376ad3815dde91f446dfa18d393ca74af987044db58848a89c9d967e4ec01",
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

type genesisMinerCfg struct {
	Coinbase string
	Vote     string
}

func GenesisPovStateKVs() (keys [][]byte, values [][]byte) {
	keys = append(keys, []byte("qlc"))
	values = append(values, []byte("create something wonderful"))

	return keys, values
}
