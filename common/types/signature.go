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

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/tinylib/msgp/msgp"
)

func init() {
	msgp.RegisterExtension(SignatureExtensionType, func() msgp.Extension { return new(Signature) })
}

const (
	//SignatureSize size of signature
	SignatureSize = ed25519.SignatureSize
)

var ZeroSignature = Signature{}

// Signature of block
type Signature [SignatureSize]byte

func BytesToSignature(b []byte) (Signature, error) {
	var sign Signature
	if len(b) != ed25519.SignatureSize {
		return ZeroSignature, fmt.Errorf("invalid signature size[%d]", len(b))
	}
	copy(sign[:], b)

	return sign, nil
}

// String implements the fmt.Stringer interface.
func (s Signature) String() string {
	return hex.EncodeToString(s[:])
}

//Of convert hex string to Signature
func (s *Signature) Of(hexString string) error {
	ss := util.TrimQuotes(hexString)
	size := hex.DecodedLen(len(ss))
	if size != SignatureSize {
		return fmt.Errorf("bad signature size: %d", size)
	}

	var signature [SignatureSize]byte
	if _, err := hex.Decode(signature[:], []byte(ss)); err != nil {
		return err
	}

	*s = signature
	return nil
}

//ExtensionType implements Extension.ExtensionType interface
func (s *Signature) ExtensionType() int8 {
	return SignatureExtensionType
}

//ExtensionType implements Extension.Len interface
func (s *Signature) Len() int {
	return SignatureSize
}

//ExtensionType implements Extension.MarshalBinaryTo interface
func (s Signature) MarshalBinaryTo(text []byte) error {
	copy(text, s[:])
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (s *Signature) UnmarshalBinary(text []byte) error {
	size := len(text)
	if len(text) != SignatureSize {
		return fmt.Errorf("bad signature size: %d", size)
	}
	copy((*s)[:], text)
	return nil
}

////MarshalJSON implements json.Marshaler interface
//func (s Signature) MarshalJSON() ([]byte, error) {
//	return []byte(s.String()), nil
//}
//
//func (s *Signature) UnmarshalJSON(b []byte) error {
//	return s.Of(string(b))
//}

func (s *Signature) UnmarshalText(text []byte) error {
	return s.Of(string(text))
}

// MarshalText implements the encoding.TextMarshaler interface.
func (s Signature) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}
