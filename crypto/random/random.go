/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package random

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	mathRand "math/rand"
)

type cryptoRandSource struct {
	data [8]byte
	err  error
}

func (s *cryptoRandSource) Int63() int64 {
	if _, err := cryptoRand.Read(s.data[:]); err != nil {
		s.err = err
		return 0
	}
	return int64(binary.BigEndian.Uint64(s.data[:]) & (1<<63 - 1))
}

func (s *cryptoRandSource) Seed(seed int64) {
	panic("the seed function is not supported")
}

// Bytes fills the given byte slice with random bytes.
func Bytes(data []byte) error {
	_, err := cryptoRand.Read(data)
	return err
}

// Perm returns a random permutation of the integers [0,n).
func Perm(n int) ([]int, error) {
	var src cryptoRandSource
	i := mathRand.New(&src).Perm(n)
	if src.err != nil {
		return nil, src.err
	}
	return i, nil
}

// Intn returns a random number in [0,n).
func Intn(n int) (int, error) {
	var src cryptoRandSource
	i := mathRand.New(&src).Intn(n)
	if src.err != nil {
		return 0, src.err
	}
	return i, nil
}
