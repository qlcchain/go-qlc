/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/tinylib/msgp/msgp"
)

func init() {
	msgp.RegisterExtension(AbiExtensionType, func() msgp.Extension { return new(ContractAbi) })
}

//go:generate msgp
type ContractAbi struct {
	Abi       []byte `msg:"abi" json:"abi"`
	AbiLength uint64 `msg:"abiLength" json:"abiLength"`
	AbiHash   Hash   `msg:"abiHash,extension" json:"abiHash"`
}

//ExtensionType implements Extension.ExtensionType interface
func (c *ContractAbi) ExtensionType() int8 {
	return AbiExtensionType
}

//ExtensionType implements Extension.Len interface
func (c *ContractAbi) Len() int {
	return len(c.Abi) + HashSize + 8
}

//ExtensionType implements Extension.MarshalBinaryTo interface
func (c *ContractAbi) MarshalBinaryTo(text []byte) error {
	var bytes [8]byte
	binary.BigEndian.PutUint64(bytes[:], c.AbiLength)
	copy(text, bytes[:])
	copy(text, c.AbiHash[:])
	copy(text, c.Abi)
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (c *ContractAbi) UnmarshalBinary(text []byte) error {
	c.AbiLength = binary.BigEndian.Uint64(text[:8])
	var err error
	c.AbiHash, err = NewHash(hex.EncodeToString(text[8 : 8+HashSize]))
	if err != nil {
		return err
	}

	copy(c.Abi, text[8+HashSize:])

	return nil
}
