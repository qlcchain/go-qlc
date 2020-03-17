package dpos

import (
	"github.com/qlcchain/go-qlc/mock"
	"testing"
	"time"
)

func TestActX_cleanFrontierVotes(t *testing.T) {
	dps := getTestDpos()

	blk := mock.StateBlockWithoutWork()
	dps.frontiersStatus.Store(blk.GetHash(), frontierChainFinished)
	dps.acTrx.addToRoots(blk)

	dps.acTrx.cleanFrontierVotes()
	if _, ok := dps.hash2el.Load(blk.GetHash()); ok {
		t.Fatal()
	}
}

func TestActX_checkVotes(t *testing.T) {
	dps := getTestDpos()

	blk := mock.StateBlockWithoutWork()
	dps.frontiersStatus.Store(blk.GetHash(), frontierChainFinished)
	el := newElection(dps, blk)
	el.lastTime = time.Now().Unix() - 180
	vk := getVoteKey(blk)
	dps.acTrx.roots.Store(vk, el)

	dps.acTrx.checkVotes()
	if _, ok := dps.acTrx.roots.Load(vk); ok {
		t.Fatal()
	}
}

func TestActX_isVoting(t *testing.T) {
	dps := getTestDpos()

	blk := mock.StateBlockWithoutWork()
	dps.acTrx.addToRoots(blk)

	if !dps.acTrx.isVoting(blk) {
		t.Fatal()
	}
}
