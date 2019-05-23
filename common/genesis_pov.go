// +build !testnet

package common

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `
{
    "hash":"13825fa8945001434d9c034a31edc130b04c3a154ef147b30b5814cd236a6b40",
    "previous":"0000000000000000000000000000000000000000000000000000000000000000",
    "merkleRoot":"0000000000000000000000000000000000000000000000000000000000000000",
    "nonce":261629,
    "voteSignature":"cdf4643477c69fc6fca32fc6e0c10963bb9adc259020d943216e61e558ee2c0c27ab555ba79944b43e871acf22660f9a69388d4a0bb9a0490bcd01fa45000000",
    "height":0,
    "timestamp":1561162088,
    "target":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000",
    "coinbase":"qlc_1afnoj8acwikgsazz1ocgakss3hck6htgxcm4wafuit7439izg8kzdu6twou",
    "txNum":0,
    "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
    "signature":"9b1bc37bb2e4c94da3479cc3b281222a8666d37962f797f0d43793b1ac7e76e19cb333caf563bb88becd6c9203b28de92f07e8fd922deedd3ce19fb079d32a08",
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
