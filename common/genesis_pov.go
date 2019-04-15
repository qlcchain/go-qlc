package common

import (
	"encoding/json"
	goqlc "github.com/qlcchain/go-qlc"
	"github.com/qlcchain/go-qlc/common/types"
	"math/big"
)

var (
	jsonPovGenesis = `{
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"merkleRoot": "0000000000000000000000000000000000000000000000000000000000000000",
        	"nonce": 1481494,
        	"voteSignature": "105a4609daecb5fd63230eff18b57ec200351f9a28d5cd29c9bbca3e405a917760da0f867ba9922f2008ec04a359b52390ca7521650f3e172698e57ec1000000",
        	"height": 0,
        	"timestamp": 1559261288,
        	"target": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000",
        	"coinbase": "qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
        	"txNum": 0,
        	"signature": "f30964377bced6a066f6c4289f2197ac62ac7ba2bd90f91854ab3495552083b2850ac3ff142fcc1442b4e1e3975d4727bd8bcb94127f587da044ad7649bb6705"
        }`

	testJsonPovGenesis = `{
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

	genesisPovBlockTime   = int64(1559261288)
	genesisPovBlockTarget *big.Int
	genesisPovBlock       types.PovBlock

	testGenesisPovBlockTime   = int64(1556582888)
	testGenesisPovBlockTarget *big.Int
	testGenesisPovBlock       types.PovBlock
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

	//main net
	genesisPovBlockTarget = new(big.Int).SetBytes(targetSigBuf[:])
	err = json.Unmarshal([]byte(jsonPovGenesis), &genesisPovBlock)
	if err != nil {
		panic(err)
	}
	genesisPovBlock.Timestamp = genesisPovBlockTime
	genesisPovBlock.Target.FromBigInt(genesisPovBlockTarget)
	genesisPovBlock.Transactions = []*types.PovTransaction{}
	genesisPovBlock.Hash = genesisPovBlock.ComputeHash()

	//test net
	testGenesisPovBlockTarget = new(big.Int).SetBytes(targetSigBuf[:])
	err = json.Unmarshal([]byte(testJsonPovGenesis), &testGenesisPovBlock)
	if err != nil {
		panic(err)
	}
	testGenesisPovBlock.Timestamp = testGenesisPovBlockTime
	testGenesisPovBlock.Target.FromBigInt(testGenesisPovBlockTarget)
	testGenesisPovBlock.Transactions = []*types.PovTransaction{}
	testGenesisPovBlock.Hash = testGenesisPovBlock.ComputeHash()
}

func GenesisPovBlock() *types.PovBlock {
	if goqlc.MAINNET {
		return &genesisPovBlock
	}
	return &testGenesisPovBlock
}

func IsGenesisPovBlock(block *types.PovBlock) bool {
	h := block.GetHash()
	if goqlc.MAINNET {
		return h == genesisPovBlock.GetHash()
	}

	return h == testGenesisPovBlock.GetHash()
}
