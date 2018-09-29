package ledger

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/tinylib/msgp/msgp"
)

//msgp:shim Enum as:string using:(Enum).String/parseString
type Enum byte

const (
	State Enum = iota
	SmartContract
)

type Block interface {
	Type() Enum
	Hash() types.Hash
	Addresses() []*types.Address
	PreviousHash() types.Hash
	Representative() types.Address
	Balance() types.Balance
	Link() types.Hash
	Signature() types.Signature
	Token() types.Hash
	Extra() types.Hash
	Work() types.Work
	msgp.Decodable
	msgp.Encodable
	msgp.Marshaler
	msgp.Unmarshaler
	json.Marshaler
	json.Unmarshaler
}

//go:generate msgp
type StateBlock struct {
	Type           Enum            `msg:"type" json:"type"`
	Address        types.Address   `msg:"addresses" json:"addresses"`
	PreviousHash   types.Hash      `msg:"previous" json:"previous"`
	Representative types.Address   `msg:"representative" json:"representative"`
	Balance        types.Balance   `msg:"balance" json:"balance"`
	Link           types.Hash      `msg:"link" json:"link"`
	Signature      types.Signature `msg:"signature" json:"signature"`
	Token          types.Hash      `msg:"token" json:"token"`
	Work           types.Work      `msg:"work" json:"work"`
}

//go:generate msgp
type SmartContractBlock struct {
	Type           Enum             `msg:"type" json:"type"`
	Address        []*types.Address `msg:"addresses" json:"addresses"`
	PreviousHash   types.Hash       `msg:"previous" json:"previous"`
	Representative types.Address    `msg:"representative" json:"representative"`
	Balance        types.Balance    `msg:"balance" json:"balance"`
	Link           types.Hash       `msg:"link" json:"link"`
	Signature      types.Signature  `msg:"signature" json:"signature"`
	Extra          types.Hash       `msg:"extra" json:"extra"`
	Work           types.Work       `msg:"work" json:"work"`
	Owner          types.Address    `msg:"owner" json:"owner"`
	Issuer         types.Address    `msg:"issuer" json:"issuer"`
}

//go:generate msgp
type BlockExtra struct {
	KeyHash types.Hash    `msg:"key" json:"key"`
	Abi     []byte        `msg:"abi" json:"abi"`
	Issuer  types.Address `msg:"issuer" json:"issuer"`
}
