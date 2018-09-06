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
	"github.com/qlcchain/go-qlc/crypto/ed25519"
)

const (
	SignatureSize = ed25519.SignatureSize
)

type Signature [SignatureSize]byte

// MarshalText implements the encoding.TextMarshaler interface.
func (s Signature) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (s *Signature) UnmarshalText(text []byte) error {
	size := hex.DecodedLen(len(text))
	if size != SignatureSize {
		return fmt.Errorf("bad signature size: %d", size)
	}

	var signature [SignatureSize]byte
	if _, err := hex.Decode(signature[:], text); err != nil {
		return err
	}

	*s = signature
	return nil
}

// String implements the fmt.Stringer interface.
func (s Signature) String() string {
	return hex.EncodeToString(s[:])
}
