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

	"golang.org/x/crypto/blake2b"
)

const (
	//HashSize size of hash
	HashSize = blake2b.Size256
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

// MarshalText implements the encoding.TextMarshaler interface.
func (h Hash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (h *Hash) UnmarshalText(text []byte) error {
	size := hex.DecodedLen(len(text))
	if size != HashSize {
		return fmt.Errorf("bad block hash size: %d", size)
	}

	var hash [HashSize]byte
	if _, err := hex.Decode(hash[:], text); err != nil {
		return err
	}

	*h = hash
	return nil
}

// String implements the fmt.Stringer interface.
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

//Of convert hex string to Hash
func (h *Hash) Of(hexString string) error {
	return h.UnmarshalText([]byte(hexString))
}
