/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"bytes"
	"encoding/json"
	"errors" //"encoding/json"
	"fmt"
	"strings"
)

var (
	ErrBadBlockType = errors.New("bad block type")
	ErrNotABlock    = errors.New("block type is not_a_block")
)

type Block interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
	fmt.Stringer
}

//msgp:shim BlockType as:string using:(BlockType).String/parseString
type BlockType byte

const (
	State BlockType = iota
	Send
	Receive
	Change
	Open
	ContractReward
	ContractSend
	ContractRefund
	ContractError
	SmartContract
	Invalid
)

func parseString(s string) BlockType {
	switch strings.ToLower(s) {
	case "state":
		return State
	case "send":
		return Send
	case "receive":
		return Receive
	case "change":
		return Change
	case "open":
		return Open
	case "contractReward":
		return ContractReward
	case "contractSend":
		return ContractSend
	case "contractRefund":
		return ContractRefund
	case "contractError":
		return ContractError
	case "smartcontract":
		return SmartContract
	default:
		return Invalid
	}
}

func (t BlockType) String() string {
	switch t {
	case State:
		return "State"
	case Send:
		return "Send"
	case Receive:
		return "Receive"
	case Change:
		return "Change"
	case Open:
		return "Open"
	case ContractReward:
		return "ContractReward"
	case ContractSend:
		return "ContractSend"
	case ContractRefund:
		return "ContractRefund"
	case ContractError:
		return "ContractError"
	case SmartContract:
		return "SmartContract"
	default:
		return "<invalid>"
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

func (e BlockType) Equal(t BlockType) bool {
	return byte(e) == byte(t)
}

func NewBlock(t BlockType) (Block, error) {
	switch t {
	case State, ContractSend, ContractRefund, ContractReward, ContractError, Send, Receive, Change, Open:
		sb := new(StateBlock)
		sb.Type = t
		return sb, nil
	case SmartContract:
		sc := new(SmartContractBlock)
		return sc, nil
	case Invalid:
		return nil, ErrNotABlock
	default:
		return nil, ErrBadBlockType
	}
}
