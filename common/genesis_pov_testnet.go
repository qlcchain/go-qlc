// +build testnet

package common

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
	"math/big"
)

var (
	jsonPovGenesis = `{
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"merkleRoot": "0000000000000000000000000000000000000000000000000000000000000000",
        	"nonce": 375666,
        	"voteSignature": "608a355ef7cdf4a253d5ee31576dc57aca97d8921c3c69d435f1182013a1c1f3f8b6c09ee39b9f07e89ccf90de85546c4303a282d65a4bd03f1d9a8fd3000000",
        	"height": 0,
        	"timestamp": 1556582888,
        	"target": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000",
        	"coinbase": "qlc_1afnoj8acwikgsazz1ocgakss3hck6htgxcm4wafuit7439izg8kzdu6twou",
        	"txNum": 0,
        	"signature": "bce8edf4cf4a03390f903a0937cda43d63d16f7313773326b9086f0182522a6f41c73d140fe24bb9d772ded778c39a70add2ab78afad3b21ca447a4061742803"
        }`

	genesisPovBlockTime   = int64(1556582888)
	genesisPovBlockTarget *big.Int
	genesisPovBlock       types.PovBlock
)

func init() {
	var err error

	var targetSigBuf [types.SignatureSize]byte
	var index int

	targetSigBuf[index] = 0x00
	index++
	targetSigBuf[index] = 0x00
	index++
	targetSigBuf[index] = 0x00
	index++

	for ; index < types.SignatureSize; index++ {
		targetSigBuf[index] = 0xFF
	}

	//test net
	genesisPovBlockTarget = new(big.Int).SetBytes(targetSigBuf[:])
	err = json.Unmarshal([]byte(jsonPovGenesis), &genesisPovBlock)
	if err != nil {
		panic(err)
	}
	genesisPovBlock.Timestamp = genesisPovBlockTime
	genesisPovBlock.Target.FromBigInt(genesisPovBlockTarget)
	genesisPovBlock.Transactions = []*types.PovTransaction{}
	genesisPovBlock.Hash = genesisPovBlock.ComputeHash()
}

func GenesisPovBlock() *types.PovBlock {
	return &genesisPovBlock
}

func IsGenesisPovBlock(block *types.PovBlock) bool {
	h := block.GetHash()
	return h == genesisPovBlock.GetHash()
}
