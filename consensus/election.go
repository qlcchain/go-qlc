package consensus

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

type electionStatus struct {
	winner types.Block
	tally  types.Balance
}

type Election struct {
	vote          *Votes
	status        electionStatus
	confirmed     bool
	dps           *DposService
	announcements uint
	supply        types.Balance
}

func NewElection(dps *DposService, block types.Block) (*Election, error) {
	vt := NewVotes(block)
	status := electionStatus{block, types.ZeroBalance}
	//tokenId := mock.GetChainTokenType()
	//ti, err := mock.GetTokenById(tokenId)
	//if err != nil {
	//	return nil, err
	//}
	b1 := types.Balance{Hi: 0, Lo: 40}

	return &Election{
		vote:          vt,
		status:        status,
		confirmed:     false,
		supply:        b1,
		dps:           dps,
		announcements: 0,
	}, nil
}

func (el *Election) voteAction(va *protos.ConfirmAckBlock) {
	valid := IsAckSignValidate(va)
	if valid != true {
		return
	}
	var shouldProcess bool
	exit, vt := el.vote.voteExit(va.Account)
	if exit == true {
		if vt.Sequence < va.Sequence {
			shouldProcess = true
		} else {
			shouldProcess = false
		}
	} else {
		shouldProcess = true
	}
	if shouldProcess {
		el.vote.repVotes[va.Account] = va
		data, err := protos.ConfirmAckBlockToProto(va)
		if err != nil {
			logger.Error("vote to proto error")
		}
		el.dps.ns.Broadcast(p2p.ConfirmAck, data)
	}
	el.haveQuorum()
}

func (el *Election) haveQuorum() {
	t := el.tally()
	if !(len(t) > 0) {
		return
	}
	var blk types.Block
	var balance = types.ZeroBalance
	for key, value := range t {
		if balance.Compare(value) == types.BalanceCompSmaller {
			balance = value
			blk = key
		}
	}
	if balance.Compare(el.supply) == types.BalanceCompBigger {
		logger.Infof("hash:%s block has confirmed", blk.GetHash())
		el.status.winner = blk
		el.confirmed = true
		el.status.tally = balance
	}
}

func (el *Election) tally() map[types.Block]types.Balance {
	totals := make(map[types.Block]types.Balance)
	for key, value := range el.vote.repVotes {
		if _, ok := totals[value.Blk]; !ok {
			totals[value.Blk] = types.ZeroBalance
		}
		weight := el.dps.ledger.Weight(key)
		totals[value.Blk] = totals[value.Blk].Add(weight)
	}
	return totals
}
