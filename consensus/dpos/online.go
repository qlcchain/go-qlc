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

const (
	onlinePeriod        = uint64(120)
	onlineRate          = uint64(60)
	heartCountPerPeriod = onlinePeriod / 2
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
	period := (dps.curPovHeight - 1) / onlinePeriod
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
	period := (dps.curPovHeight - 1) / onlinePeriod

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
	period := (dps.curPovHeight-1)/onlinePeriod - 1

	//the first period will be ignored
	if s, err := dps.online.Get(period); err == nil {
		repPeriod := s.(*RepOnlinePeriod)
		repPeriod.lock.RLock()
		defer repPeriod.lock.RUnlock()

		if repPeriod.BlockCount == 0 {
			if v, ok := repPeriod.Statistic[addr]; ok {
				if v.HeartCount*100 > heartCountPerPeriod*onlineRate {
					dps.logger.Debugf("[%s] heart online: heart[%d] expect[%d]", addr, v.HeartCount, heartCountPerPeriod)
					return true
				}
				dps.logger.Debugf("[%s] heart offline: heart[%d] expect[%d]", addr, v.HeartCount, heartCountPerPeriod)
			}
			dps.logger.Debugf("[%s] heart offline: no heart", addr)
		} else {
			if v, ok := repPeriod.Statistic[addr]; ok {
				if v.VoteCount*100 > repPeriod.BlockCount*onlineRate {
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

func (dps *DPoS) sendOnlineWithAccount(acc *types.Account) {
	weight := dps.ledger.Weight(acc.Address())
	if weight.Compare(dps.minVoteWeight) == types.BalanceCompSmaller {
		return
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

func (dps *DPoS) onPovHeightChange(pb *types.PovBlock) {
	dps.logger.Debugf("pov height changed [%d]->[%d]", dps.curPovHeight, pb.Header.BasHdr.Height)
	dps.curPovHeight = pb.Header.BasHdr.Height

	if dps.getPovSyncState() == common.Syncdone {
		if dps.curPovHeight%2 == 0 {
			go func() {
				err := dps.findOnlineRepresentatives()
				if err != nil {
					dps.logger.Error(err)
				}
				dps.cleanOnlineReps()
			}()
		}

		if dps.curPovHeight-dps.lastSendHeight >= onlinePeriod && (dps.curPovHeight%onlinePeriod >= 30 || dps.curPovHeight%onlinePeriod <= 90) {
			dps.sendOnline()
			dps.lastSendHeight = pb.Header.BasHdr.Height
		}
	}
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
