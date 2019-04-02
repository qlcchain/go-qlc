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

	goqlc "github.com/qlcchain/go-qlc"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	jsonMintage = `{
        	"type": "ContractSend",
        	"token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
        	"address": "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
        	"balance": "0",
        	"vote": "0",
        	"network": "0",
        	"storage": "0",
        	"oracle": "0",
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"link": "681bf5253c64672fc54c93d3b5b9a20d28965cb8f80ba70460ed3f99cb547bd5",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "6TrdxEXdIXzZ/4n3tkztpIhsxo3enfpHqKQi0WXizm+ag0+tAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADVKa6ehgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAhoG/UlPGRnL8VMk9O1uaINKJZcuPgLpwRg7T+Zy1R71QAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADUUxDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1FMQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        	"povHeight": 0,
        	"timestamp": 1553990401,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        	"work": "000000000048f5b9",
        	"signature": "0d09a049ae11a005648b9ea9f4d7f2ce2fed985a9ec0b7009c87970b99c4de7b3c53c44768e6c9565babd627c83971bae8a32001f20f7acf888a80a0897f8409"
        }`
	jsonGenesis = `{
        	"type": "ContractReward",
        	"token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
        	"address": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        	"balance": "60000000000000000",
        	"vote": "0",
        	"network": "0",
        	"storage": "0",
        	"oracle": "0",
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"link": "c0d330096ec4ab6ccf5481e06cc54e74b14f534e99e38df486f47d1123cbd1ae",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "Rd0hfNn/ife2TO2kiGzGjd6d+keopCLRZeLOb5qDT60AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANUprp6GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACGgb9SU8ZGcvxUyT07W5og0olly4+AunBGDtP5nLVHvVAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGgb9SU8ZGcvxUyT07W5og0olly4+AunBGDtP5nLVHvVAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANRTEMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADUUxDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
        	"povHeight": 0,
        	"timestamp": 1553990410,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        	"work": "000000000048f5b9",
        	"signature": "76fb8d1e5decf145616ee1b5713ad81b5792cadcc054a9d999d5acf026924e8ccb23f96994cf8a056b5e0a74706e4ccb7314dc989b1fccbcd88c024bb31c6b0d"
        }`

	jsonMintageQGAS = `{
        	"type": "ContractSend",
        	"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
        	"address": "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
        	"balance": "0",
        	"vote": "0",
        	"network": "0",
        	"storage": "0",
        	"oracle": "0",
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"link": "dd803c6bdb9878a35a3e447e4d88b8df4f65eecd35d7506157750ea46e2c6133",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "6TrdxOqEIjTk3FsXwzs1+ZtbhhEaOvC9jkqIImArhmcR3m2BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAjhvJvwQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAjdgDxr25h4o1o+RH5NiLjfT2XuzTXXUGFXdQ6kbixhMwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEUUdBUwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABFFHQVMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        	"povHeight": 0,
        	"timestamp": 1553990401,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        	"work": "000000000048f5b9",
        	"signature": "175df7302eb8d830da4d86047819b4680f5a3c425ed8bc6593a4dd724c847aceac68b63bebfbde94d368b4e24182dd503f05ac474d231c6a17455fa131f0d10f"
        }`
	jsonGenesisQGAS = `{
        	"type": "ContractReward",
        	"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
        	"address": "qlc_3qe19joxq85rnff5wj5ybp6djqtheqqetfgqc3iogxagnjq4rrbmbp1ews7d",
        	"balance": "10000000000000000",
        	"vote": "0",
        	"network": "0",
        	"storage": "0",
        	"oracle": "0",
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"link": "bdac41b3ff7ac35aee3028d60eabeb9578ea6f7bd148d611133a3b26dfa6a9be",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "6oQiNOTcWxfDOzX5m1uGERo68L2OSogiYCuGZxHebYEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACOG8m/BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACN2APGvbmHijWj5Efk2IuN9PZe7NNddQYVd1DqRuLGEzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAN2APGvbmHijWj5Efk2IuN9PZe7NNddQYVd1DqRuLGEzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARRR0FTAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEUUdBUwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
        	"povHeight": 0,
        	"timestamp": 1553990410,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        	"work": "000000000048f5b9",
        	"signature": "4e605d4ccfe3f061c1139eecec25b1129dbe491b12c42eb34e3febadc40971ff42e7240f9f42432591172cc5a91292406fe7b3115c757945e9c8848b83836001"
        }`

	testJsonMintage = `{
        	"type": "ContractSend",
        	"token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
        	"address": "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
        	"balance": "0",
        	"vote": "0",
        	"network": "0",
        	"storage": "0",
        	"oracle": "0",
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"link": "bf86c83fb4bfb9f49b9b8fa593c8cf4128c9e21720c487565b52fc6640a9e8f3",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "6TrdxKfo+jDAY+lqSJpHvEOQlQW9hnNdpKEJ3KKL6TYRioWCAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADVKa6ehgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAi/hsg/tL+59Jubj6WTyM9BKMniFyDEh1ZbUvxmQKno8wAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADUUxDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1FMQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        	"povHeight": 0,
        	"timestamp": 1553990401,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        	"work": "000000000048f5b9",
        	"signature": "18fcd023d9c14bede4a5abad71eca562c8bd5df48ac293466b48b3c2ece42fa00c0e64b68cccf5cfa0f0077ebdc9e05efd757999e3f3c68280975458e9cad40e"
        }`

	testJsonGenesis = `{
        	"type": "ContractReward",
        	"token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
        	"address": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        	"balance": "60000000000000000",
        	"vote": "0",
        	"network": "0",
        	"storage": "0",
        	"oracle": "0",
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"link": "8b54787c668dddd4f22ad64a8b0d241810871b9a52a989eb97670f345ad5dc90",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "p+j6MMBj6WpImke8Q5CVBb2Gc12koQncoovpNhGKhYIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANUprp6GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACL+GyD+0v7n0m5uPpZPIz0EoyeIXIMSHVltS/GZAqejzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAL+GyD+0v7n0m5uPpZPIz0EoyeIXIMSHVltS/GZAqejzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANRTEMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADUUxDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
        	"povHeight": 0,
        	"timestamp": 1553990410,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        	"work": "000000000048f5b9",
        	"signature": "a717e690216e357d1b4e200478ef74c51d0b6ab28893fd5cf22aff6f1403d60996ec97bd84214c7a7aed4b0428671e81048afa9b86126c7484b3a88b725e1202"
        }`

	testJsonMintageQGAS = `{
        	"type": "ContractSend",
        	"token": "89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4",
        	"address": "qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1",
        	"balance": "0",
        	"vote": "0",
        	"network": "0",
        	"storage": "0",
        	"oracle": "0",
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"link": "e813e51a6d8abea178a2f376d532df983cca71b4e4cf5bdd2d7864ee30cf8ba5",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "6TrdxIkGbXR6PHT/HeyOpqcBG94BDdQErsRUiA8j1Yy/koDkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAjhvJvwQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAjoE+UabYq+oXii83bVMt+YPMpxtOTPW90teGTuMM+LpQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEUUdBUwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABFFHQVMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        	"povHeight": 0,
        	"timestamp": 1553990401,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        	"work": "000000000048f5b9",
        	"signature": "441b26cf4318cea394fe07a5e30cde18f967406a9c26158417bcd29abd5a4c79d05746f838bc42f0a7d681cf4a3b4e6b29992fcd7fa7cafe72a4e00e133d310f"
        }`

	testJsonGenesisQGAS = `{
        	"type": "ContractReward",
        	"token": "89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4",
        	"address": "qlc_3t1mwnf8u4oyn7wc7wuptnsfz83wsbrubs8hdhgkty56xrrez4x7fcttk5f3",
        	"balance": "10000000000000000",
        	"vote": "0",
        	"network": "0",
        	"storage": "0",
        	"oracle": "0",
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"link": "f798089896ffdf45ccce2e039666014b8c666ea0f47f0df4ee7e73b49dac0945",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "iQZtdHo8dP8d7I6mpwEb3gEN1ASuxFSIDyPVjL+SgOQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACOG8m/BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACOgT5Rptir6heKLzdtUy35g8ynG05M9b3S14ZO4wz4ulAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAOgT5Rptir6heKLzdtUy35g8ynG05M9b3S14ZO4wz4ulAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARRR0FTAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEUUdBUwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
        	"povHeight": 0,
        	"timestamp": 1553990410,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        	"work": "000000000048f5b9",
        	"signature": "69903343b5188cedc3b301288a553f3e094bffdf8d1173eb897e630860642a4d62d1022b002c290ca996027d5e424056adad04b340f52d0185362dfd41e07e0c"
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

	//main net gas
	gasToken, _     = types.NewHash("ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81")
	gasAddress, _   = types.HexToAddress("qlc_3qe19joxq85rnff5wj5ybp6djqtheqqetfgqc3iogxagnjq4rrbmbp1ews7d")
	gasMintageBlock types.StateBlock
	gasMintageHash  types.Hash
	gasBlock        types.StateBlock
	gasBlockHash    types.Hash

	//test net
	testChainToken, _       = types.NewHash("a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582")
	testGenesisAddress, _   = types.HexToAddress("qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4")
	testGenesisMintageBlock types.StateBlock
	testGenesisMintageHash  types.Hash
	testGenesisBlock        types.StateBlock
	testGenesisBlockHash    types.Hash

	//test net gas
	testGasToken, _     = types.NewHash("89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4")
	testGasAddress, _   = types.HexToAddress("qlc_3t1mwnf8u4oyn7wc7wuptnsfz83wsbrubs8hdhgkty56xrrez4x7fcttk5f3")
	testGasMintageBlock types.StateBlock
	testGasMintageHash  types.Hash
	testGasBlock        types.StateBlock
	testGasBlockHash    types.Hash
)

func init() {
	_ = json.Unmarshal([]byte(jsonMintage), &genesisMintageBlock)
	_ = json.Unmarshal([]byte(jsonGenesis), &genesisBlock)
	genesisMintageHash = genesisMintageBlock.GetHash()
	genesisBlockHash = genesisBlock.GetHash()
	//main net gas
	_ = json.Unmarshal([]byte(jsonMintageQGAS), &gasMintageBlock)
	_ = json.Unmarshal([]byte(jsonGenesisQGAS), &gasBlock)
	gasMintageHash = gasMintageBlock.GetHash()
	gasBlockHash = gasBlock.GetHash()

	_ = json.Unmarshal([]byte(testJsonMintage), &testGenesisMintageBlock)
	_ = json.Unmarshal([]byte(testJsonGenesis), &testGenesisBlock)
	testGenesisMintageHash = testGenesisMintageBlock.GetHash()
	testGenesisBlockHash = testGenesisBlock.GetHash()
	//test net gas
	_ = json.Unmarshal([]byte(testJsonMintageQGAS), &testGasMintageBlock)
	_ = json.Unmarshal([]byte(testJsonGenesisQGAS), &testGasBlock)
	testGasMintageHash = testGasMintageBlock.GetHash()
	testGasBlockHash = testGasBlock.GetHash()
}

func GenesisAddress() types.Address {
	if goqlc.MAINNET {
		return genesisAddress
	}
	return testGenesisAddress
}

func GasAddress() types.Address {
	if goqlc.MAINNET {
		return gasAddress
	}
	return testGasAddress
}

func ChainToken() types.Hash {
	if goqlc.MAINNET {
		return chainToken
	}
	return testChainToken
}

func GasToken() types.Hash {
	if goqlc.MAINNET {
		return gasToken
	}
	return testGasToken
}

func GenesisMintageBlock() types.StateBlock {
	if goqlc.MAINNET {
		return genesisMintageBlock
	}
	return testGenesisMintageBlock
}

func GasMintageBlock() types.StateBlock {
	if goqlc.MAINNET {
		return gasMintageBlock
	}
	return testGasMintageBlock
}

func GenesisMintageHash() types.Hash {
	if goqlc.MAINNET {
		return genesisMintageHash
	}
	return testGenesisMintageHash
}

func GasMintageHash() types.Hash {
	if goqlc.MAINNET {
		return gasMintageHash
	}
	return testGasMintageHash
}

func GenesisBlock() types.StateBlock {
	if goqlc.MAINNET {
		return genesisBlock
	}
	return testGenesisBlock
}

func GasBlock() types.StateBlock {
	if goqlc.MAINNET {
		return gasBlock
	}
	return testGasBlock
}

func GenesisBlockHash() types.Hash {
	if goqlc.MAINNET {
		return genesisBlockHash
	}
	return testGenesisBlockHash
}

func GasBlockHash() types.Hash {
	if goqlc.MAINNET {
		return gasBlockHash
	}
	return testGasBlockHash
}

// IsGenesis check block is chain token genesis
func IsGenesisBlock(block *types.StateBlock) bool {
	h := block.GetHash()
	if goqlc.MAINNET {
		return h == genesisMintageHash || h == genesisBlockHash || h == gasMintageHash || h == gasBlockHash
	}

	return h == testGenesisMintageHash || h == testGenesisBlockHash || h == testGasMintageHash || h == testGasBlockHash
}

// IsGenesis check token is chain token genesis
func IsGenesisToken(hash types.Hash) bool {
	if goqlc.MAINNET {
		return hash == chainToken || hash == gasToken
	}
	return hash == testChainToken || hash == testGasToken
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
