package consensus

import (
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/p2p"

	"github.com/qlcchain/go-qlc/p2p/protos"

	"github.com/qlcchain/go-qlc/common/types"
)

const (
	announcementMin        = 4 //Minimum number of block announcements
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
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (act *ActiveTrx) addToRoots(block types.Block) bool {
	if _, ok := act.roots.Load(block.Root()); !ok {
		ele, err := NewElection(act.dps, block)
		if err != nil {
			act.dps.logger.Errorf("block :%s add to roots error", block.GetHash())
			return false
		}
		act.roots.Store(block.Root(), ele)
		act.dps.logger.Info("add hash...............to root:", block.GetHash())
		return true
	} else {
		act.dps.logger.Infof("block :%s already exit in roots", block.GetHash())
		return false
	}
}

func (act *ActiveTrx) announceVotes() {
	var count = 0
	act.roots.Range(func(key, value interface{}) bool {
		if value.(*Election).confirmed { //&& value.(*Election).announcements >= announcementMin-1 {
			act.dps.logger.Info("this block is already confirmed")
			act.dps.ns.MessageEvent().GetEvent("consensus").Notify(p2p.EventConfirmedBlock, value.(*Election).status.winner)
			act.inactive = append(act.inactive, value.(*Election).vote.id)
			act.rollBack(value.(*Election).status.loser)
			act.addWinner2Ledger(value.(*Election).status.winner)
		} else {
			act.dps.priInfos.Range(func(k, v interface{}) bool {
				count++
				isRep := act.dps.isThisAccountRepresentation(k.(types.Address))
				if isRep {
					act.dps.putRepresentativesToOnline(k.(types.Address))
					va, err := act.dps.voteGenerate(value.(*Election).status.winner, k.(types.Address), v.(*types.Account))
					if err != nil {
						act.dps.logger.Error("vote generate error")
					} else {
						act.dps.logger.Infof("vote:send confirm ack for hash %s,previous hash is %s", value.(*Election).status.winner.GetHash(), value.(*Election).status.winner.Root())
						act.dps.ns.Broadcast(p2p.ConfirmAck, va)
						act.vote(va)
					}
				} else {
					act.dps.logger.Infof("vote:send publish for hash %s,previous hash is %s", value.(*Election).status.winner.GetHash(), value.(*Election).status.winner.Root())
					act.dps.ns.Broadcast(p2p.PublishReq, value.(*Election).status.winner)
				}
				return true
			})
			if count == 0 {
				act.dps.logger.Info("this is just a node,not a wallet")
				act.dps.ns.Broadcast(p2p.PublishReq, value.(*Election).status.winner)
			}
			//value.(*Election).announcements++
		}
		return true
	})
	//for _, v := range act.roots {
	//	if v.confirmed && v.announcements >= announcementmin-1 {
	//		act.dps.logger.Info("this block is already confirmed")
	//		act.dps.ns.MessageEvent().GetEvent("consensus").Notify(p2p.EventConfirmedBlock, v.status.winner)
	//		act.inactive = append(act.inactive, v.vote.id)
	//	} else {
	//		act.dps.priInfos.Range(func(key, value interface{}) bool {
	//			count++
	//			isrep := act.dps.isThisAccountRepresentation(key.(types.Address))
	//			if isrep {
	//				act.dps.logger.Infof("send confirm ack for hash %s,previous hash is %s", v.status.winner.GetHash(), v.status.winner.Root())
	//				act.dps.putRepresentativesToOnline(key.(types.Address))
	//				va, err := act.dps.voteGenerate(v.status.winner, key.(types.Address), value.(*types.Account))
	//				if err != nil {
	//					act.dps.logger.Error("vote generate error")
	//				} else {
	//					act.vote(va)
	//				}
	//			} else {
	//				act.dps.logger.Infof("send confirm req for hash %s,previous hash is %s", v.status.winner.GetHash(), v.status.winner.Root())
	//				err := act.dps.sendConfirmReq(v.status.winner)
	//				if err != nil {
	//					act.dps.logger.Error("send confirm req fail.")
	//				}
	//			}
	//			return true
	//		})
	//		if count == 0 {
	//			act.dps.logger.Info("this is just a node,not a wallet")
	//			act.dps.sendConfirmReq(v.status.winner)
	//		}
	//		v.announcements++
	//	}
	//}
	for _, value := range act.inactive {
		if _, ok := act.roots.Load(value); ok {
			act.roots.Delete(value)
		}
	}
	act.inactive = act.inactive[:0:0]
}

func (act *ActiveTrx) addWinner2Ledger(block types.Block) {
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

func (act *ActiveTrx) rollBack(blocks []types.Block) {
	act.dps.logger.Info("22222222333333333333444444444")
	act.dps.logger.Info("len.........:", len(blocks))
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
