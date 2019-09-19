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
            "version":0,
            "previous":"0000000000000000000000000000000000000000000000000000000000000000",
            "merkleRoot":"beaafa1f9789bacde739050287da41450a05c3c7c7b00786e1707aa97cdbafc0",
            "timestamp":1569024000,
            "bits":504365040,
            "nonce":1289884,
            "hash":"145b6cf60a4b0254f69a245ed89f342966e75acbb6254b5e52567ebbb2040000",
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
                    "address":"qlc_1afnoj8acwikgsazz1ocgakss3hck6htgxcm4wafuit7439izg8kzdu6twou",
                    "value":"456621004"
                },
                {
                    "address":"qlc_111111111111111111111111111111111111111111111111111ommygmckp",
                    "value":"114155251"
                }
            ],
            "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
            "txNum":1,
            "hash":"beaafa1f9789bacde739050287da41450a05c3c7c7b00786e1707aa97cdbafc0"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"beaafa1f9789bacde739050287da41450a05c3c7c7b00786e1707aa97cdbafc0"
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
