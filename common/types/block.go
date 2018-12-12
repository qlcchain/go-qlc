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

const (
	preambleSize = 32
	//blockSizeCommon = SignatureSize + WorkSize
	//blockSizeState  = blockSizeCommon
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
	switch strings.ToLower(s) {
	case "state":
		return State
	case "smartcontract":
		return SmartContract
	default:
		return Invalid
	}
}

func (e *Enum) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*e = parseString(j)
	return nil
}

func (e Enum) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(e.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

type Block interface {
	GetType() Enum
	GetHash() Hash
	GetAddress() Address
	GetPreviousHash() Hash
	GetRepresentative() Address
	GetBalance() Balance
	GetExtraAddress() []Address
	GetLink() Hash
	GetSignature() Signature
	GetToken() Hash
	GetExtra() Hash
	GetWork() Work

	ID() Enum
	Root() Hash
	Size() int
	Valid(threshold uint64) bool

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
	Extra          Hash      `msg:"extra,extension" json:"extra"`
	Work           Work      `msg:"work,extension" json:"work"`
}

func (b *StateBlock) GetType() Enum {
	return State
}

func (b *StateBlock) GetHash() Hash {
	var preamble [preambleSize]byte
	preamble[len(preamble)-1] = byte(State)
	return hashBytes(preamble[:], b.Address[:], b.PreviousHash[:], b.Representative[:], b.Balance.Bytes(binary.BigEndian), b.Link[:], b.Extra[:])
}

func (b *StateBlock) GetPreviousHash() Hash {
	return b.PreviousHash
}

func (b *StateBlock) GetAddress() Address {
	return b.Address
}

func (b *StateBlock) GetExtraAddress() []Address {
	return nil
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

func (b *StateBlock) GetSignature() Signature {
	return b.Signature
}

func (b *StateBlock) GetToken() Hash {
	return b.Token
}

func (b *StateBlock) GetExtra() Hash {
	return b.Extra
}

func (b *StateBlock) GetWork() Work {
	return b.Work
}

//go:generate msgp
type SmartContractBlock struct {
	Type           Enum      `msg:"type" json:"type"`
	Address        Address   `msg:"address,extension" json:"address"`
	PreviousHash   Hash      `msg:"previous,extension" json:"previous"`
	Representative Address   `msg:"representative,extension" json:"representative"`
	Balance        Balance   `msg:"balance,extension" json:"balance"`
	Link           Hash      `msg:"link,extension" json:"link"`
	Signature      Signature `msg:"signature,extension" json:"signature"`
	Extra          Hash      `msg:"extra,extension" json:"extra"`
	Work           Work      `msg:"work,extension" json:"work"`
	Owner          Address   `msg:"owner,extension" json:"owner"`
	Issuer         Address   `msg:"issuer,extension" json:"issuer"`
	ExtraAddress   []Address `msg:"extraAddress,extension" json:"extraAddress"`
}

func (sc *SmartContractBlock) GetAddress() Address {
	return sc.Address
}

func (sc *SmartContractBlock) GetExtraAddress() []Address {
	return sc.ExtraAddress
}

func (sc *SmartContractBlock) GetType() Enum {
	return SmartContract
}

func (sc *SmartContractBlock) GetHash() Hash {
	var preamble [preambleSize]byte
	preamble[len(preamble)-1] = byte(SmartContract)
	var a []byte
	for _, addr := range sc.ExtraAddress {
		a = append(a, addr[:]...)
	}
	return hashBytes(preamble[:], a, sc.Address[:], sc.PreviousHash[:], sc.Representative[:], sc.Link[:], sc.Extra[:], sc.Owner[:], sc.Issuer[:])
}

func (sc *SmartContractBlock) GetPreviousHash() Hash {
	return sc.PreviousHash
}

func (sc *SmartContractBlock) GetRepresentative() Address {
	return sc.Representative
}

func (sc *SmartContractBlock) GetBalance() Balance {
	return Balance{}
}

func (sc *SmartContractBlock) GetLink() Hash {
	return sc.Link
}

func (sc *SmartContractBlock) GetSignature() Signature {
	return sc.Signature
}

func (sc *SmartContractBlock) GetToken() Hash {
	return Hash{}
}

func (sc *SmartContractBlock) GetExtra() Hash {
	return sc.Extra
}

func (sc *SmartContractBlock) GetWork() Work {
	return sc.Work
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
	var preamble [preambleSize]byte
	preamble[len(preamble)-1] = byte(State)
	return hashBytes(preamble[:], b.Address[:], b.PreviousHash[:], b.Representative[:], b.Balance.Bytes(binary.BigEndian), b.Link[:], b.Extra[:])
}

func (b *StateBlock) ID() Enum {
	return State
}

func (b *StateBlock) Root() Hash {
	if b.IsOpen() {
		return b.PreviousHash
	}

	return b.Link
}

func (b *StateBlock) Size() int {
	return b.Msgsize()
}

func (b *StateBlock) Valid(threshold uint64) bool {
	if b.IsOpen() {
		return b.Work.IsValid(b.PreviousHash, threshold)
	}

	return b.Work.IsValid(Hash(b.Address), threshold)
}

func (b *StateBlock) IsOpen() bool {
	return !b.PreviousHash.IsZero()
}

func (sc *SmartContractBlock) ID() Enum {
	return SmartContract
}

func (sc *SmartContractBlock) Root() Hash {
	return sc.PreviousHash
}

func (sc *SmartContractBlock) Size() int {
	return sc.Msgsize()
}

func (sc *SmartContractBlock) Valid(threshold uint64) bool {
	panic("implement me")
}
