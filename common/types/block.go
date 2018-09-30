/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/json"

	"github.com/tinylib/msgp/msgp"
)

//msgp:shim Enum as:string using:(Enum).String/parseString
type Enum byte

const (
	State Enum = iota
	SmartContract
	Invalid
)

func (e Enum) String() string {
	switch e {
	case State:
		return "State"
	case SmartContract:
		return "SmartContract"
	default:
		return "<invalid>"
	}
}

func parseString(s string) Enum {
	switch s {
	case "State":
		return State
	case "SmartContract":
		return SmartContract
	default:
		return Invalid
	}
}

type Block interface {
	Type() Enum
	Hash() Hash
	Addresses() []*Address
	PreviousHash() Hash
	Representative() Address
	Balance() Balance
	Link() Hash
	Signature() Signature
	Token() Hash
	Extra() Hash
	Work() Work
	msgp.Decodable
	msgp.Encodable
	msgp.Marshaler
	msgp.Unmarshaler
	json.Marshaler
	json.Unmarshaler
}

//go:generate msgp
type StateBlock struct {
	Type           Enum      `msg:"type" json:"type"`
	Address        Address   `msg:"addresses,extension" json:"addresses"`
	PreviousHash   Hash      `msg:"previous,extension" json:"previous"`
	Representative Address   `msg:"representative,extension" json:"representative"`
	Balance        Balance   `msg:"balance,extension" json:"balance"`
	Link           Hash      `msg:"link,extension" json:"link"`
	Signature      Signature `msg:"signature,extension" json:"signature"`
	Token          Hash      `msg:"token,extension" json:"token"`
	Work           Work      `msg:"work,extension" json:"work"`
}

//go:generate msgp
type SmartContractBlock struct {
	Type           Enum       `msg:"type" json:"type"`
	Address        []*Address `msg:"addresses,extension" json:"addresses"`
	PreviousHash   Hash       `msg:"previous,extension" json:"previous"`
	Representative Address    `msg:"representative,extension" json:"representative"`
	Balance        Balance    `msg:"balance,extension" json:"balance"`
	Link           Hash       `msg:"link,extension" json:"link"`
	Signature      Signature  `msg:"signature,extension" json:"signature"`
	Extra          Hash       `msg:"extra,extension" json:"extra"`
	Work           Work       `msg:"work,extension" json:"work"`
	Owner          Address    `msg:"owner,extension" json:"owner"`
	Issuer         Address    `msg:"issuer,extension" json:"issuer"`
}

//go:generate msgp
type BlockExtra struct {
	KeyHash Hash    `msg:"key,extension" json:"key"`
	Abi     []byte  `msg:"abi" json:"abi"`
	Issuer  Address `msg:"issuer,extension" json:"issuer"`
}
