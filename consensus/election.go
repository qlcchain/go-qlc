package consensus

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/test/mock"
)

type BlockReceivedVotes struct {
	block   types.Block
	balance types.Balance
}

type electionStatus struct {
	winner types.Block
	tally  types.Balance
	loser  []types.Block
}

type Election struct {
	vote      *Votes
	status    electionStatus
	confirmed bool
	dps       *DposService
	//announcements uint
}

func NewElection(dps *DposService, block types.Block) (*Election, error) {
	vt := NewVotes(block)
	status := electionStatus{block, types.ZeroBalance, nil}

	return &Election{
		vote:      vt,
		status:    status,
		confirmed: false,
		dps:       dps,
		//announcements: 0,
	}, nil
}

func (el *Election) voteAction(va *protos.ConfirmAckBlock) {
	valid := IsAckSignValidate(va)
	if !valid {
		return
	}
	exit, _ := el.vote.voteExit(va.Account)
	if !exit {
		el.vote.repVotes[va.Account] = va
	}
	el.haveQuorum()
}

func (el *Election) haveQuorum() {
	t := el.tally()
	if !(len(t) > 0) {
		return
	}
	var balance = types.ZeroBalance
	var blk types.Block
	for _, value := range t {
		if balance.Compare(value.balance) == types.BalanceCompSmaller {
			balance = value.balance
			blk = value.block
		}
	}
	//supply := el.getOnlineRepresentativesBalance()
	supply, err := el.getGenesisBalance()
	if err != nil {
		return
	}
	b, err := supply.Div(2)
	if err != nil {
		return
	}
	if balance.Compare(b) == types.BalanceCompBigger {
		confirmedHash := blk.GetHash()
		el.dps.logger.Infof("hash:%s block has confirmed", confirmedHash)
		if el.status.winner.GetHash().String() != confirmedHash.String() {
			el.dps.logger.Infof("hash:%s ...is loser", el.status.winner.GetHash().String())
			el.status.loser = append(el.status.loser, el.status.winner)
		}
		el.status.winner = blk
		el.confirmed = true
		el.status.tally = balance
		for _, value := range t {
			if value.block.GetHash().String() != confirmedHash.String() {
				el.status.loser = append(el.status.loser, value.block)
			}
		}
	} else {
		el.dps.logger.Infof("wait for enough rep vote,current vote is [%s]", balance.String())
	}
}

func (el *Election) tally() map[types.Hash]*BlockReceivedVotes {
	totals := make(map[types.Hash]*BlockReceivedVotes)
	var hash types.Hash
	for key, value := range el.vote.repVotes {
		hash = value.Blk.GetHash()
		if _, ok := totals[hash]; !ok {
			totals[hash] = &BlockReceivedVotes{
				block:   value.Blk,
				balance: types.ZeroBalance,
			}
		}
		weight := el.dps.ledger.Weight(key)
		totals[hash].balance = totals[hash].balance.Add(weight)
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

func (el *Election) getGenesisBalance() (types.Balance, error) {
	hash := mock.GetChainTokenType()
	i, err := mock.GetTokenById(hash)
	if err != nil {
		return types.ZeroBalance, err
	}
	return i.TotalSupply, nil
}
