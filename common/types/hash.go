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
	"strings"

	"github.com/qlcchain/go-qlc/common/types/internal/util"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
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
		return fmt.Errorf("bad block hash size: %d", size)
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

//UnmarshalJSON implements json.UnmarshalJSON interface
func (h *Hash) UnmarshalJSON(b []byte) error {
	err := h.Of(string(b))
	return err
}

// Sign hash by privateKey
func (h *Hash) Sign(privateKey ed25519.PrivateKey) (Signature, error) {
	sig := hex.EncodeToString(ed25519.Sign(privateKey, h[:]))
	s := Signature{}
	err := s.Of(strings.ToUpper(sig))
	if err != nil {
		return s, err
	}
	return s, nil
}

//func (h *Hash) UnmarshalText(text []byte) error {
//	s := util.TrimQuotes(string(text))
//	err := h.Of(s)
//	return err
//}
//
//func (h Hash) MarshalText() (text []byte, err error) {
//	//fmt.Println(h.String())
//	return []byte(h.String()), nil
//}
