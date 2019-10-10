package dpos

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/p2p"
	"sync"
)

type onlineKind byte

const (
	onlineKindHeart onlineKind = iota
	onlineKindVote
)

type RepAckStatistics struct {
	HeartCount      uint64 `json:"heartCount"`
	LastHeartHeight uint64 `json:"-"`
	VoteCount       uint64 `json:"voteCount"`
}

type RepOnlinePeriod struct {
	Period     uint64                              `json:"period"`
	Statistic  map[types.Address]*RepAckStatistics `json:"statistics"`
	BlockCount uint64                              `json:"blockCount"`
	lock       *sync.RWMutex
}

func (op *RepOnlinePeriod) String() string {
	bytes, _ := json.Marshal(op)
	return string(bytes)
}

func (dps *DPoS) heartAndVoteInc(hash types.Hash, addr types.Address, kind onlineKind) {
	period := (dps.curPovHeight - 1) / common.DPosOnlinePeriod
	var repPeriod *RepOnlinePeriod

	if s, err := dps.online.Get(period); err == nil {
		repPeriod = s.(*RepOnlinePeriod)
	} else {
		repPeriod = &RepOnlinePeriod{
			Period:     period,
			Statistic:  make(map[types.Address]*RepAckStatistics),
			BlockCount: 0,
			lock:       new(sync.RWMutex),
		}

		err := dps.online.Set(period, repPeriod)
		if err != nil {
			dps.logger.Error("set online period err", err)
		}
	}

	repPeriod.lock.Lock()
	defer repPeriod.lock.Unlock()

	if v, ok := repPeriod.Statistic[addr]; ok {
		if kind == onlineKindHeart && dps.curPovHeight-v.LastHeartHeight >= 2 {
			v.HeartCount++
			v.LastHeartHeight = dps.curPovHeight
		} else if kind == onlineKindVote {
			c, err := dps.confirmedBlocks.Get(hash)
			if err != nil {
				reps := make(map[types.Address]struct{})
				err := dps.confirmedBlocks.Set(hash, reps)
				if err != nil {
					dps.logger.Error("add confirmed block cache err", err)
				}

				reps[addr] = struct{}{}
				v.VoteCount++
			} else {
				reps := c.(map[types.Address]struct{})
				if _, ok := reps[addr]; !ok {
					v.VoteCount++
					reps[addr] = struct{}{}
				}
			}
		}
	} else {
		repPeriod.Statistic[addr] = &RepAckStatistics{}

		if kind == onlineKindHeart {
			repPeriod.Statistic[addr].HeartCount++
			repPeriod.Statistic[addr].LastHeartHeight = dps.curPovHeight
		} else {
			c, err := dps.confirmedBlocks.Get(hash)
			if err != nil {
				reps := make(map[types.Address]struct{})
				err := dps.confirmedBlocks.Set(hash, reps)
				if err != nil {
					dps.logger.Error("add confirmed block cache err", err)
				}

				reps[addr] = struct{}{}
				repPeriod.Statistic[addr].VoteCount++
			} else {
				reps := c.(map[types.Address]struct{})
				reps[addr] = struct{}{}
				repPeriod.Statistic[addr].VoteCount++
			}
		}
	}
}

func (dps *DPoS) confirmedBlockInc() {
	period := (dps.curPovHeight - 1) / common.DPosOnlinePeriod

	if s, err := dps.online.Get(period); err == nil {
		repPeriod := s.(*RepOnlinePeriod)
		repPeriod.lock.Lock()
		repPeriod.BlockCount++
		repPeriod.lock.Unlock()
	} else {
		repPeriod := &RepOnlinePeriod{
			Period:     period,
			Statistic:  make(map[types.Address]*RepAckStatistics),
			BlockCount: 1,
			lock:       new(sync.RWMutex),
		}

		err := dps.online.Set(period, repPeriod)
		if err != nil {
			dps.logger.Error("set online period err", err)
		}
	}
}

func (dps *DPoS) isOnline(addr types.Address) bool {
	period := (dps.curPovHeight-1)/common.DPosOnlinePeriod - 1

	//the first period will be ignored
	if s, err := dps.online.Get(period); err == nil {
		repPeriod := s.(*RepOnlinePeriod)
		repPeriod.lock.RLock()
		defer repPeriod.lock.RUnlock()

		if repPeriod.BlockCount == 0 {
			if v, ok := repPeriod.Statistic[addr]; ok {
				if v.HeartCount*100 > common.DPosHeartCountPerPeriod*common.DPosOnlineRate {
					dps.logger.Debugf("[%s] heart online: heart[%d] expect[%d]", addr, v.HeartCount, common.DPosHeartCountPerPeriod)
					return true
				}
				dps.logger.Debugf("[%s] heart offline: heart[%d] expect[%d]", addr, v.HeartCount, common.DPosHeartCountPerPeriod)
			}
			dps.logger.Debugf("[%s] heart offline: no heart", addr)
		} else {
			if v, ok := repPeriod.Statistic[addr]; ok {
				if v.VoteCount*100 > repPeriod.BlockCount*common.DPosOnlineRate {
					dps.logger.Debugf("[%s] vote online: vote[%d] expect[%d]", addr, v.VoteCount, repPeriod.BlockCount)
					return true
				}
				dps.logger.Debugf("[%s] vote offline: vote[%d] expect[%d]", addr, v.VoteCount, repPeriod.BlockCount)
			}
			dps.logger.Debugf("[%s] vote offline: no vote", addr)
		}
	} else {
		dps.logger.Debugf("[%s] offline: no heart and no vote", addr)
	}

	return false
}

func (dps *DPoS) sendOnline(povHeight uint64) {
	for _, acc := range dps.accounts {
		weight := dps.ledger.Weight(acc.Address())
		if weight.Compare(dps.minVoteWeight) == types.BalanceCompSmaller {
			continue
		}

		blk, err := dps.ledger.GenerateOnlineBlock(acc.Address(), acc.PrivateKey(), povHeight)
		if err != nil {
			dps.logger.Error("generate online block err", err)
			return
		}

		dps.logger.Debugf("send online block[%s]", blk.GetHash())
		dps.eb.Publish(common.EventBroadcast, p2p.PublishReq, blk)

		bs := &consensus.BlockSource{
			Block:     blk,
			BlockFrom: types.UnSynchronized,
			Type:      consensus.MsgGenerateBlock,
		}
		dps.ProcessMsg(bs)
	}
}

func (dps *DPoS) sendOnlineWithAccount(acc *types.Account, povHeight uint64) {
	weight := dps.ledger.Weight(acc.Address())
	if weight.Compare(dps.minVoteWeight) == types.BalanceCompSmaller {
		return
	}

	blk, err := dps.ledger.GenerateOnlineBlock(acc.Address(), acc.PrivateKey(), povHeight)
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

func (dps *DPoS) onPovHeightChange(pb *types.PovBlock) {
	dps.povChange <- pb
}

func (dps *DPoS) onGetOnlineInfo(in interface{}, out interface{}) {
	op := out.(map[uint64]*RepOnlinePeriod)

	online := dps.online.GetALL(true)
	if online != nil {
		for p, o := range online {
			op[p.(uint64)] = o.(*RepOnlinePeriod)
		}
	}
}
