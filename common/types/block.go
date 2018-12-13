/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors" //"encoding/json"
	"strings"

	"github.com/tinylib/msgp/msgp"
	"golang.org/x/crypto/blake2b"
)

//msgp:shim BlockType as:string using:(BlockType).String/parseString
type BlockType byte

const (
	State BlockType = iota
	SmartContract
	Invalid
)

var (
	ErrBadBlockType = errors.New("bad block type")
	ErrNotABlock    = errors.New("block type is not_a_block")
)

const (
	preambleSize = 32
	//blockSizeCommon = SignatureSize + WorkSize
	//blockSizeState  = blockSizeCommon
)

func (e BlockType) String() string {
	switch e {
	case State:
		return "State"
	case SmartContract:
		return "SmartContract"
	default:
		return "<invalid>"
	}
}

func parseString(s string) BlockType {
	switch strings.ToLower(s) {
	case "state":
		return State
	case "smartcontract":
		return SmartContract
	default:
		return Invalid
	}
}

func (e *BlockType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*e = parseString(j)
	return nil
}

func (e BlockType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(e.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func NewBlock(t BlockType) (Block, error) {
	switch t {
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

type Block interface {
	GetType() BlockType
	GetHash() Hash
	GetAddress() Address
	GetSignature() Signature
	GetWork() Work
	GetPrevious() Hash
	Root() Hash
	Size() int
	IsValid() bool

	msgp.Decodable
	msgp.Encodable
	msgp.Marshaler
	msgp.Unmarshaler
	//json.Marshaler
	//json.Unmarshaler
}

//go:generate msgp
type CommonBlock struct {
	Type      BlockType `msg:"type" json:"type"`
	Address   Address   `msg:"addresses,extension" json:"addresses"`
	Previous  Hash      `msg:"previous,extension" json:"previous"`
	Signature Signature `msg:"signature,extension" json:"signature"`
	Work      Work      `msg:"work,extension" json:"work"`
	Extra     Hash      `msg:"extra,extension" json:"extra"`
}

func (b *CommonBlock) GetAddress() Address {
	return b.Address
}

func (b *CommonBlock) GetPrevious() Hash {
	return b.Previous
}

func (b *CommonBlock) GetSignature() Signature {
	return b.Signature
}

func (b *CommonBlock) GetWork() Work {
	return b.Work
}

func (b *CommonBlock) GetExtra() Hash {
	return b.Extra
}

//go:generate msgp
type StateBlock struct {
	CommonBlock
	Token          Hash    `msg:"token,extension" json:"token"`
	Balance        Balance `msg:"balance,extension" json:"balance"`
	Link           Hash    `msg:"link,extension" json:"link"`
	Representative Address `msg:"representative,extension" json:"representative"`
}

func (b *StateBlock) GetType() BlockType {
	return State
}

func (b *StateBlock) GetHash() Hash {
	var preamble [preambleSize]byte
	preamble[len(preamble)-1] = byte(State)
	return hashBytes(preamble[:], b.Address[:], b.Previous[:], b.Representative[:], b.Balance.Bytes(binary.BigEndian), b.Link[:], b.Extra[:])
}

func (b *StateBlock) GetRepresentative() Address {
	return b.Representative
}

func (b *StateBlock) GetBalance() Balance {
	return b.Balance
}

func (b *StateBlock) GetLink() Hash {
	return b.Link
}

func (b *StateBlock) GetToken() Hash {
	return b.Token
}

func (b *StateBlock) Root() Hash {
	if b.isOpen() {
		return b.Address.ToHash()
	}

	return b.Previous
}

func (b *StateBlock) Size() int {
	return b.Msgsize()
}

func (b *StateBlock) IsValid() bool {
	if b.isOpen() {
		return b.Work.IsValid(b.Previous)
	}

	return b.Work.IsValid(Hash(b.Address))
}

func (b *StateBlock) isOpen() bool {
	return !b.Previous.IsZero()
}

//go:generate msgp
type BlockExtra struct {
	KeyHash Hash    `msg:"key,extension" json:"key"`
	Abi     []byte  `msg:"abi" json:"abi"`
	Issuer  Address `msg:"issuer,extension" json:"issuer"`
}

//go:generate msgp
type SmartContractBlock struct {
	CommonBlock
	Owner        Address   `msg:"owner,extension" json:"owner"`
	Issuer       Address   `msg:"issuer,extension" json:"issuer"`
	ExtraAddress []Address `msg:"extraAddress,extension" json:"extraAddress"`
}

func (sc *SmartContractBlock) Root() Hash {
	return sc.Address.ToHash()
}

func (sc *SmartContractBlock) Size() int {
	return sc.Msgsize()
}

//TODO: improvement
func (sc *SmartContractBlock) IsValid() bool {
	return sc.Work.IsValid(sc.Address.ToHash())
}

func (sc *SmartContractBlock) GetExtraAddress() []Address {
	return sc.ExtraAddress
}

func (sc *SmartContractBlock) GetType() BlockType {
	return SmartContract
}

func (sc *SmartContractBlock) GetHash() Hash {
	var preamble [preambleSize]byte
	preamble[len(preamble)-1] = byte(SmartContract)
	var a []byte
	for _, addr := range sc.ExtraAddress {
		a = append(a, addr[:]...)
	}
	return hashBytes(preamble[:], a, sc.Address[:], sc.Previous[:], sc.Extra[:], sc.Owner[:], sc.Issuer[:])
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
