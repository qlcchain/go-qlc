/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"testing"
)

func TestBlockWork(t *testing.T) {
	work := Work(0x880ab6aa90a59d5d)
	threshold := uint64(0xfffffe0000000000)
	var hash Hash
	err := hash.Of("2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E")

	if err != nil {
		t.Errorf("hash not valid")
	}

	if !work.IsValid(hash, threshold) {
		t.Errorf("work not valid")
	}

	worker, err := NewWorker(work, hash, threshold)
	if err != nil {
		t.Fatal("NewWorker failed.")
	}

	v := worker.NewWork()
	if v != work {
		t.Fatalf("work not equal, expect:%s but %s", work.String(), v.String())
	}
}
