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
            "merkleRoot":"7d3a1a8178b8865a829bd9cbd9b3be378d8127226b102458cadcbb6c5b16cb87",
            "timestamp":1566345600,
            "bits":504365040,
            "nonce":2299442235,
            "hash":"461b360695d1eb8c468766bbeae6114daed22e7c88cf752bdcdc959e5f0d0000",
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
                    "value":"456621004",
                    "address":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj"
                },
                {
                    "value":"114155251",
                    "address":"qlc_111111111111111111111111111111111111111111111111111ommygmckp"
                }
            ],
            "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
            "txNum":1,
            "hash":"7d3a1a8178b8865a829bd9cbd9b3be378d8127226b102458cadcbb6c5b16cb87"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"7d3a1a8178b8865a829bd9cbd9b3be378d8127226b102458cadcbb6c5b16cb87"
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
