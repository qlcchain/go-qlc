/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/qlcchain/go-qlc"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonMintage = `{
        "type": "ContractSend",
        "token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
        "address": "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
        "balance": "0",
        "previous": "0000000000000000000000000000000000000000000000000000000000000000",
        "link": "0000000000000000000000000000000000000000000000000000000000000000",
        "message": "0000000000000000000000000000000000000000000000000000000000000000",
        "data": "RtDOi0XdIXzZ/4n3tkztpIhsxo3enfpHqKQi0WXizm+ag0+tAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA4AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADVKa6ehgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1FMQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANRTEMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
        "quota": 0,
        "timestamp": 1553990401,
        "extra": "0000000000000000000000000000000000000000000000000000000000000000",
        "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        "work": "0000000000000000",
        "signature": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
    }`
	jsonGenesis = `{
        "type": "ContractReward",
        "token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
        "address": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        "balance": "60000000000000000",
        "previous": "0000000000000000000000000000000000000000000000000000000000000000",
        "link": "bf1cb34e79f8739367ad7de4a16c87c0e72ea483521fec0f0ddf7b5e90d03abd",
        "message": "0000000000000000000000000000000000000000000000000000000000000000",
        "data": "Rd0hfNn/ife2TO2kiGzGjd6d+keopCLRZeLOb5qDT60AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANUprp6GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACGgb9SU8ZGcvxUyT07W5og0olly4+AunBGDtP5nLVHvVAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADUUxDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1FMQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        "quota": 0,
        "timestamp": 1553990410,
        "extra": "0000000000000000000000000000000000000000000000000000000000000000",
        "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        "work": "000000000054b46a",
        "signature": "3420612d28da9e5990bec8aa53d5dab89c9bcc70d020ce1dc8c74c844274e6eaf3eb7bba24bca95e566b036b17c9f01936835042dde70d0bd8f49659f83c0306"
    }`

	testJsonMintage = `{
        "type": "ContractSend",
        "token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
        "address": "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
        "balance": "0",
        "previous": "0000000000000000000000000000000000000000000000000000000000000000",
        "link": "0000000000000000000000000000000000000000000000000000000000000000",
        "message": "0000000000000000000000000000000000000000000000000000000000000000",
        "data": "RtDOi6fo+jDAY+lqSJpHvEOQlQW9hnNdpKEJ3KKL6TYRioWCAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA4AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADVKa6ehgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1FMQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANRTEMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
        "quota": 0,
        "timestamp": 1553990401,
        "extra": "0000000000000000000000000000000000000000000000000000000000000000",
        "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        "work": "0000000000000000",
        "signature": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
    }`

	testJsonGenesis = `{
        "type": "ContractReward",
        "token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
        "address": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        "balance": "60000000000000000",
        "previous": "0000000000000000000000000000000000000000000000000000000000000000",
        "link": "f1042ef1472b02ee61048cf4eba923b81d0a64876665888b44bb2c46534f9655",
        "message": "0000000000000000000000000000000000000000000000000000000000000000",
        "quota": 0,
        "data": "p+j6MMBj6WpImke8Q5CVBb2Gc12koQncoovpNhGKhYIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANUprp6GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACL+GyD+0v7n0m5uPpZPIz0EoyeIXIMSHVltS/GZAqejzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADUUxDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1FMQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        "timestamp": 1553990410,
        "extra": "0000000000000000000000000000000000000000000000000000000000000000",
        "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        "work": "000000000048f5b9",
        "signature": "e89c3f7e28a51b5fed0af6cae206a82bcae436576c08db5e056a168e4337aa803d8b38a6d0438f377e57ece1d83cbf18a4cab9bc176bf640d1c62452e5c0f003"
    }`

	units = map[string]*big.Int{
		"qlc":  big.NewInt(1),
		"Kqlc": big.NewInt(1e5),
		"QLC":  big.NewInt(1e8),
		"MQLC": big.NewInt(1e11),
	}

	//main net
	chainToken, _       = types.NewHash("45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad")
	genesisAddress, _   = types.HexToAddress("qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44")
	genesisMintageBlock types.StateBlock
	genesisMintageHash  types.Hash
	genesisBlock        types.StateBlock
	genesisBlockHash    types.Hash

	//test net
	testChainToken, _       = types.NewHash("a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582")
	testGenesisAddress, _   = types.HexToAddress("qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4")
	testGenesisMintageBlock types.StateBlock
	testGenesisMintageHash  types.Hash
	testGenesisBlock        types.StateBlock
	testGenesisBlockHash    types.Hash
)

func init() {
	_ = json.Unmarshal([]byte(jsonMintage), &genesisMintageBlock)
	_ = json.Unmarshal([]byte(jsonGenesis), &genesisBlock)
	genesisMintageHash = genesisMintageBlock.GetHash()
	genesisBlockHash = genesisBlock.GetHash()

	_ = json.Unmarshal([]byte(testJsonMintage), &testGenesisMintageBlock)
	_ = json.Unmarshal([]byte(testJsonGenesis), &testGenesisBlock)
	testGenesisMintageHash = testGenesisMintageBlock.GetHash()
	testGenesisBlockHash = testGenesisBlock.GetHash()
}

func GenesisAddress() types.Address {
	if goqlc.MAINNET {
		return genesisAddress
	}
	return testGenesisAddress
}

func ChainToken() types.Hash {
	if goqlc.MAINNET {
		return chainToken
	}
	return testChainToken
}

func GenesisMintageBlock() types.StateBlock {
	if goqlc.MAINNET {
		return genesisMintageBlock
	}
	return testGenesisMintageBlock
}

func GenesisMintageHash() types.Hash {
	if goqlc.MAINNET {
		return genesisMintageHash
	}
	return testGenesisMintageHash
}

func GenesisBlock() types.StateBlock {
	if goqlc.MAINNET {
		return genesisBlock
	}
	return testGenesisBlock
}

func GenesisBlockHash() types.Hash {
	if goqlc.MAINNET {
		return genesisBlockHash
	}
	return testGenesisBlockHash
}

// IsGenesis check block is chain token genesis
func IsGenesisBlock(block *types.StateBlock) bool {
	h := block.GetHash()
	if goqlc.MAINNET {
		return h == genesisMintageHash || h == genesisBlockHash
	}

	return h == testGenesisMintageHash || h == testGenesisBlockHash
}

func BalanceToRaw(b types.Balance, unit string) (types.Balance, error) {
	if v, ok := units[unit]; ok {
		//v = v.Div(v, units["raw"])
		return types.Balance{Int: new(big.Int).Mul(b.Int, v)}, nil
	}
	return b, fmt.Errorf("invalid unit %s", unit)
}

func RawToBalance(b types.Balance, unit string) (types.Balance, error) {
	if v, ok := units[unit]; ok {
		//v = v.Div(v, units["raw"])
		return types.Balance{Int: new(big.Int).Div(b.Int, v)}, nil
	}
	return b, fmt.Errorf("invalid unit %s", unit)
}
