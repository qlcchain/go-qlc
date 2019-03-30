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
        	"vote": "0",
        	"network": "0",
        	"storage": "0",
        	"oracle": "0",
        	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
        	"link": "0000000000000000000000000000000000000000000000000000000000000000",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "6TrdxEXdIXzZ/4n3tkztpIhsxo3enfpHqKQi0WXizm+ag0+tAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADVKa6ehgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAhoG/UlPGRnL8VMk9O1uaINKJZcuPgLpwRg7T+Zy1R71QAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADUUxDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1FMQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        	"povHeight": 0,
        	"timestamp": 1553990401,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        	"work": "000000000048f5b9",
        	"signature": "02c2d305fae4717cf132c8f653be5bcd8cd01a671af33ad144353e66fcbe7262224de0f929934d4871b8a1c971cf29131e4a4adceff856edc69a15bb8d759e0a"
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
        	"link": "90f28436423396887ccb08362b62061ca4b3c5a297a84e30f405e8973f652484",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "Rd0hfNn/ife2TO2kiGzGjd6d+keopCLRZeLOb5qDT60AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANUprp6GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACGgb9SU8ZGcvxUyT07W5og0olly4+AunBGDtP5nLVHvVAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGgb9SU8ZGcvxUyT07W5og0olly4+AunBGDtP5nLVHvVAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANRTEMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADUUxDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
        	"povHeight": 0,
        	"timestamp": 1553990410,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        	"work": "000000000048f5b9",
        	"signature": "d468c4c750eca3cb3a4005918f4dba4b72f4882a39caafc057d32f120e54e8173cd548f1779890cdff02f11ac2e09d62003327b0e027d2c09d78db6302b9fc07"
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
        	"link": "0000000000000000000000000000000000000000000000000000000000000000",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "6TrdxOqEIjTk3FsXwzs1+ZtbhhEaOvC9jkqIImArhmcR3m2BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAjhvJvwQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAjdgDxr25h4o1o+RH5NiLjfT2XuzTXXUGFXdQ6kbixhMwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEUUdBUwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABFFHQVMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        	"povHeight": 0,
        	"timestamp": 1553990401,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        	"work": "000000000048f5b9",
        	"signature": "ce19037a1ebde7a276704096ca40ab566644cfc25cdcf6648de095d79e4cbb25f327f4d98c506d3e249b79fbbae887d3a6aec8d7f1a4f3d00623b20e491ded00"
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
        	"link": "3b8c3acfbef2a93d9ba506073976f293cc1cca98892b7c545603945dd78f824f",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "6oQiNOTcWxfDOzX5m1uGERo68L2OSogiYCuGZxHebYEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACOG8m/BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACN2APGvbmHijWj5Efk2IuN9PZe7NNddQYVd1DqRuLGEzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAN2APGvbmHijWj5Efk2IuN9PZe7NNddQYVd1DqRuLGEzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARRR0FTAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEUUdBUwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
        	"povHeight": 0,
        	"timestamp": 1553990410,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        	"work": "000000000048f5b9",
        	"signature": "7c41aa00427489b264f04f73c7d3e40391107eb50774d55543be3714d664a2627b3c71ef4e1ef87aa45564369d68605b907b9f8097cd53c53fd5435050d95d01"
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
        	"link": "0000000000000000000000000000000000000000000000000000000000000000",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "6TrdxKfo+jDAY+lqSJpHvEOQlQW9hnNdpKEJ3KKL6TYRioWCAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADVKa6ehgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAi/hsg/tL+59Jubj6WTyM9BKMniFyDEh1ZbUvxmQKno8wAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADUUxDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1FMQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        	"povHeight": 0,
        	"timestamp": 1553990401,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        	"work": "000000000048f5b9",
        	"signature": "296a9af6b30d6b05216a4cb86996a54442d1e8d0e6fd7ad7636f294d31475474ec61147ab9524a8f98d9c110852591910f95cede1b125752c6217d3b6e8a4906"
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
        	"link": "67513e803863279bc62d8e49a087b623895c8e2b21160a874f337ce147c859f1",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "p+j6MMBj6WpImke8Q5CVBb2Gc12koQncoovpNhGKhYIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANUprp6GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACL+GyD+0v7n0m5uPpZPIz0EoyeIXIMSHVltS/GZAqejzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAL+GyD+0v7n0m5uPpZPIz0EoyeIXIMSHVltS/GZAqejzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANRTEMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADUUxDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
        	"povHeight": 0,
        	"timestamp": 1553990410,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        	"work": "000000000048f5b9",
        	"signature": "60f7265b7379cc41d7ec533a4d2de08840e8d15d63599aa7e665c12d199d91b6ead07ab4a2b437e3e3957e60cf63bb79dcb1a82670339f7b3568b5207fa83b0d"
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
        	"link": "0000000000000000000000000000000000000000000000000000000000000000",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "6TrdxIkGbXR6PHT/HeyOpqcBG94BDdQErsRUiA8j1Yy/koDkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAjhvJvwQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAjoE+UabYq+oXii83bVMt+YPMpxtOTPW90teGTuMM+LpQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEUUdBUwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABFFHQVMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        	"povHeight": 0,
        	"timestamp": 1553990401,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        	"work": "000000000048f5b9",
        	"signature": "d1391de3eb0b197065fbaa39008cb1f18db440e208811883f8f9edfc7a34dbff4990947da3f25cdd1d9a210d064e38c30cf70344ae44c7a02a1a2848446a4f03"
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
        	"link": "327531148b1a6302632aa7ad6eb369437d8269a08a55b344bd06b514e4e6ae97",
        	"message": "0000000000000000000000000000000000000000000000000000000000000000",
        	"data": "iQZtdHo8dP8d7I6mpwEb3gEN1ASuxFSIDyPVjL+SgOQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACOG8m/BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACOgT5Rptir6heKLzdtUy35g8ynG05M9b3S14ZO4wz4ulAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAOgT5Rptir6heKLzdtUy35g8ynG05M9b3S14ZO4wz4ulAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARRR0FTAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEUUdBUwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
        	"povHeight": 0,
        	"timestamp": 1553990410,
        	"extra": "0000000000000000000000000000000000000000000000000000000000000000",
        	"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        	"work": "000000000048f5b9",
        	"signature": "856b60691076be6ce272389fbd7b81480451a2b22bf35d9c15110ecf1c8d2b994f6c02fe2c3e2bf0f22b69df37d50cc8ee1bf7c6989b2996aac357d0b8904f00"
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
	return hash == testChainToken || hash == testGasBlockHash
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
