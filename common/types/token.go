/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import "math/big"

type TokenInfo struct {
	TokenId       Hash     `json:"tokenId"`
	TokenName     string   `json:"tokenName"`
	TokenSymbol   string   `json:"tokenSymbol"`
	TotalSupply   *big.Int `json:"totalSupply"`
	Decimals      uint8    `json:"decimals"`
	Owner         Address  `json:"owner"`
	PledgeAmount  *big.Int `json:"pledgeAmount"`
	WithdrawTime  int64    `json:"withdrawTime"`
	PledgeAddress Address  `json:"pledgeAddress"`
	NEP5TxId      string   `json:"NEP5TxId"`
}

type GenesisTokenInfo struct {
	TokenId       Hash     `json:"tokenId"`
	TokenName     string   `json:"tokenName"`
	TokenSymbol   string   `json:"tokenSymbol"`
	TotalSupply   *big.Int `json:"totalSupply"`
	Decimals      uint8    `json:"decimals"`
	Owner         Address  `json:"owner"`
	PledgeAmount  *big.Int `json:"pledgeAmount"`
	WithdrawTime  int64    `json:"withdrawTime"`
	PledgeAddress Address  `json:"pledgeAddress"`
}
