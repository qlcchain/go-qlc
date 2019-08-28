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
            "merkleRoot":"d3462e06cd044a474de6300d019fc100fd8f11ba2d3ea80d1800f4f28f9edb72",
            "timestamp":1569024000,
            "bits":504365040,
            "nonce":173744,
            "hash":"e0405fe21c8e9c0d60515e6bd3e265d740500d1e7243fc98c2d61feaca020000",
            "height":0
        },
        "auxHdr":null,
        "cbtx":{
            "txIns":[
                {
                    "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
                    "txNum":1,
                    "extraData":null
                }
            ],
            "txOuts":[
                {
                    "address":"qlc_1afnoj8acwikgsazz1ocgakss3hck6htgxcm4wafuit7439izg8kzdu6twou",
                    "value":"228310501"
                },
                {
                    "address":"qlc_111111111111111111111111111111111111111111111111111ommygmckp",
                    "value":"57077625"
                }
            ],
            "signature":"1915ad7da3f452db2b570785f05652426e302f393da9b44b317a3ca74158fe3b18d97eb5a236fd429988e861b4a894564451638223cb124ea06e4dbb3a95f707",
            "hash":"d3462e06cd044a474de6300d019fc100fd8f11ba2d3ea80d1800f4f28f9edb72"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"d3462e06cd044a474de6300d019fc100fd8f11ba2d3ea80d1800f4f28f9edb72"
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
