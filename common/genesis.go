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
	"github.com/qlcchain/go-qlc/common/types"
	"math/big"
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

	units = map[string]*big.Int{
		"qlc":  big.NewInt(1),
		"Kqlc": big.NewInt(1e5),
		"QLC":  big.NewInt(1e8),
		"MQLC": big.NewInt(1e11),
	}

	QLCChainToken, _         = types.NewHash("45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad")
	GenesisAccountAddress, _ = types.HexToAddress("qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44")
	GenesisMintageBlock      types.StateBlock
	GenesisMintageHash       types.Hash
	QLCGenesisBlock          types.StateBlock
	QLCGenesisBlockHash      types.Hash
)

func init() {
	_ = json.Unmarshal([]byte(jsonMintage), &GenesisMintageBlock)
	_ = json.Unmarshal([]byte(jsonGenesis), &QLCGenesisBlock)
	GenesisMintageHash = GenesisMintageBlock.GetHash()
	QLCGenesisBlockHash = QLCGenesisBlock.GetHash()
}

// IsGenesis check block is chain token genesis
func IsGenesisBlock(block *types.StateBlock) bool {
	h := block.GetHash()
	return h == GenesisMintageHash || h == QLCGenesisBlockHash
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
