package dpos

import (
	"encoding/json"
	"sync"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/p2p"
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

func (dps *DPoS) isValidVote(hash types.Hash, addr types.Address) bool {
	if dps.confirmedBlocks.has(hash) {
		if dps.ledger.HasVoteHistory(hash, addr) {
			return false
		} else {
			err := dps.ledger.AddVoteHistory(hash, addr)
			if err != nil {
				dps.logger.Error("add vote history err", err)
			}
			return true
		}
	}

	return false
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
			if dps.isValidVote(hash, addr) {
				v.VoteCount++
			}
		}
	} else {
		repPeriod.Statistic[addr] = &RepAckStatistics{}

		if kind == onlineKindHeart {
			repPeriod.Statistic[addr].HeartCount++
			repPeriod.Statistic[addr].LastHeartHeight = dps.curPovHeight
		} else {
			if dps.isValidVote(hash, addr) {
				repPeriod.Statistic[addr].VoteCount++
			}
		}
	}
}

func (dps *DPoS) confirmedBlockInc(hash types.Hash) {
	period := (dps.curPovHeight - 1) / common.DPosOnlinePeriod

	dps.confirmedBlocks.set(hash, nil)

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
		if !dps.isOnline(acc.Address()) {
			continue
		}

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
