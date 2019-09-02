/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"gitlab.com/samli88/go-x11-hash"
	"golang.org/x/crypto/scrypt"
	"math/big"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/tinylib/msgp/msgp"
	"golang.org/x/crypto/blake2b"
)

func init() {
	msgp.RegisterExtension(HashExtensionType, func() msgp.Extension { return new(Hash) })
}

const (
	//HashSize size of hash
	HashSize = blake2b.Size256
)

var ZeroHash = Hash{}
var FFFFHash, _ = NewHash("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

//Hash blake2b hash
//go:generate msgp
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

func (h Hash) ReverseString() string {
	for i := 0; i < HashSize/2; i++ {
		h[i], h[HashSize-1-i] = h[HashSize-1-i], h[i]
	}
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

func (h Hash) Cmp(h2 Hash) int {
	lhs := h[:]
	rhs := h2[:]

	for i := 0; i < HashSize; i++ {
		if lhs[i] > rhs[i] {
			return 1
		} else if lhs[i] < rhs[i] {
			return -1
		}
	}

	return 0
}

func (h *Hash) ToBigInt() *big.Int {
	return HashToBig(h)
}

func (h *Hash) FromBigInt(num *big.Int) error {
	h0 := BigToHash(num)
	if h0 != nil {
		*h = *h0
	}

	return nil
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

func Sha256D_HashData(data []byte) Hash {
	h, _ := Sha256D_HashBytes(data)
	return h
}

//Sha256D_HashBytes hash data by sha256
func Sha256D_HashBytes(inputs ...[]byte) (Hash, error) {
	hash := sha256.New()

	for _, data := range inputs {
		hash.Write(data)
	}
	first := hash.Sum(nil)

	hash.Reset()
	hash.Write(first[:])

	second := hash.Sum(nil)

	var result Hash
	copy(result[:], second)
	return result, nil
}

func Scrypt_HashData(data []byte) Hash {
	scryptHash, err := scrypt.Key(data, data, 1024, 1, 1, 32)
	if err != nil {
		return ZeroHash
	}

	var result Hash
	copy(result[:], scryptHash)
	return result
}

func Scrypt_HashBytes(inputs ...[]byte) (Hash, error) {
	buf := new(bytes.Buffer)

	for _, data := range inputs {
		buf.Write(data)
	}

	scryptHash, err := scrypt.Key(buf.Bytes(), buf.Bytes(), 1024, 1, 1, 32)
	if err != nil {
		return ZeroHash, err
	}

	var result Hash
	copy(result[:], scryptHash)
	return result, nil
}

func X11_HashData(data []byte) Hash {
	out := make([]byte, 32)
	hs := x11.New()
	hs.Hash(data, out)

	var result Hash
	copy(result[:], out)
	return result
}
