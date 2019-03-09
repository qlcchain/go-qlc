package consensus

import (
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	//announcementMin        = 4 //Minimum number of block announcements
	announceIntervalSecond = 16 * time.Second
	refreshPriInfoHour     = 1 * time.Hour
)

type ActiveTrx struct {
	confirmed electionStatus
	dps       *DposService
	//roots     map[types.Hash]*Election
	roots    *sync.Map
	quitCh   chan bool
	inactive []types.Hash
}

func NewActiveTrx() *ActiveTrx {
	return &ActiveTrx{
		quitCh: make(chan bool, 1),
		//roots:  make(map[types.Hash]*Election),
		roots: new(sync.Map),
	}
}

func (act *ActiveTrx) SetDposService(dps *DposService) {
	act.dps = dps
}

func (act *ActiveTrx) start() {
	timer2 := time.NewTicker(announceIntervalSecond)
	timer3 := time.NewTicker(refreshPriInfoHour)
	for {
		select {
		case <-timer2.C:
			act.dps.logger.Info("begin check roots.")
			act.announceVotes()
		case <-act.quitCh:
			act.dps.logger.Info("Stopped ActiveTrx Loop.")
			return
		case <-timer3.C:
			act.dps.logger.Info("refresh pri info.")
			go act.dps.refreshPriInfo()
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (act *ActiveTrx) addToRoots(block *types.StateBlock) bool {
	if _, ok := act.roots.Load(block.Root()); !ok {
		ele, err := NewElection(act.dps, block)
		if err != nil {
			act.dps.logger.Errorf("block :%s add to roots error", block.GetHash())
			return false
		}
		act.roots.Store(block.Root(), ele)
		return true
	} else {
		act.dps.logger.Infof("block :%s already exit in roots", block.GetHash())
		return false
	}
}

func (act *ActiveTrx) announceVotes() {
	var count = 0
	act.roots.Range(func(key, value interface{}) bool {
		block := value.(*Election).status.winner
		hash := block.GetHash()
		if act.dps.cfg.PerformanceTest.Enabled {
			if value.(*Election).announcements == 0 {
				if p, err := act.dps.ledger.GetPerformanceTime(hash); p != nil && err == nil {
					t := &types.PerformanceTime{
						Hash: hash,
						T0:   p.T0,
						T1:   p.T1,
						T2:   time.Now().UnixNano(),
						T3:   p.T3,
					}
					act.dps.ledger.AddOrUpdatePerformance(t)
					if err != nil {
						act.dps.logger.Info("AddOrUpdatePerformance error T2")
					}
				} else {
					act.dps.logger.Info("get performanceTime error T2")
				}
			}
		}
		if value.(*Election).confirmed { //&& value.(*Election).announcements >= announcementMin-1 {
			if act.dps.cfg.PerformanceTest.Enabled {
				var t *types.PerformanceTime
				if p, err := act.dps.ledger.GetPerformanceTime(hash); p != nil && err == nil {
					if value.(*Election).announcements == 0 {
						t = &types.PerformanceTime{
							Hash: hash,
							T0:   p.T0,
							T1:   time.Now().UnixNano(),
							T2:   p.T2,
							T3:   time.Now().UnixNano(),
						}
					} else {
						t = &types.PerformanceTime{
							Hash: hash,
							T0:   p.T0,
							T1:   time.Now().UnixNano(),
							T2:   p.T2,
							T3:   p.T3,
						}
					}
					err := act.dps.ledger.AddOrUpdatePerformance(t)
					if err != nil {
						act.dps.logger.Info("AddOrUpdatePerformance error T1")
					}
				} else {
					act.dps.logger.Info("get performanceTime error T1")
				}
			}
			act.dps.logger.Infof("block [%s] is already confirmed", hash)
			act.dps.ns.MessageEvent().GetEvent("consensus").Notify(p2p.EventConfirmedBlock, block)
			act.inactive = append(act.inactive, value.(*Election).vote.id)
			act.rollBack(value.(*Election).status.loser)
			act.addWinner2Ledger(block)
		} else {
			act.dps.priInfos.Range(func(k, v interface{}) bool {
				count++
				isRep := act.dps.isThisAccountRepresentation(k.(types.Address))
				if isRep {
					act.dps.putRepresentativesToOnline(k.(types.Address))
					va, err := act.dps.voteGenerate(block, k.(types.Address), v.(*types.Account))
					if err != nil {
						act.dps.logger.Error("vote generate error")
					} else {
						act.dps.logger.Infof("vote:send confirm ack for hash %s,previous hash is %s", hash, block.Root())
						act.dps.ns.Broadcast(p2p.ConfirmAck, va)
						value.(*Election).voteAction(va)
					}
				} else {
					act.dps.logger.Infof("vote:send confirmReq for hash %s,previous hash is %s", hash, block.Root())
					act.dps.ns.Broadcast(p2p.ConfirmReq, block)
				}
				return true
			})
			if count == 0 {
				act.dps.logger.Info("this is just a node,not a wallet")
				act.dps.ns.Broadcast(p2p.ConfirmReq, block)
			}
			if act.dps.cfg.PerformanceTest.Enabled {
				if value.(*Election).announcements == 0 {
					if p, err := act.dps.ledger.GetPerformanceTime(hash); p != nil && err == nil {
						t := &types.PerformanceTime{
							Hash: hash,
							T0:   p.T0,
							T1:   p.T1,
							T2:   p.T2,
							T3:   time.Now().UnixNano(),
						}
						act.dps.ledger.AddOrUpdatePerformance(t)
						if err != nil {
							act.dps.logger.Info("AddOrUpdatePerformance error T3")
						}
					} else {
						act.dps.logger.Info("get performanceTime error T3")
					}
				}
			}
			value.(*Election).announcements++
		}
		return true
	})

	for _, value := range act.inactive {
		if _, ok := act.roots.Load(value); ok {
			act.roots.Delete(value)
		}
	}
	act.inactive = act.inactive[:0:0]
}

func (act *ActiveTrx) addWinner2Ledger(block *types.StateBlock) {
	hash := block.GetHash()
	if exist, err := act.dps.ledger.HasStateBlock(hash); !exist && err == nil {
		err := act.dps.ledger.BlockProcess(block)
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
		}
	}
}

func (act *ActiveTrx) vote(va *protos.ConfirmAckBlock) {
	if v, ok := act.roots.Load(va.Blk.Root()); ok {
		v.(*Election).voteAction(va)
	}
}

func (act *ActiveTrx) stop() {
	act.quitCh <- true
}
