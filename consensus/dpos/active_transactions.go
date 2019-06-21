package dpos

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/p2p"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	announcementMax  = 40
	announceInterval = 60
)

type voteKey [1 + types.HashSize]byte

type ActiveTrx struct {
	confirmed electionStatus
	dps       *DPoS
	roots     *sync.Map
	quitCh    chan bool
	perfCh    chan *PerformanceTime
}

type PerformanceTime struct {
	hash      types.Hash
	curTime   int64
	confirmed bool
}

func NewActiveTrx() *ActiveTrx {
	return &ActiveTrx{
		roots:  new(sync.Map),
		quitCh: make(chan bool, 1),
		perfCh: make(chan *PerformanceTime, 1024000),
	}
}

func (act *ActiveTrx) SetDposService(dps *DPoS) {
	act.dps = dps
}

func (act *ActiveTrx) start() {
	timerAnnounce := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-act.quitCh:
			act.dps.logger.Info("act stopped")
			return
		case perf := <-act.perfCh:
			act.updatePerformanceTime(perf.hash, perf.curTime, perf.confirmed)
		case <-timerAnnounce.C:
			act.announceVotes()
		}
	}
}

func (act *ActiveTrx) stop() {
	act.quitCh <- true
}

func getVoteKey(block *types.StateBlock) voteKey {
	var key voteKey

	if block.IsOpen() {
		key[0] = 1
		copy(key[1:], block.Link[:])
	} else {
		key[0] = 0
		copy(key[1:], block.Previous[:])
	}

	return key
}

func (act *ActiveTrx) addToRoots(block *types.StateBlock) bool {
	vk := getVoteKey(block)

	if _, ok := act.roots.Load(vk); !ok {
		ele, err := NewElection(act.dps, block)
		if err != nil {
			act.dps.logger.Errorf("block :%s add to roots error", block.GetHash())
			return false
		}
		act.roots.Store(vk, ele)
		return true
	} else {
		act.dps.logger.Infof("block :%s already exit in roots", block.GetHash())
		return false
	}
}

func (act *ActiveTrx) updatePerfTime(hash types.Hash, curTime int64, confirmed bool) {
	if !act.dps.cfg.PerformanceEnabled {
		return
	}

	perf := &PerformanceTime{
		hash:      hash,
		curTime:   curTime,
		confirmed: confirmed,
	}

	act.perfCh <- perf
}

func (act *ActiveTrx) updatePerformanceTime(hash types.Hash, curTime int64, confirmed bool) {
	if confirmed {
		if p, err := act.dps.ledger.GetPerformanceTime(hash); p != nil && err == nil {
			t := &types.PerformanceTime{
				Hash: hash,
				T0:   p.T0,
				T1:   curTime,
				T2:   p.T2,
				T3:   p.T3,
			}

			err := act.dps.ledger.AddOrUpdatePerformance(t)
			if err != nil {
				act.dps.logger.Info("AddOrUpdatePerformance error T1")
			}
		} else {
			act.dps.logger.Info("get performanceTime error T1")
		}
	} else {
		t := &types.PerformanceTime{
			Hash: hash,
			T0:   curTime,
			T1:   0,
			T2:   0,
			T3:   0,
		}

		err := act.dps.ledger.AddOrUpdatePerformance(t)
		if err != nil {
			act.dps.logger.Infof("AddOrUpdatePerformance error T0")
		}
	}
}

func (act *ActiveTrx) announceVotes() {
	nowTime := time.Now().Unix()

	act.roots.Range(func(key, value interface{}) bool {
		el := value.(*Election)
		if nowTime-el.lastTime < announceInterval {
			return true
		} else {
			el.lastTime = nowTime
		}

		block := el.status.winner
		hash := block.GetHash()

		if !el.confirmed {
			act.dps.logger.Infof("vote:send confirmReq for block [%s]", hash)
			el.announcements++
			if el.announcements == announcementMax {
				act.roots.Delete(el.vote.id)
			} else {
				act.dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmReq, block)
			}
		}

		return true
	})
}

func (act *ActiveTrx) addWinner2Ledger(block *types.StateBlock) {
	hash := block.GetHash()

	if exist, err := act.dps.ledger.HasStateBlock(hash); !exist && err == nil {
		err := act.dps.lv.BlockProcess(block)
		if err != nil {
			act.dps.logger.Error(err)
		} else {
			act.dps.logger.Debugf("save block[%s]", hash.String())
		}
	} else {
		act.dps.logger.Debugf("%s, %v", hash.String(), err)
	}
}

func (act *ActiveTrx) rollBack(blocks []*types.StateBlock) {
	for _, v := range blocks {
		hash := v.GetHash()
		act.dps.logger.Info("loser hash is :", hash.String())
		h, err := act.dps.ledger.HasStateBlock(hash)
		if err != nil {
			act.dps.logger.Errorf("error [%s] when run HasStateBlock func ", err)
			continue
		}
		if h {
			err = act.dps.ledger.Rollback(hash)
			if err != nil {
				act.dps.logger.Errorf("error [%s] when rollback hash [%s]", err, hash.String())
			}
			act.dps.rollbackUnchecked(hash)
		}
	}
}

func (act *ActiveTrx) vote(va *protos.ConfirmAckBlock) {
	vk := getVoteKey(va.Blk)

	if v, ok := act.roots.Load(vk); ok {
		v.(*Election).voteAction(va)
	}
}

func (act *ActiveTrx) isVoting(block *types.StateBlock) bool {
	vk := getVoteKey(block)

	if _, ok := act.roots.Load(vk); ok {
		return true
	}

	return false
}
