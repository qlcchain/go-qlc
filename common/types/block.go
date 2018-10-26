/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/json"
	"errors" //"encoding/json"

	"github.com/tinylib/msgp/msgp"
	"golang.org/x/crypto/blake2b"
)

//msgp:shim Enum as:string using:(Enum).String/parseString
type Enum byte

const (
	State Enum = iota
	SmartContract
	Invalid
)

var (
	ErrBadBlockType = errors.New("bad block type")
	ErrNotABlock    = errors.New("block type is not_a_block")
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
	//Type() Enum
	//Hash() Hash
	//Addresses() []*Address
	//PreviousHash() Hash
	//Representative() Address
	//Balance() Balance
	//Link() Hash
	//Signature() Signature
	//Token() Hash
	//Extra() Hash
	//Work() Work
	Hash() Hash
	ID() Enum
	Root() Hash
	Size() int

	msgp.Decodable
	msgp.Encodable
	msgp.Marshaler
	msgp.Unmarshaler
	//json.Marshaler
	//json.Unmarshaler
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

func NewBlock(blockType byte) (Block, error) {
	switch Enum(blockType) {
	case State:
		return new(StateBlock), nil
	case SmartContract:
		return new(SmartContractBlock), nil
	case Invalid:
		return nil, ErrNotABlock
	default:
		return nil, ErrBadBlockType
	}
}

func hashBytes(inputs ...[]byte) Hash {
	hash, err := blake2b.New(blake2b.Size256, nil)
	if err != nil {
		panic(err)
	}

	for _, data := range inputs {
		hash.Write(data)
	}

	var result Hash
	copy(result[:], hash.Sum(nil))
	return result
}

func (b *StateBlock) Hash() Hash {
	return hashBytes(b.Link[:], b.Representative[:], b.Address[:])
}

func (b *StateBlock) ID() Enum {
	return State
}

func (b *StateBlock) Root() Hash {
	panic("implement me")
}

func (b *StateBlock) Size() int {
	panic("implement me")
}

func ParseStateBlock(b []byte) (*StateBlock, error) {
	var blk StateBlock
	var values map[string]interface{}
	err := json.Unmarshal(b, &values)
	if err != nil {
		return nil, err
	}
	id, ok := values["type"]
	if !ok || id != "state" {
		return nil, errors.New("err block type")
	}
	blk.Type = State
	if blk.Address, err = HexToAddress(values["address"].(string)); err != nil {
		return nil, err
	}
	if err = blk.PreviousHash.Of(values["previousHash"].(string)); err != nil {
		return nil, err
	}
	if blk.Representative, err = HexToAddress(values["representative"].(string)); err != nil {
		return nil, err
	}
	if blk.Balance, err = ParseBalance(values["balance"].(string), "Mqlc"); err != nil {
		return nil, err
	}
	if err = blk.Link.Of(values["link"].(string)); err != nil {
		return nil, err
	}
	if err = blk.Signature.Of(values["signature"].(string)); err != nil {
		return nil, err
	}
	if err = blk.Token.Of(values["token"].(string)); err != nil {
		return nil, err
	}
	if err = blk.Work.ParseWorkHexString(values["work"].(string)); err != nil {
		return nil, err
	}
	return &blk, nil
}

func (b *SmartContractBlock) Hash() Hash {
	return hashBytes(b.Link[:], b.Representative[:])
}

func (b *SmartContractBlock) ID() Enum {
	return SmartContract
}

func (b *SmartContractBlock) Root() Hash {
	panic("implement me")
}

func (b *SmartContractBlock) Size() int {
	panic("implement me")
}
