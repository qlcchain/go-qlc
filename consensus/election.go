package consensus

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

type BlockReceivedVotes struct {
	block   *types.StateBlock
	balance types.Balance
}

type electionStatus struct {
	winner *types.StateBlock
	tally  types.Balance
	loser  []*types.StateBlock
}

type Election struct {
	vote          *Votes
	status        electionStatus
	confirmed     bool
	dps           *DPoS
	announcements uint
}

func NewElection(dps *DPoS, block *types.StateBlock) (*Election, error) {
	vt := NewVotes(block)
	status := electionStatus{block, types.ZeroBalance, nil}

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
	if !valid {
		return
	}
	result := el.vote.voteStatus(va)
	if result == confirm {
		return
	}
	el.haveQuorum()
}

func (el *Election) haveQuorum() {
	t := el.tally()
	if !(len(t) > 0) {
		return
	}
	var balance = types.ZeroBalance
	blk := new(types.StateBlock)
	for _, value := range t {
		if balance.Compare(value.balance) == types.BalanceCompSmaller {
			balance = value.balance
			blk = value.block
		}
	}
	//supply := el.getOnlineRepresentativesBalance()
	supply := common.GenesisBlock().Balance
	//supply, err := el.getGenesisBalance()
	//if err != nil {
	//	return
	//}
	b, err := supply.Div(2)
	if err != nil {
		return
	}
	confirmedHash := blk.GetHash()
	if balance.Compare(b) == types.BalanceCompBigger {
		el.dps.logger.Infof("hash:%s block has confirmed,total vote is [%s]", confirmedHash, balance.String())
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
		el.dps.logger.Infof("wait for enough rep vote for block [%s],current vote is [%s]", confirmedHash, balance.String())
	}
}

func (el *Election) tally() map[types.Hash]*BlockReceivedVotes {
	totals := make(map[types.Hash]*BlockReceivedVotes)
	var hash types.Hash
	el.vote.repVotes.Range(func(key, value interface{}) bool {
		hash = value.(*protos.ConfirmAckBlock).Blk.GetHash()
		if _, ok := totals[hash]; !ok {
			totals[hash] = &BlockReceivedVotes{
				block:   value.(*protos.ConfirmAckBlock).Blk,
				balance: types.ZeroBalance,
			}
		}
		weight := el.dps.ledger.Weight(key.(types.Address))
		totals[hash].balance = totals[hash].balance.Add(weight)
		return true
	})
	return totals
}

func (el *Election) getOnlineRepresentativesBalance() types.Balance {
	b := types.ZeroBalance
	reps := el.dps.GetOnlineRepresentatives()
	for _, addr := range reps {
		if b1, err := el.dps.ledger.GetRepresentation(addr); err != nil {
			b = b.Add(b1)
		}
	}
	return b
}

func (el *Election) getGenesisBalance() (types.Balance, error) {
	hash := common.ChainToken()
	i, err := el.dps.ledger.GetTokenById(hash)
	if err != nil {
		return types.ZeroBalance, err
	}
	return types.Balance{Int: i.TotalSupply}, nil
}
