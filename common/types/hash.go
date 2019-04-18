/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/util"
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

var ZeroHash = Hash{}

//Hash blake2b hash
type Hash [HashSize]byte

func NewHash(hexStr string) (Hash, error) {
	h := Hash{}
	err := h.Of(hexStr)
	if err != nil {
		return h, err
	}
	return h, nil
}

func BytesToHash(data []byte) (Hash, error) {
	if len(data) != HashSize {
		return ZeroHash, errors.New("invalid Hash size")
	}

	var hash [HashSize]byte
	copy(hash[:], data)
	return hash, nil
}

//IsZero check hash is zero
func (h *Hash) IsZero() bool {
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
	s := util.TrimQuotes(hexString)
	size := hex.DecodedLen(len(s))
	if size != HashSize {
		return fmt.Errorf("hash convert, bad hash size: %d", size)
	}

	var hash [HashSize]byte
	if _, err := hex.Decode(hash[:], []byte(s)); err != nil {
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
func (h Hash) MarshalBinaryTo(text []byte) error {
	copy(text, h[:])
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (h *Hash) UnmarshalBinary(text []byte) error {
	size := len(text)
	if len(text) != HashSize {
		return fmt.Errorf("hash unmarshal, bad hash size: %d", size)
	}
	copy((*h)[:], text)
	return nil
}

//UnmarshalText implements encoding.TextUnmarshaler
func (h *Hash) UnmarshalText(text []byte) error {
	s := util.TrimQuotes(string(text))
	err := h.Of(s)
	return err
}

// MarshalText implements the encoding.TextMarshaler interface.
func (h Hash) MarshalText() (text []byte, err error) {
	return []byte(h.String()), nil
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func HashData(data []byte) Hash {
	h, _ := HashBytes(data)
	return h
}

//HashBytes hash data by blake2b
func HashBytes(inputs ...[]byte) (Hash, error) {
	hash, err := blake2b.New(blake2b.Size256, nil)
	if err != nil {
		return ZeroHash, err
	}

	for _, data := range inputs {
		hash.Write(data)
	}

	var result Hash
	copy(result[:], hash.Sum(nil))
	return result, nil
}
