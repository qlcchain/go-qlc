package dpos

import (
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/qlcchain/go-qlc/common/topic"

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
	Statistic  *sync.Map                           `json:"-"`
	Stat       map[types.Address]*RepAckStatistics `json:"statistics"`
	BlockCount uint64                              `json:"blockCount"`
}

func (op *RepOnlinePeriod) String() string {
	op.Stat = make(map[types.Address]*RepAckStatistics)

	op.Statistic.Range(func(key, value interface{}) bool {
		addr := key.(types.Address)
		ras := value.(*RepAckStatistics)
		op.Stat[addr] = ras
		return true
	})

	bytes, _ := json.Marshal(op)
	return string(bytes)
}

func (dps *DPoS) isValidVote(hash types.Hash, addr types.Address) bool {
	if dps.confirmedBlocks.has(hash) {
		if !dps.ledger.HasVoteHistory(hash, addr) {
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
	period := dps.curPovHeight / common.DPosOnlinePeriod
	var repPeriod *RepOnlinePeriod

	if s, err := dps.online.Get(period); err == nil {
		repPeriod = s.(*RepOnlinePeriod)
	} else {
		repPeriod = &RepOnlinePeriod{
			Period:     period,
			Statistic:  new(sync.Map),
			BlockCount: 0,
		}

		err := dps.online.Set(period, repPeriod)
		if err != nil {
			dps.logger.Error("set online period err", err)
		}
	}

	val, ok := repPeriod.Statistic.LoadOrStore(addr, &RepAckStatistics{})
	stat := val.(*RepAckStatistics)
	lk, _ := dps.lockPool.LoadOrStore(addr, new(sync.RWMutex))
	lock := lk.(*sync.RWMutex)
	lock.Lock()
	defer lock.Unlock()

	if ok {
		if kind == onlineKindHeart && dps.curPovHeight-stat.LastHeartHeight >= 2 {
			stat.HeartCount++
			stat.LastHeartHeight = dps.curPovHeight
		} else if kind == onlineKindVote {
			if dps.isValidVote(hash, addr) {
				stat.VoteCount++
			}
		}
	} else {
		if kind == onlineKindHeart {
			stat.HeartCount++
			stat.LastHeartHeight = dps.curPovHeight
		} else {
			if dps.isValidVote(hash, addr) {
				stat.VoteCount++
			}
		}
	}
}

func (dps *DPoS) confirmedBlockInc(hash types.Hash) {
	period := dps.curPovHeight / common.DPosOnlinePeriod
	dps.confirmedBlocks.set(hash, nil)

	if s, err := dps.online.Get(period); err == nil {
		repPeriod := s.(*RepOnlinePeriod)
		atomic.AddUint64(&repPeriod.BlockCount, 1)
	} else {
		repPeriod := &RepOnlinePeriod{
			Period:     period,
			Statistic:  new(sync.Map),
			BlockCount: 1,
		}

		err := dps.online.Set(period, repPeriod)
		if err != nil {
			dps.logger.Error("set online period err", err)
		}
	}
}

func (dps *DPoS) isOnline(addr types.Address) bool {
	period := dps.curPovHeight/common.DPosOnlinePeriod - 1

	lk, _ := dps.lockPool.LoadOrStore(addr, new(sync.RWMutex))
	lock := lk.(*sync.RWMutex)

	//the first period will be ignored
	if s, err := dps.online.Get(period); err == nil {
		repPeriod := s.(*RepOnlinePeriod)

		if val, ok := repPeriod.Statistic.Load(addr); ok {
			lock.RLock()
			defer lock.RUnlock()
			v := val.(*RepAckStatistics)

			if atomic.LoadUint64(&repPeriod.BlockCount) == 0 {
				if v.HeartCount*100 > common.DPosHeartCountPerPeriod*common.DPosOnlineRate {
					dps.logger.Debugf("[%s] heart online: heart[%d] expect[%d]", addr, v.HeartCount, common.DPosHeartCountPerPeriod)
					return true
				} else {
					dps.logger.Debugf("[%s] heart offline: heart[%d] expect[%d]", addr, v.HeartCount, common.DPosHeartCountPerPeriod)
				}
			} else {
				if v.VoteCount*100 > repPeriod.BlockCount*common.DPosOnlineRate {
					dps.logger.Debugf("[%s] vote online: vote[%d] expect[%d]", addr, v.VoteCount, repPeriod.BlockCount)
					return true
				} else {
					dps.logger.Debugf("[%s] vote offline: vote[%d] expect[%d]", addr, v.VoteCount, repPeriod.BlockCount)
				}
			}
		}
	}

	dps.logger.Debugf("[%s] offline: no heart and no vote", addr)
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
		dps.eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.PublishReq, Message: blk})

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

	dps.eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.PublishReq, Message: blk})

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
			rop := o.(*RepOnlinePeriod)
			rop.Stat = make(map[types.Address]*RepAckStatistics)

			rop.Statistic.Range(func(key, value interface{}) bool {
				addr := key.(types.Address)
				ras := value.(*RepAckStatistics)
				rop.Stat[addr] = ras
				return true
			})

			op[p.(uint64)] = rop
		}
	}
}
