// +build !testnet

package common

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `
{
    "hash":"288a70e4fd0a738f6e963c7dad253e89a1eda85e57b8dd0079ce11f68246cb75",
    "previous":"0000000000000000000000000000000000000000000000000000000000000000",
    "merkleRoot":"0000000000000000000000000000000000000000000000000000000000000000",
    "nonce":375666,
    "voteSignature":"608a355ef7cdf4a253d5ee31576dc57aca97d8921c3c69d435f1182013a1c1f3f8b6c09ee39b9f07e89ccf90de85546c4303a282d65a4bd03f1d9a8fd3000000",
    "height":0,
    "timestamp":1558138088,
    "target":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000",
    "coinbase":"qlc_1afnoj8acwikgsazz1ocgakss3hck6htgxcm4wafuit7439izg8kzdu6twou",
    "txNum":0,
    "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
    "signature":"03fc8dee4ceaca10a6f33ccb5500332482e3a70f51d201f17f9e3fffed2a5fd42e605d3b3ad8ddb81ad1e3e3b86d5d9540fba42fc68fbd71921e739d6815bd05",
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

	var miners []*genesisMinerCfg
	for _, miner := range miners {
		addr, _ := types.HexToAddress(miner.Coinbase)
		as := types.NewPovAccountState()
		as.RepState = types.NewPovRepState()
		as.RepState.Vote = types.StringToBalance(miner.Vote)
		as.RepState.Total = as.RepState.CalcTotal()
		asBytes, _ := as.Serialize()
		keys = append(keys, addr.Bytes())
		values = append(values, asBytes)
	}

	return keys, values
}
