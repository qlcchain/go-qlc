/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/tinylib/msgp/msgp"
	"golang.org/x/crypto/blake2b"
)

func init() {
	msgp.RegisterExtension(SeedExtensionType, func() msgp.Extension { return new(Seed) })
}

const (
	// SeedSize size of the seed
	SeedSize          = 32
	SeedExtensionType = 102
)

//Seed of account
type Seed [SeedSize]byte

//NewSeed generate new seed
func NewSeed() (*Seed, error) {
	seed := new(Seed)
	if err := random.Bytes(seed[:]); err != nil {
		return nil, err
	}

	return seed, nil
}

//Key get private key by index from seed
func (s *Seed) Key(index uint32) (ed25519.PrivateKey, error) {
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, index)

	hash, err := blake2b.New(blake2b.Size256, nil)
	if err != nil {
		panic(err)
	}
	hash.Write(s[:])
	hash.Write(indexBytes)

	sum := hash.Sum(nil)
	sumReader := bytes.NewReader(sum)
	_, key, err := ed25519.GenerateKey(sumReader)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// String seed tostring
func (s *Seed) String() string {
	return hex.EncodeToString(s[:])
}

//ExtensionType implements Extension.ExtensionType interface
func (s *Seed) ExtensionType() int8 {
	return SeedExtensionType
}

//ExtensionType implements Extension.Len interface
func (s *Seed) Len() int {
	return AddressSize
}

//ExtensionType implements Extension.MarshalBinaryTo interface
func (s *Seed) MarshalBinaryTo(text []byte) error {
	copy(text, (*s)[:])
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (s *Seed) UnmarshalBinary(text []byte) error {
	size := len(text)
	if len(text) != SeedSize {
		return fmt.Errorf("bad signature size: %d", size)
	}
	copy((*s)[:], text)
	return nil
}

//MarshalJSON implements json.Marshaler interface
func (s *Seed) MarshalJSON() ([]byte, error) {
	return []byte(s.String()), nil
}
