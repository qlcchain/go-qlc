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
	"math/big"

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

func NewSignature(hexStr string) (Signature, error) {
	s := Signature{}
	err := s.Of(hexStr)
	if err != nil {
		return s, err
	}
	return s, nil
}

func BytesToSignature(data []byte) (Signature, error) {
	if len(data) != SignatureSize {
		return ZeroSignature, errors.New("invalid Signature size")
	}

	var signature [SignatureSize]byte
	copy(signature[:], data)
	return signature, nil
}

//IsZero check signature is zero
func (s *Signature) IsZero() bool {
	for _, b := range s {
		if b != 0 {
			return false
		}
	}
	return true
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

func (s Signature) Bytes() []byte {
	return s[:]
}

// ToBigInt converts a types.Signature into a big.Int that can be used to
// perform math comparisons.
func (s *Signature) ToBigInt() *big.Int {
	// A Signature is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	var sigBuf [SignatureSize]byte
	copy(sigBuf[:], s[:])

	buf := sigBuf
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	//fmt.Println("ToBigInt", hex.EncodeToString(s[:]), hex.EncodeToString(sigBuf[:]), hex.EncodeToString(buf[:]))

	return new(big.Int).SetBytes(buf[:])
}

// FromBigInt converts a big.Int into a types.Signature.
func (s *Signature) FromBigInt(num *big.Int) error {
	// A big.Int is in big-endian, but a Signature is in little-endian, so reverse them.
	numBuf := num.Bytes()
	if len(numBuf) > SignatureSize {
		return fmt.Errorf("bad big.Int bytes size: %d", len(numBuf))
	}

	var sigBuf [SignatureSize]byte

	startPos := SignatureSize - len(numBuf)
	copy(sigBuf[startPos:], numBuf)

	buf := sigBuf
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	*s = buf

	return nil
}