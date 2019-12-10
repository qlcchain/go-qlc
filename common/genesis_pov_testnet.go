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
            "merkleRoot":"199d4181fe00e8eb8a181307da78bca0d494296d1d9b93702c69fe1d50c58312",
            "timestamp":1575590403,
            "bits":486604799,
            "nonce":1903979056,
            "hash":"091c23777ec5605488f71481ec56e228d05056e1451ce8ed33bd3b1a00000000",
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
                    "address":"qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj"
                },
                {
                    "value":"228310502",
                    "address":"qlc_111111111111111111111111111111111111111111111111111ommygmckp"
                }
            ],
            "stateHash":"1e78dcddbe569968e758251ada684d313104ca72285285e21cc381770fd3ee49",
            "txNum":1,
            "hash":"199d4181fe00e8eb8a181307da78bca0d494296d1d9b93702c69fe1d50c58312"
        }
    },
    "body":{
        "txs":[
            {
                "hash":"199d4181fe00e8eb8a181307da78bca0d494296d1d9b93702c69fe1d50c58312"
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
