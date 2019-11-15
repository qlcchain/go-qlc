package dpos

import (
	"github.com/qlcchain/go-qlc/common"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestOnGetOnlineInfo(t *testing.T) {
	dps := getTestDpos()

	rep := &RepOnlinePeriod{
		Period:     1,
		Statistic:  make(map[types.Address]*RepAckStatistics),
		BlockCount: 2,
		lock:       nil,
	}

	ack := &RepAckStatistics{
		HeartCount:      10,
		LastHeartHeight: 0,
		VoteCount:       20,
	}
	addr := mock.Address()
	rep.Statistic[addr] = ack

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
	if s, ok := repg.Statistic[addr]; ok {
		if s.VoteCount != 20 || s.HeartCount != 10 {
			t.Fatal()
		}
	} else {
		t.Fatal()
	}
}
