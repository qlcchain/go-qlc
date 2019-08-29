// +build testnet

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
            "merkleRoot":"4f428e57588be436f16fd7249f255cc64bf50036fb08be3e9729ad16db75c4b8",
            "timestamp":1566345600,
            "bits":504365040,
            "nonce":742494,
            "hash":"a01179dc04100cb497859a5c5f3c2399ff0a9d0cbfb2a073084b7f19af0d0000",
            "height":0
        },
        "auxHdr":null,
        "cbtx":{
            "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
            "txNum":1,
            "txOuts":[
                {
                    "address":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
                    "value":"228310501"
                },
                {
                    "address":"qlc_111111111111111111111111111111111111111111111111111ommygmckp",
                    "value":"57077625"
                }
            ],
            "extra":null,
            "signature":"cd158adccc61b28a1e59ab234a3b370a676f7d5f9f642357aaf020cfd3a254075fd6aa35c536963724a66a2ec7d13703a9736024816fabd98dd4b9d077579706",
            "hash":"4f428e57588be436f16fd7249f255cc64bf50036fb08be3e9729ad16db75c4b8"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"4f428e57588be436f16fd7249f255cc64bf50036fb08be3e9729ad16db75c4b8"
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
