// +build testnet

package common

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `
{
    "hash":"a1c619a4781884413af833aceeed2d2c849dc10788936505daff82d0191eb878",
    "previous":"0000000000000000000000000000000000000000000000000000000000000000",
    "merkleRoot":"0000000000000000000000000000000000000000000000000000000000000000",
    "nonce":1481494,
    "voteSignature":"105a4609daecb5fd63230eff18b57ec200351f9a28d5cd29c9bbca3e405a917760da0f867ba9922f2008ec04a359b52390ca7521650f3e172698e57ec1000000",
    "height":0,
    "timestamp":1555891688,
    "target":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000",
    "coinbase":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
    "txNum":0,
    "stateHash":"a50efe4473fcad4c1008e4c567b4ef3dc516818977231f125409dbfa5181e4bb",
    "signature":"f98e79158c18ed76c0cea1f7543dbc09af72f24e3460f878c8aee8bcf589352a3373c576748f869d6cc3bab03449f6c6727ccbbb4af1e2d5c48d57366fa42902",
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
	miners = append(miners, &genesisMinerCfg{
		Coinbase: "qlc_1szuejgo9nxdre1uwpsxni4fg7p8kx7micbsdtpnchmc3cfk4wt1i37uncmy",
		Vote:     "500000",
	})
	miners = append(miners, &genesisMinerCfg{
		Coinbase: "qlc_1h1oyd1h98cigxe9u1xkf7h973cartstf44djpx54ea7ize7bhg5caz6cm7b",
		Vote:     "500000",
	})
	miners = append(miners, &genesisMinerCfg{
		Coinbase: "qlc_1ojc6yxwfbdijokaustj16jyppm4ixtnpkcod6fiskj8zqgtgxph9nnfkthu",
		Vote:     "500000",
	})
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
