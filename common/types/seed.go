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
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/crypto/random"
	"golang.org/x/crypto/blake2b"
)

const (
	SeedSize = 32
)

type Seed [SeedSize]byte

func GenerateSeed() (*Seed, error) {
	seed := new(Seed)
	if err := random.Bytes(seed[:]); err != nil {
		return nil, err
	}

	return seed, nil
}

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

func (s *Seed) String() string {
	return hex.EncodeToString(s[:])
}
