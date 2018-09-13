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
	"golang.org/x/crypto/blake2b"
)

const (
	// SeedSize size of the seed
	SeedSize = 32
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

// From convert hex seed string to Seed
func (s *Seed) From(hexString string) error {
	b, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}

	if len(b) != SeedSize {
		return fmt.Errorf("invalid seed size, expect %d but %d", SeedSize, len(b))

	}
	copy(s[:], b)
	return nil
}
