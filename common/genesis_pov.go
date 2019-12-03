// +build !testnet

package common

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonPovGenesis = `{
    "header":{
        "basHdr":{
            "version":256,
            "previous":"0000000000000000000000000000000000000000000000000000000000000000",
            "merkleRoot":"0ca279bf861acd7acb33bbc6c44471185ea88f4fad36fb3f973423b73aa9aa07",
            "timestamp":1569024018,
            "bits":506640333,
            "nonce":645962481,
            "hash":"5e66d8b374409cdf3b94dad3e6d3854803077a654ce0f770311b6fa429395626",
            "height":0
        },
        "auxHdr":null,
        "cbtx":{
            "version":1,
            "txIns":[
                {
                    "prevTxHash":"0000000000000000000000000000000000000000000000000000000000000000",
                    "prevTxIdx":4294967295,
                    "extra":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
                    "sequence":4294967295
                }
            ],
            "txOuts":[
                {
                    "value":"342465753",
                    "address":"qlc_1afnoj8acwikgsazz1ocgakss3hck6htgxcm4wafuit7439izg8kzdu6twou"
                },
                {
                    "value":"228310502",
                    "address":"qlc_111111111111111111111111111111111111111111111111111ommygmckp"
                }
            ],
            "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
            "txNum":1,
            "hash":"0ca279bf861acd7acb33bbc6c44471185ea88f4fad36fb3f973423b73aa9aa07"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"0ca279bf861acd7acb33bbc6c44471185ea88f4fad36fb3f973423b73aa9aa07"
            }
        ]
    }
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
