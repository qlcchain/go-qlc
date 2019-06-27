package dpos

import (
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
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
	lastTime      int64
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
		lastTime:      time.Now().Unix(),
	}, nil
}

func (el *Election) voteAction(va *protos.ConfirmAckBlock) {
	valid := consensus.IsAckSignValidate(va)
	if !valid {
		el.dps.logger.Errorf("ack sign err %s", va.Blk.GetHash())
		return
	}

	result := el.vote.voteStatus(va)
	if result == confirm {
		el.dps.logger.Infof("recv same ack %s", va.Account.String())
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

		el.dps.acTrx.roots.Delete(el.vote.id)
		el.dps.acTrx.updatePerfTime(blk.GetHash(), time.Now().UnixNano(), true)

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

		el.dps.acTrx.rollBack(el.status.loser)
		el.dps.acTrx.addWinner2Ledger(blk)
		el.dps.blocksAcked <- blk.GetHash()
		el.dps.eb.Publish(string(common.EventConfirmedBlock), blk)
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
		el.dps.logger.Infof("rep[%s] ack block[%s] weight[%s]", key.(types.Address), hash, weight)
		return true
	})

	return totals
}

func (el *Election) getOnlineRepresentativesBalance() types.Balance {
	b := types.ZeroBalance
	reps := el.dps.GetOnlineRepresentatives()

	for _, addr := range reps {
		if b1, err := el.dps.ledger.GetRepresentation(addr); err != nil {
			b = b.Add(b1.Total)
		}
	}

	return b
}

func (el *Election) getGenesisBalance() (types.Balance, error) {
	genesis := common.GenesisBlock()
	return genesis.Balance, nil
}
