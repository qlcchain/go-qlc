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
            "merkleRoot":"4debe8c3411023be01b86b2a70fa99110a37390695fac1e37bc1b059a2b6c637",
            "timestamp":1569024000,
            "bits":504365040,
            "nonce":653691,
            "hash":"b3459203c8e69d2125932163496de55661a24611b85e2f14a800cd548c040000",
            "height":0
        },
        "auxHdr":null,
        "cbtx":{
            "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
            "txNum":1,
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
            "extra":null,
            "signature":"d2e273d62d7d04f2407af95f5444f25afbd840b9514f9d034c5fff41d7d0daf2fc9ae388349b16c16d0e769975179ef20a938891517e2cbffe98bc1f921e5e02",
            "hash":"4debe8c3411023be01b86b2a70fa99110a37390695fac1e37bc1b059a2b6c637"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"4debe8c3411023be01b86b2a70fa99110a37390695fac1e37bc1b059a2b6c637"
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
