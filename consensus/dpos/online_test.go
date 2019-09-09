package dpos

import (
	"github.com/prometheus/common/log"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
	"testing"
	"time"
)

func TestOnGetOnlineInfo(t *testing.T) {
	rep := &RepOnlinePeriod{
		Period:     1,
		Statistic:  make(map[types.Address]*RepAckStatistics),
		BlockCount: 2,
		lock:       nil,
	}

	ack := &RepAckStatistics{
		HeartCount:    10,
		LastHeartTime: time.Time{},
		VoteCount:     20,
	}
	rep.Statistic[mock.Address()] = ack

	log.Infof("%s", rep)
}
