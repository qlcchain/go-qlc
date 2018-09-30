/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/hex"
	"fmt"

	"github.com/tinylib/msgp/msgp"
	"golang.org/x/crypto/blake2b"
)

func init() {
	msgp.RegisterExtension(HashExtensionType, func() msgp.Extension { return new(Hash) })
}

const (
	//HashSize size of hash
	HashSize          = blake2b.Size256
	HashExtensionType = 101
)

//Hash blake2b hash
type Hash [HashSize]byte

//IsZero check hash is zero
func (h Hash) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}

// String implements the fmt.Stringer interface.
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

//Of convert hex string to Hash
func (h *Hash) Of(hexString string) error {
	size := hex.DecodedLen(len(hexString))
	if size != HashSize {
		return fmt.Errorf("bad block hash size: %d", size)
	}

	var hash [HashSize]byte
	if _, err := hex.Decode(hash[:], []byte(hexString)); err != nil {
		return err
	}

	*h = hash
	return nil
}

//ExtensionType implements Extension.ExtensionType interface
func (h *Hash) ExtensionType() int8 {
	return HashExtensionType
}

//ExtensionType implements Extension.Len interface
func (h *Hash) Len() int {
	return HashSize
}

//ExtensionType implements Extension.MarshalBinaryTo interface
func (h *Hash) MarshalBinaryTo(text []byte) error {
	copy(text, (*h)[:])
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (h *Hash) UnmarshalBinary(text []byte) error {
	size := len(text)
	if len(text) != HashSize {
		return fmt.Errorf("bad signature size: %d", size)
	}
	copy((*h)[:], text)
	return nil
}

//MarshalJSON implements json.Marshaler interface
func (h *Hash) MarshalJSON() ([]byte, error) {
	return []byte(h.String()), nil
}
