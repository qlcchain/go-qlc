package dpos

import (
	"sync"
	"testing"

	"github.com/qlcchain/go-qlc/common"

	"github.com/qlcchain/go-qlc/mock"
)

func TestGetOnlineInfo(t *testing.T) {
	dps := getTestDpos()

	rep := &RepOnlinePeriod{
		Period:     1,
		Statistic:  new(sync.Map),
		BlockCount: 2,
	}

	ack := &RepAckStatistics{
		HeartCount:      10,
		LastHeartHeight: 0,
		VoteCount:       20,
	}
	addr := mock.Address()
	rep.Statistic.Store(addr, ack)

	period := dps.curPovHeight / common.DPosOnlinePeriod
	err := dps.online.Set(period, rep)
	if err != nil {
		t.Fatal()
	}

	val, err := dps.online.Get(period)
	if err != nil {
		t.Fatal()
	}

	repg := val.(*RepOnlinePeriod)
	if val, ok := repg.Statistic.Load(addr); ok {
		s := val.(*RepAckStatistics)
		if s.VoteCount != 20 || s.HeartCount != 10 {
			t.Fatal()
		}
	} else {
		t.Fatal()
	}
}

func TestVoteHistoryUpdate(t *testing.T) {
	dps := getTestDpos()

	hash := mock.Hash()
	addr := mock.Address()

	if dps.isValidVote(hash, addr) {
		t.Fatal()
	}

	dps.confirmedBlockInc(hash)
	if !dps.isValidVote(hash, addr) {
		t.Fatal()
	}

	if dps.isValidVote(hash, addr) {
		t.Fatal()
	}
}
