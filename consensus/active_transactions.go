package consensus

import (
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/p2p"

	"github.com/qlcchain/go-qlc/common"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	//announcementMin        = 4 //Minimum number of block announcements
	announcementMax        = 20 //Max number of block announcements
	announceIntervalSecond = 16 * time.Second
	refreshPriInterval     = 5 * time.Minute
)

type ActiveTrx struct {
	confirmed electionStatus
	dps       *DPoS
	roots     *sync.Map
	quitCh    chan bool
	inactive  []types.Hash
}

func NewActiveTrx() *ActiveTrx {
	return &ActiveTrx{
		quitCh: make(chan bool, 1),
		roots:  new(sync.Map),
	}
}

func (act *ActiveTrx) SetDposService(dps *DPoS) {
	act.dps = dps
}

func (act *ActiveTrx) start() {
	timer2 := time.NewTicker(announceIntervalSecond)
	timer3 := time.NewTicker(refreshPriInterval)
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
			go func() {
				act.dps.refreshAccount()
			}()
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (act *ActiveTrx) addToRoots(block *types.StateBlock) bool {
	if _, ok := act.roots.Load(block.Parent()); !ok {
		ele, err := NewElection(act.dps, block)
		if err != nil {
			act.dps.logger.Errorf("block :%s add to roots error", block.GetHash())
			return false
		}
		act.roots.Store(block.Parent(), ele)
		return true
	} else {
		act.dps.logger.Infof("block :%s already exit in roots", block.GetHash())
		return false
	}
}

func (act *ActiveTrx) announceVotes() {
	var address types.Address
	var count uint32
	act.roots.Range(func(key, value interface{}) bool {
		block := value.(*Election).status.winner
		hash := block.GetHash()
		if act.dps.cfg.PerformanceEnabled {
			if value.(*Election).announcements == 0 {
				if p, err := act.dps.ledger.GetPerformanceTime(hash); p != nil && err == nil {
					t := &types.PerformanceTime{
						Hash: hash,
						T0:   p.T0,
						T1:   p.T1,
						T2:   time.Now().UnixNano(),
						T3:   p.T3,
					}
					err = act.dps.ledger.AddOrUpdatePerformance(t)
					if err != nil {
						act.dps.logger.Info("AddOrUpdatePerformance error T2")
					}
				} else {
					act.dps.logger.Info("get performanceTime error T2")
				}
			}
		}
		if value.(*Election).confirmed { //&& value.(*Election).announcements >= announcementMin-1 {
			if act.dps.cfg.PerformanceEnabled {
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
			act.dps.eb.Publish(string(common.EventConfirmedBlock), block)
			//act.dps.ns.MessageEvent().GetEvent("consensus").Notify(p2p.EventConfirmedBlock, block)
			act.inactive = append(act.inactive, value.(*Election).vote.id)
			act.rollBack(value.(*Election).status.loser)
			act.addWinner2Ledger(block)
		} else {
			localRepAccount.Range(func(k, v interface{}) bool {
				count++
				address = k.(types.Address)
				act.dps.saveOnlineRep(address)
				va, err := act.dps.voteGenerate(block, address, v.(*types.Account))
				if err != nil {
					act.dps.logger.Error("vote generate error")
				} else {
					act.dps.logger.Infof("vote:send confirm ack for hash %s,previous hash is %s", hash, block.Parent())
					//act.dps.ns.Broadcast(p2p.ConfirmAck, va)
					act.dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmAck, va)
					value.(*Election).voteAction(va)
				}
				return true
			})
			if count == 0 {
				act.dps.logger.Infof("vote:send confirmReq for block [%s]", hash)
				//act.dps.ns.Broadcast(p2p.ConfirmReq, block)
				act.dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmReq, block)
			}
			if act.dps.cfg.PerformanceEnabled {
				if value.(*Election).announcements == 0 {
					if p, err := act.dps.ledger.GetPerformanceTime(hash); p != nil && err == nil {
						t := &types.PerformanceTime{
							Hash: hash,
							T0:   p.T0,
							T1:   p.T1,
							T2:   p.T2,
							T3:   time.Now().UnixNano(),
						}
						err = act.dps.ledger.AddOrUpdatePerformance(t)
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
		if value.(*Election).announcements == announcementMax {
			if _, ok := act.roots.Load(value); !ok {
				act.inactive = append(act.inactive, value.(*Election).vote.id)
			}
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
		err := act.dps.verifier.BlockProcess(block)
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
		act.deleteUncheckedDependFork(hash)
	}
}

func (act *ActiveTrx) deleteUncheckedDependFork(hash types.Hash) {
	blkLink, _, _ := act.dps.ledger.GetUncheckedBlock(hash, types.UncheckedKindLink)
	blkPrevious, _, _ := act.dps.ledger.GetUncheckedBlock(hash, types.UncheckedKindPrevious)

	if blkLink == nil && blkPrevious == nil {
		return
	}
	if blkLink != nil {
		err := act.dps.ledger.DeleteUncheckedBlock(blkLink.GetHash(), types.UncheckedKindLink)
		if err != nil {
			act.dps.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindLink", err, blkLink.GetHash())
		}
		act.deleteUncheckedDependFork(blkLink.GetHash())
	}
	if blkPrevious != nil {
		err := act.dps.ledger.DeleteUncheckedBlock(blkPrevious.GetHash(), types.UncheckedKindPrevious)
		if err != nil {
			act.dps.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindPrevious", err, blkPrevious.GetHash())
		}
		act.deleteUncheckedDependFork(blkPrevious.GetHash())
	}
}

func (act *ActiveTrx) vote(va *protos.ConfirmAckBlock) {
	if v, ok := act.roots.Load(va.Blk.Parent()); ok {
		v.(*Election).voteAction(va)
	}
}

func (act *ActiveTrx) stop() {
	act.quitCh <- true
}
