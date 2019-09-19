/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"math/big"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/abi"
)

const (
	jsonDestroy = `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      { "name": "owner", "type": "address" },
      { "name": "previous", "type": "hash" },
      { "name": "tokenId", "type": "tokenId" },
      { "name": "amount", "type": "uint256" },
      { "name": "signature", "type": "signature" }
    ]
  },
  {
    "type": "variable",
    "name": "destroyInfo",
    "inputs": [
      { "name": "owner", "type": "address" },
      { "name": "previous", "type": "hash" },
      { "name": "tokenId", "type": "tokenId" },
      { "name": "amount", "type": "uint256" },
      { "name": "time", "type": "int64" }
    ]
  }
]
`

	MethodNameDestroy   = "Destroy"
	VariableDestroyInfo = "destroyInfo"
)

var (
	BlackHoleABI, _ = abi.JSONToABIContract(strings.NewReader(jsonDestroy))
)

type DestroyParam struct {
	Owner    types.Address   `json:"owner"`
	Token    types.Hash      `json:"token"`
	Previous types.Hash      `json:"previous"`
	Amount   *big.Int        `json:"amount"`
	Sign     types.Signature `json:"signature"`
}

type DestroyInfo struct {
	Owner    types.Address `json:"owner"`
	Token    types.Hash    `json:"token"`
	Previous types.Hash    `json:"previous"`
	Amount   *big.Int      `json:"amount"`
	Time     int64         `json:"time"`
}
