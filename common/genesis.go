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
        "token": "581150e94662b73130656d996ae3049ac49d21840c4cbf5b483179b25f98ebf4",
        "address": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        "balance": "0",
        "previous": "0000000000000000000000000000000000000000000000000000000000000000",
        "link": "de32f02da71ef2fccd06634bfe29d3a7514a1880873478382704e3edeeaff982",
        "message": "0000000000000000000000000000000000000000000000000000000000000000",
        "data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAOAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADVKa6ehgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAhoG/UlPGRnL8VMk9O1uaINKJZcuPgLpwRg7T+Zy1R71QAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACFFMQyBDb2luAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANRTEMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
        "quota": 0,
        "timestamp": 1553990401,
        "extra": "0000000000000000000000000000000000000000000000000000000000000000",
        "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        "work": "000000000054b46a",
        "signature": "9f52e689a7b03722ef8b98701fe16e18f84e020279ee021012bd2f4b32aefa228bfd9e16555149ddca435c2401ce4d44c539f0a93953a95bc225de806beea900"
    }`
	jsonGenesis = `{
        "type": "ContractReward",
        "token": "581150e94662b73130656d996ae3049ac49d21840c4cbf5b483179b25f98ebf4",
        "address": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        "balance": "60000000000000000",
        "previous": "6d1c9d414d486493a29aad8f6baef08fe05539f0e7665337351d931bede40ec7",
        "link": "de32f02da71ef2fccd06634bfe29d3a7514a1880873478382704e3edeeaff982",
        "message": "0000000000000000000000000000000000000000000000000000000000000000",
        "quota": 0,
        "timestamp": 1553990410,
        "extra": "0000000000000000000000000000000000000000000000000000000000000000",
        "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
        "work": "00000000008f6cf4",
        "signature": "4087552d5d2f7451e39bf7d2e7e9e0f481c433c0a3a1ee4f7cea6acfcf05bd1f86c9510fc2ec0feff4f71d179778f1c00adebd288d3c1cad30d1169d3c116e04"
    }`

	units = map[string]*big.Int{
		"qlc":  big.NewInt(1),
		"Kqlc": big.NewInt(1e5),
		"Mqlc": big.NewInt(1e8),
		"Gqlc": big.NewInt(1e11),
	}

	QLCChainToken, _         = types.NewHash("581150e94662b73130656d996ae3049ac49d21840c4cbf5b483179b25f98ebf4")
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
