package dpos

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestOnGetOnlineInfo(t *testing.T) {
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
	rep.Statistic[mock.Address()] = ack

	t.Logf("%s", rep)
}
