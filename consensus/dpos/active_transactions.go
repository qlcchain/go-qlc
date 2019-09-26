package dpos

import (
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p"
)

const (
	confirmReqMaxTimes = 30
	confirmReqInterval = 60
)

type voteKey [types.HashSize]byte

type ActiveTrx struct {
	dps    *DPoS
	roots  *sync.Map
	quitCh chan bool
	perfCh chan *PerformanceTime
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
		el := newElection(act.dps, block)
		act.roots.Store(vk, el)
		return true
	} else {
		act.dps.logger.Debugf("block :%s already exist in roots", block.GetHash())
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
	dps := act.dps

	act.roots.Range(func(key, value interface{}) bool {
		el := value.(*Election)
		if nowTime-el.lastTime < confirmReqInterval {
			return true
		} else {
			el.lastTime = nowTime
		}

		block := el.status.winner
		hash := block.GetHash()

		el.announcements++
		if el.announcements == confirmReqMaxTimes || el.status.winner.Type == types.Online {
			if !el.ifValidAndSetInvalid() {
				return true
			}

			dps.logger.Infof("block[%s] was not confirmed after 30 times resend", hash)
			act.roots.Delete(el.vote.id)
			_ = dps.lv.Rollback(hash)
			el.cleanBlockInfo()
			dps.frontiersStatus.Delete(hash)
			dps.syncBlockRollback(hash)
		} else {
			dps.logger.Infof("resend confirmReq for block[%s]", hash)
			confirmReqBlocks := make([]*types.StateBlock, 0)
			confirmReqBlocks = append(confirmReqBlocks, block)
			dps.eb.Publish(common.EventBroadcast, p2p.ConfirmReq, confirmReqBlocks)
		}

		return true
	})
}

func (act *ActiveTrx) addWinner2Ledger(block *types.StateBlock) {
	hash := block.GetHash()
	dps := act.dps
	dps.logger.Debugf("block[%s] confirmed", hash)

	if exist, err := dps.ledger.HasStateBlockConfirmed(hash); !exist && err == nil {
		err := dps.lv.BlockProcess(block)
		if err != nil {
			dps.logger.Error(err)
		} else {
			dps.logger.Debugf("save block[%s]", hash)
		}
	} else {
		dps.logger.Debugf("%s, %v", hash, err)
	}
}

func (act *ActiveTrx) addSyncBlock2Ledger(block *types.StateBlock) {
	hash := block.GetHash()
	dps := act.dps
	dps.logger.Debugf("sync block[%s] confirmed", hash)

	if exist, err := dps.ledger.HasStateBlockConfirmed(hash); !exist && err == nil {
		err := dps.lv.BlockSyncProcess(block)
		if err != nil {
			dps.logger.Error(err)
		} else {
			dps.logger.Debugf("save block[%s]", hash)
		}
	} else {
		dps.logger.Debugf("%s, %v", hash, err)
	}

	dps.chainFinished(hash)
}

func (act *ActiveTrx) rollBack(blocks []*types.StateBlock) {
	dps := act.dps

	for _, v := range blocks {
		hash := v.GetHash()
		dps.logger.Info("loser hash is:", hash)

		has, err := dps.ledger.HasStateBlock(hash)
		if err != nil {
			dps.logger.Errorf("error [%s] when run HasStateBlock func ", err)
			continue
		}

		if has {
			err = dps.lv.Rollback(hash)
			if err != nil {
				dps.logger.Errorf("error [%s] when rollback hash [%s]", err, hash)
			}
		}
	}
}

func (act *ActiveTrx) vote(vi *voteInfo) {
	if v, ok := act.dps.hash2el.Load(vi.hash); ok {
		el := v.(*Election)
		el.voteAction(vi)
	}
}

func (act *ActiveTrx) voteFrontier(vi *voteInfo) (confirmed bool) {
	if v, ok := act.dps.hash2el.Load(vi.hash); ok {
		el := v.(*Election)
		return el.voteFrontier(vi)
	}
	return false
}

func (act *ActiveTrx) isVoting(block *types.StateBlock) bool {
	vk := getVoteKey(block)

	if _, ok := act.roots.Load(vk); ok {
		return true
	}

	return false
}

func (act *ActiveTrx) getVoteInfo(block *types.StateBlock) *Election {
	vk := getVoteKey(block)

	if v, ok := act.roots.Load(vk); ok {
		return v.(*Election)
	}

	return nil
}

func (act *ActiveTrx) setVoteHash(block *types.StateBlock) {
	vk := getVoteKey(block)

	if v, ok := act.roots.Load(vk); ok {
		v.(*Election).voteHash = block.GetHash()
	}
}
