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
            "merkleRoot":"be2bc3eec8cdc75a53dee7992501d7bb97cbd8894f8071dbc897161b4b867190",
            "timestamp":1566345600,
            "bits":504365040,
            "nonce":92723,
            "hash":"acf6860543409cbbf09dbf0aa3c9b1100ab7aa35f66b648501c734abe4080000",
            "height":0
        },
        "auxHdr":null,
        "cbtx":{
            "txNum":1,
            "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
            "reward":"285388127",
            "coinBase":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
            "extra":null,
            "signature":"1ea87b68ebf2bc596a8b8f19df000ac9850f9a7bb6ff5f37ee3944988c60e8cc0b058d0b2eac0839f11e4e1aeee26474ee23fcc1376b359a8bef88def002000d",
            "hash":"be2bc3eec8cdc75a53dee7992501d7bb97cbd8894f8071dbc897161b4b867190"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"be2bc3eec8cdc75a53dee7992501d7bb97cbd8894f8071dbc897161b4b867190"
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
