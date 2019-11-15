package dpos

import (
	"github.com/qlcchain/go-qlc/mock"
	"testing"
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
	if s, ok := repPeriod.Statistic[addr]; ok {
		if s.VoteCount != 1 {
			t.Fatal()
		}
	} else {
		t.Fatal()
	}
}
