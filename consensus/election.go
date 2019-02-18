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
}

func NewElection(dps *DposService, block types.Block) (*Election, error) {
	vt := NewVotes(block)
	status := electionStatus{block, types.ZeroBalance}

	return &Election{
		vote:          vt,
		status:        status,
		confirmed:     false,
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
		el.dps.ns.Broadcast(p2p.ConfirmAck, va)
	}
	el.haveQuorum()
}

func (el *Election) haveQuorum() {
	t := el.tally()
	if !(len(t) > 0) {
		return
	}
	var hash types.Hash
	var balance = types.ZeroBalance
	for key, value := range t {
		if balance.Compare(value) == types.BalanceCompSmaller {
			balance = value
			hash = key
		}
	}
	blk, err := el.dps.ledger.GetStateBlock(hash)
	if err != nil {
		el.dps.logger.Infof("err:[%s] when get block", err)
	}
	supply := el.getOnlineRepresentativesBalance()
	b, err := supply.Div(2)
	if err != nil {
		return
	}
	if balance.Compare(b) == types.BalanceCompBigger {
		el.dps.logger.Infof("hash:%s block has confirmed", blk.GetHash())
		el.status.winner = blk
		el.confirmed = true
		el.status.tally = balance
	} else {
		el.dps.logger.Infof("wait for enough rep vote,current vote is [%s]", balance.String())
	}
}

func (el *Election) tally() map[types.Hash]types.Balance {
	totals := make(map[types.Hash]types.Balance)
	var hash types.Hash
	for key, value := range el.vote.repVotes {
		hash = value.Blk.GetHash()
		if _, ok := totals[hash]; !ok {
			totals[hash] = types.ZeroBalance
		}
		weight := el.dps.ledger.Weight(key)
		totals[hash] = totals[hash].Add(weight)
	}
	return totals
}

func (el *Election) getOnlineRepresentativesBalance() types.Balance {
	b := types.ZeroBalance
	addresses := el.dps.onlineRepAddresses
	for _, v := range addresses {
		b1, err := el.dps.ledger.GetRepresentation(v)
		if err != nil {
			b = b.Add(b1)
		}
	}
	return b
}
