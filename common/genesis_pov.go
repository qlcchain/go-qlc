// +build !testnet

package common

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `
{
    "header":{
        "basHdr":{
            "version":0,
            "previous":"0000000000000000000000000000000000000000000000000000000000000000",
            "merkleRoot":"4cfbbb9e2ba385de65389588f9801d8e5a3ea29ee41ea0291fd30376c4eec98a",
            "timestamp":1569024000,
            "bits":504365040,
            "nonce":479752,
            "hash":"778dd8f1bdedc851e8c4177474aaa9e4fbf138a9be8611cd00cc50e80d0c0000",
            "height":0
        },
        "auxHdr":null,
        "cbtx":{
            "txNum":1,
            "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
            "reward":"285388127",
            "coinBase":"qlc_1afnoj8acwikgsazz1ocgakss3hck6htgxcm4wafuit7439izg8kzdu6twou",
            "extra":null,
            "signature":"ddb47aa6bcc8bdb71e04434f96fa29a8ae5e307970185391df78e203d884d17565028a25222a64b4425f7ea53d9aa418e73d599b0f1cbcc4b1837300fb2a1809",
            "Hash":"4cfbbb9e2ba385de65389588f9801d8e5a3ea29ee41ea0291fd30376c4eec98a"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"4cfbbb9e2ba385de65389588f9801d8e5a3ea29ee41ea0291fd30376c4eec98a"
            }
        ]
    }
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
