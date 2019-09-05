package dpos

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/p2p"
	"sync"
	"time"
)

type onlineKind byte

const (
	onlineKindHeart onlineKind = iota
	onlineKindVote
)

const (
	heartCountPerPeriod = 60
)

type repAckStatistics struct {
	heartCount uint64
	lastHeartTime	time.Time
	voteCount  uint64
}

type repOnlinePeriod struct {
	statistic  map[types.Address]*repAckStatistics
	blockCount uint64
	lock       *sync.RWMutex
}

func (dps *DPoS) heartAndVoteInc(hash types.Hash, addr types.Address, kind onlineKind) {
	period := (dps.curPovHeight - 1) / onlinePeriod
	now := time.Now()
	var repPeriod *repOnlinePeriod

	if s, err := dps.online.Get(period); err == nil {
		repPeriod = s.(*repOnlinePeriod)
	} else {
		repPeriod = &repOnlinePeriod{
			statistic:  make(map[types.Address]*repAckStatistics),
			blockCount: 0,
			lock:       new(sync.RWMutex),
		}

		err := dps.online.Set(period, repPeriod)
		if err != nil {
			dps.logger.Error("set online period err", err)
		}
	}

	repPeriod.lock.Lock()
	defer repPeriod.lock.Unlock()

	if v, ok := repPeriod.statistic[addr]; ok {
		if kind == onlineKindHeart && now.Sub(v.lastHeartTime) >= findOnlineRepInterval {
			v.heartCount++
			v.lastHeartTime = now
		} else if kind == onlineKindVote {
			c, err := dps.confirmedBlocks.Get(hash)
			if err != nil {
				reps := make(map[types.Address]struct{})
				err := dps.confirmedBlocks.Set(hash, reps)
				if err != nil {
					dps.logger.Error("add confirmed block cache err", err)
				}

				reps[addr] = struct{}{}
				v.voteCount++
			} else {
				reps := c.(map[types.Address]struct{})
				if _, ok := reps[addr]; !ok {
					v.voteCount++
					reps[addr] = struct{}{}
				}
			}
		}
	} else {
		repPeriod.statistic[addr] = &repAckStatistics{}

		if kind == onlineKindHeart {
			repPeriod.statistic[addr].heartCount++
			v.lastHeartTime = now
		} else {
			c, err := dps.confirmedBlocks.Get(hash)
			if err != nil {
				reps := make(map[types.Address]struct{})
				err := dps.confirmedBlocks.Set(hash, reps)
				if err != nil {
					dps.logger.Error("add confirmed block cache err", err)
				}

				reps[addr] = struct{}{}
				repPeriod.statistic[addr].voteCount++
			} else {
				reps := c.(map[types.Address]struct{})
				reps[addr] = struct{}{}
				repPeriod.statistic[addr].voteCount++
			}
		}
	}
}

func (dps *DPoS) confirmedBlockInc() {
	period := (dps.curPovHeight - 1) / onlinePeriod

	if s, err := dps.online.Get(period); err == nil {
		repPeriod := s.(*repOnlinePeriod)
		repPeriod.lock.Lock()
		repPeriod.blockCount++
		repPeriod.lock.Unlock()
	} else {
		repPeriod := &repOnlinePeriod{
			statistic:  make(map[types.Address]*repAckStatistics),
			blockCount: 1,
			lock:       new(sync.RWMutex),
		}

		err := dps.online.Set(period, repPeriod)
		if err != nil {
			dps.logger.Error("set online period err", err)
		}
	}
}

func (dps *DPoS) isOnline(addr types.Address) bool {
	period := (dps.curPovHeight - 1) / onlinePeriod - 1

	//the first period will be ignored
	if s, err := dps.online.Get(period); err == nil {
		repPeriod := s.(*repOnlinePeriod)
		repPeriod.lock.RLock()
		defer repPeriod.lock.RUnlock()

		if repPeriod.blockCount == 0 {
			if v, ok := repPeriod.statistic[addr]; ok {
				if v.heartCount * 100 > heartCountPerPeriod * onlineRate {
					return true
				}
			}
		} else {
			if v, ok := repPeriod.statistic[addr]; ok {
				if v.voteCount * 100 > repPeriod.blockCount * onlineRate {
					return true
				}
			}
		}
	}

	return false
}

func (dps *DPoS) sendOnline() {
	for _, acc := range dps.accounts {
		weight := dps.ledger.Weight(acc.Address())
		if weight.Compare(dps.minVoteWeight) == types.BalanceCompSmaller {
			continue
		}

		blk, err := dps.ledger.GenerateOnlineBlock(acc.Address(), acc.PrivateKey())
		if err != nil {
			dps.logger.Error("generate online block err", err)
			return
		}

		dps.eb.Publish(common.EventBroadcast, p2p.PublishReq, blk)

		bs := &consensus.BlockSource{
			Block:     blk,
			BlockFrom: types.UnSynchronized,
			Type:      consensus.MsgGenerateBlock,
		}
		dps.ProcessMsg(bs)
	}
}

func (dps *DPoS) onPovHeightChange(pb *types.PovBlock) {
	dps.curPovHeight = pb.Height

	if dps.getPovSyncState() == common.Syncdone {
		if pb.Height - dps.lastSendHeight >= 120 {
			dps.sendOnline()
			dps.lastSendHeight = pb.Height
		}
	}
}
