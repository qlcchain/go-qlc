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
            "merkleRoot":"203fbe89d109278fd220b1022f891067ea6c027297b4022b7893734c3516f070",
            "timestamp":1569024000,
            "bits":504365040,
            "nonce":2729537,
            "hash":"163448f1f6e63eaa6bf3c2b89fa843bf193fd875ee2c02f17959c46307090000",
            "height":0
        },
        "auxHdr":null,
        "cbtx":{
            "version":0,
            "txIns":[
                {
                    "prevTxHash":"0000000000000000000000000000000000000000000000000000000000000000",
                    "prevTxIdx":4294967295,
                    "extra":"Hnjc3b5WmWjnWCUa2mhNMTEEynIoUoXiHMOBdw/T7kk=",
                    "sequence":4294967295
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
            "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
            "txNum":1,
            "hash":"203fbe89d109278fd220b1022f891067ea6c027297b4022b7893734c3516f070"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"203fbe89d109278fd220b1022f891067ea6c027297b4022b7893734c3516f070"
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
