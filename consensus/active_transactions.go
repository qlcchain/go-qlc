package consensus

import (
	"time"

	"github.com/qlcchain/go-qlc/p2p"

	"github.com/qlcchain/go-qlc/p2p/protos"

	"github.com/qlcchain/go-qlc/common/types"
)

const (
	announcementmin    = 4 //Minimum number of block announcements
	announceIntervalms = 16 * time.Second
)

type ActiveTrx struct {
	confirmed electionStatus
	dps       *DposService
	roots     map[types.Hash]*Election
	quitCh    chan bool
	inactive  []types.Hash
}

func NewActiveTrx() *ActiveTrx {
	return &ActiveTrx{
		quitCh: make(chan bool, 1),
		roots:  make(map[types.Hash]*Election),
	}
}

func (act *ActiveTrx) SetDposService(dps *DposService) {
	act.dps = dps
}

func (act *ActiveTrx) start() {
	timer2 := time.NewTicker(announceIntervalms)
	for {
		select {
		case <-timer2.C:
			logger.Info("begin check roots.")
			act.announceVotes()
		case <-act.quitCh:
			logger.Info("Stopped ActiveTrx Loop.")
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (act *ActiveTrx) addToRoots(block types.Block) bool {
	if _, ok := act.roots[block.Root()]; !ok {
		ele, err := NewElection(act.dps, block)
		if err != nil {
			logger.Infof("block :%s add to roots error", block.GetHash())
			return false
		}
		act.roots[block.Root()] = ele
		return true
	} else {
		logger.Infof("block :%s already exit in roots", block.GetHash())
		return false
	}
}

func (act *ActiveTrx) announceVotes() {
	for _, v := range act.roots {
		if v.confirmed && v.announcements >= announcementmin-1 {
			logger.Info("this block is already confirmed")
			act.dps.ns.MessageEvent().GetEvent("consensus").Notify(p2p.EventConfirmedBlock, v.status.winner)
			act.inactive = append(act.inactive, v.vote.id)
		} else {
			accounts := act.dps.getAccounts()
			for _, k := range accounts {
				isrep := act.dps.isThisAccountRepresentation(k)
				if isrep == true {
					logger.Infof("send confirm ack for hash %s,previous hash is %s", v.status.winner.GetHash(), v.status.winner.Root())
					act.dps.sendConfirmAck(v.status.winner, k)
				} else {
					logger.Infof("send confirm req for hash %s,previous hash is %s", v.status.winner.GetHash(), v.status.winner.Root())
					act.dps.sendConfirmReq(v.status.winner)
				}
			}
			if len(accounts) == 0 {
				logger.Info("this is just a node,not a wallet")
				act.dps.sendConfirmReq(v.status.winner)
			}
			v.announcements++
		}
	}
	for _, value := range act.inactive {
		if _, ok := act.roots[value]; ok {
			delete(act.roots, value)
		}
	}
	act.inactive = act.inactive[:0:0]
}

func (act *ActiveTrx) vote(va *protos.ConfirmAckBlock) {
	if value, ok := act.roots[va.Blk.Root()]; ok {
		value.voteAction(va)
	}
}

func (act *ActiveTrx) stop() {
	act.quitCh <- true
}
