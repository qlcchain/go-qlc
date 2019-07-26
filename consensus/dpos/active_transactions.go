package dpos

import (
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	confirmReqMaxTimes = 30
	confirmReqInterval = 60
)

type voteKey [types.HashSize]byte

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

func newActiveTrx() *ActiveTrx {
	return &ActiveTrx{
		roots:  new(sync.Map),
		quitCh: make(chan bool, 1),
		perfCh: make(chan *PerformanceTime, 1024000),
	}
}

func (act *ActiveTrx) setDposService(dps *DPoS) {
	act.dps = dps
}

func (act *ActiveTrx) start() {
	timerCheckVotes := time.NewTicker(time.Second)

	for {
		select {
		case <-act.quitCh:
			act.dps.logger.Info("act stopped")
			return
		case perf := <-act.perfCh:
			act.updatePerformanceTime(perf.hash, perf.curTime, perf.confirmed)
		case <-timerCheckVotes.C:
			act.checkVotes()
		}
	}
}

func (act *ActiveTrx) stop() {
	act.quitCh <- true
}

func getVoteKey(block *types.StateBlock) voteKey {
	var key voteKey

	if block.IsOpen() {
		hash, _ := types.HashBytes(block.Address[:], block.Token[:])
		copy(key[:], hash[:])
	} else {
		copy(key[:], block.Previous[:])
	}

	return key
}

func (act *ActiveTrx) addToRoots(block *types.StateBlock) bool {
	vk := getVoteKey(block)

	if _, ok := act.roots.Load(vk); !ok {
		ele, err := newElection(act.dps, block)
		if err != nil {
			act.dps.logger.Errorf("block :%s add to roots error", block.GetHash())
			return false
		}
		act.roots.Store(vk, ele)
		return true
	} else {
		act.dps.logger.Infof("block :%s already exist in roots", block.GetHash())
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

func (act *ActiveTrx) checkVotes() {
	nowTime := time.Now().Unix()

	act.roots.Range(func(key, value interface{}) bool {
		el := value.(*Election)
		if nowTime-el.lastTime < confirmReqInterval {
			return true
		} else {
			el.lastTime = nowTime
		}

		block := el.status.winner
		hash := block.GetHash()

		if !el.confirmed {
			act.dps.logger.Infof("vote:resend confirmReq for block [%s]", hash)
			el.announcements++
			if el.announcements == confirmReqMaxTimes {
				act.roots.Delete(el.vote.id)
				act.dps.deleteBlockCache(block)
				act.dps.rollbackUncheckedFromDb(hash)
			} else {
				act.dps.eb.Publish(common.EventBroadcast, p2p.ConfirmReq, block)
			}
		}

		return true
	})
}

func (act *ActiveTrx) addWinner2Ledger(block *types.StateBlock) {
	hash := block.GetHash()
	act.dps.logger.Debugf("block [%s] acked", hash)

	if exist, err := act.dps.ledger.HasStateBlockConfirmed(hash); !exist && err == nil {
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
			verfier := process.NewLedgerVerifier(act.dps.ledger)
			err = verfier.Rollback(hash)
			if err != nil {
				act.dps.logger.Errorf("error [%s] when rollback hash [%s]", err, hash.String())
			}
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
