/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/hex"
	"testing"
)

func TestBlockWork(t *testing.T) {
	work := Work(0x5bfec9aa51a02630)
	threshold := uint64(0xffffffc000000000)

	hash := mustDecodeHash(t, "f0f3cef7d852c776a5d54ce54856564f523e4c4d59542a585b2e135916a2ba0e")
	if !work.Valid(hash, threshold) {
		t.Errorf("work not valid")
	}

	worker := NewWorker(work, hash, threshold)
	if worker.Generate() != work {
		t.Fatal("work not equal")
	}

	worker.Reset()
	worker.Generate()
}

func mustDecodeHash(t *testing.T, s string) Hash {
	var hash Hash
	bytes, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	copy(hash[:], bytes)
	return hash
}
