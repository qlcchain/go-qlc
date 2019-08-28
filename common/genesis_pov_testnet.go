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
            "merkleRoot":"160f22138bbdbcadb4e84c511e8900683b8964fbe9259f56646ad83f42ab7626",
            "timestamp":1566345600,
            "bits":504365040,
            "nonce":217468,
            "hash":"7be84840ec28602b4602d242b4e034c641ab09e2b4ccf05707034722bc0c0000",
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
                    "address":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
                    "value":"228310501"
                },
                {
                    "address":"qlc_111111111111111111111111111111111111111111111111111ommygmckp",
                    "value":"57077625"
                }
            ],
            "signature":"c438e7ce2735f5bbdbe18af6c44f96c4cd8eb0c9b45b12d72d95e390b2ceaea67c21a52ec48b33d355c8ec524f6d746c959fc8f83b9cc518ac2a406bf1073f08",
            "hash":"160f22138bbdbcadb4e84c511e8900683b8964fbe9259f56646ad83f42ab7626"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"160f22138bbdbcadb4e84c511e8900683b8964fbe9259f56646ad83f42ab7626"
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
