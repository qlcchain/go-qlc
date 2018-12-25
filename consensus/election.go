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
	b1, err := types.ParseBalance("40.00000000", "Mqlc")
	if err != nil {
		return nil, err
	}
	return &Election{
		vote:          vt,
		status:        status,
		confirmed:     false,
		supply:        b1,
		dps:           dps,
		announcements: 0,
	}, nil
}
func (el *Election) voteAction(vote_a *protos.ConfirmAckBlock) {
	valid := IsAckSignValidate(vote_a)
	if valid != true {
		return
	}
	var should_process bool
	exit, vt := el.vote.voteExit(vote_a.Account)
	if exit == true {
		if vt.Sequence < vote_a.Sequence {
			should_process = true
		} else {
			should_process = false
		}
	} else {
		should_process = true
	}
	if should_process {
		el.vote.rep_votes[vote_a.Account] = vote_a
		data, err := protos.ConfirmAckBlockToProto(vote_a)
		if err != nil {
			logger.Error("vote to proto error")
		}
		el.dps.ns.Broadcast(p2p.ConfirmAck, data)
	}
	el.haveQuorum()
}
func (el *Election) haveQuorum() {
	tally_l := el.tally()
	if !(len(tally_l) > 0) {
		return
	}
	var blk types.Block
	var balance = types.ZeroBalance
	for key, value := range tally_l {
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
	for key, value := range el.vote.rep_votes {
		if _, ok := totals[value.Blk]; !ok {
			totals[value.Blk] = types.ZeroBalance
		}
		weight := el.dps.ledger.Weight(key)
		totals[value.Blk] = totals[value.Blk].Add(weight)
	}
	return totals
}
