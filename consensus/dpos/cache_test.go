package dpos

import (
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestCache_set(t *testing.T) {
	dps := getTestDpos()

	hash := mock.Hash()
	dps.confirmedBlocks.set(hash, nil)

	if !dps.confirmedBlocks.has(hash) {
		t.Fatal()
	}

	addr := mock.Address()
	dps.heartAndVoteInc(hash, addr, onlineKindVote)

	val, err := dps.online.Get(uint64(0))
	if err != nil {
		t.Fatal(err)
	}

	repPeriod := val.(*RepOnlinePeriod)
	if val, ok := repPeriod.Statistic.Load(addr); ok {
		s := val.(*RepAckStatistics)
		if s.VoteCount != 1 {
			t.Fatal()
		}
	} else {
		t.Fatal()
	}
}

func TestCache_expiration(t *testing.T) {
	dps := getTestDpos()
	dps.confirmedBlocks = newCache(1, 3*time.Second)

	hash := mock.Hash()
	dps.confirmedBlocks.set(hash, nil)

	if !dps.confirmedBlocks.has(hash) {
		t.Fatal()
	}

	time.Sleep(3 * time.Second)

	if dps.confirmedBlocks.has(hash) {
		t.Fatal()
	}
}

func TestCache_len(t *testing.T) {
	dps := getTestDpos()
	discardHash := types.ZeroHash

	dps.confirmedBlocks = newCache(1, 3*time.Minute)
	dps.confirmedBlocks.evictedFunc = func(k interface{}, v interface{}) {
		discardHash = k.(types.Hash)
	}

	hash1 := mock.Hash()
	dps.confirmedBlocks.set(hash1, nil)

	if dps.confirmedBlocks.len() != 1 {
		t.Fatal()
	}

	hash2 := mock.Hash()
	dps.confirmedBlocks.set(hash2, nil)

	if discardHash != hash1 {
		t.Fatal()
	}

	if dps.confirmedBlocks.len() != 1 {
		t.Fatal()
	}
}

func TestCache_gc(t *testing.T) {
	dps := getTestDpos()
	dps.confirmedBlocks = newCache(3, 3*time.Second)

	hash1 := mock.Hash()
	dps.confirmedBlocks.set(hash1, nil)

	hash2 := mock.Hash()
	dps.confirmedBlocks.set(hash2, nil)

	hash3 := mock.Hash()
	dps.confirmedBlocks.set(hash3, nil)

	time.Sleep(3 * time.Second)

	dps.confirmedBlocks.gc()

	if dps.confirmedBlocks.len() != 0 {
		t.Fatal()
	}
}
