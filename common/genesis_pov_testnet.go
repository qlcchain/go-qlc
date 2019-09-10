// +build testnet

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
            "merkleRoot":"c7b6e1929787ba843e494b2e3c9c78b5f8f467d01fb64e24470215cafd04f1ba",
            "timestamp":1566345600,
            "bits":504365040,
            "nonce":1944040,
            "hash":"71e5dd2668d7a1e991c8645ffc65d54cc76eaeea8c16c735d111af69360d0000",
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
                    "address":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
                    "value":"228310501"
                },
                {
                    "address":"qlc_111111111111111111111111111111111111111111111111111ommygmckp",
                    "value":"57077625"
                }
            ],
            "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
            "txNum":1,
            "hash":"c7b6e1929787ba843e494b2e3c9c78b5f8f467d01fb64e24470215cafd04f1ba"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"c7b6e1929787ba843e494b2e3c9c78b5f8f467d01fb64e24470215cafd04f1ba"
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
