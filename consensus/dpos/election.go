package dpos

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/qlcchain/go-qlc/common/topic"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
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
	dps           *DPoS
	announcements uint
	lastTime      int64
	voteHash      types.Hash //vote for this hash
	blocks        *sync.Map
	valid         int32
}

func newElection(dps *DPoS, block *types.StateBlock) *Election {
	vt := newVotes(block)
	hash := block.GetHash()
	status := electionStatus{block, types.ZeroBalance, nil}

	el := &Election{
		vote:          vt,
		status:        status,
		dps:           dps,
		announcements: 0,
		lastTime:      time.Now().Unix(),
		voteHash:      types.ZeroHash,
		blocks:        new(sync.Map),
		valid:         1,
	}

	el.blocks.Store(hash, block)
	dps.hash2el.Store(hash, el)

	return el
}

func (el *Election) voteAction(vi *voteInfo) {
	if !el.isValid() {
		return
	}

	result := el.vote.voteStatus(vi)
	if result == confirm {
		el.dps.logger.Infof("recv same ack %s", vi.account)
		return
	}

	el.haveQuorum()
}

func (el *Election) updateVoteStatistic(confirmedHash types.Hash) {
	dps := el.dps

	//ignore fork ack
	el.vote.repVotes.Range(func(key, value interface{}) bool {
		vi := value.(*voteInfo)
		if vi.hash == confirmedHash {
			dps.heartAndVoteInc(confirmedHash, vi.account, onlineKindVote)
		}
		return true
	})
}

func (el *Election) voteFrontier(vi *voteInfo) bool {
	if !el.isValid() {
		return false
	}

	result := el.vote.voteStatus(vi)
	if result == confirm {
		el.dps.logger.Infof("recv same ack %s", vi.account)
		return false
	}

	t := el.tally(true)
	if !(len(t) > 0) {
		return false
	}

	var balance = types.ZeroBalance
	for _, value := range t {
		if balance.Compare(value.balance) == types.BalanceCompSmaller {
			balance = value.balance
		}
	}

	if balance.Compare(el.dps.voteThreshold) == types.BalanceCompBigger {
		if !el.ifValidAndSetInvalid() {
			return true
		}

		loser := make([]*types.StateBlock, 0)
		el.blocks.Range(func(key, value interface{}) bool {
			if key.(types.Hash) != vi.hash {
				loser = append(loser, value.(*types.StateBlock))
			}
			return true
		})

		el.cleanBlockInfo()
		el.dps.acTrx.rollBack(loser)
		el.dps.acTrx.roots.Delete(el.vote.id)
		return true
	}

	return false
}

func (el *Election) haveQuorum() {
	dps := el.dps

	t := el.tally(false)
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

	confirmedHash := blk.GetHash()
	if balance.Compare(el.dps.voteThreshold) == types.BalanceCompBigger {
		if !el.ifValidAndSetInvalid() {
			return
		}

		dps.acTrx.roots.Delete(el.vote.id)
		el.dps.logger.Infof("hash:%s block has confirmed,total vote is [%s]", confirmedHash, balance)
		dps.acTrx.updatePerfTime(blk.GetHash(), time.Now().UnixNano(), true)

		if el.status.winner.GetHash() != confirmedHash {
			dps.logger.Infof("hash:%s ...is loser", el.status.winner.GetHash().String())
			el.status.loser = append(el.status.loser, el.status.winner)
		}

		for _, value := range t {
			thash := value.block.GetHash()
			if thash != confirmedHash {
				el.status.loser = append(el.status.loser, value.block)
			}
		}

		dps.acTrx.rollBack(el.status.loser)
		dps.acTrx.addWinner2Ledger(blk)
		el.updateVoteStatistic(confirmedHash)
		dps.dispatchAckedBlock(blk, confirmedHash, -1)
		dps.eb.Publish(common.EventConfirmedBlock, blk)
		el.cleanBlockInfo()
	} else {
		dps.logger.Infof("wait for enough rep vote for block [%s],current vote is [%s]", confirmedHash, balance)
	}
}

func (el *Election) tally(isSync bool) map[types.Hash]*BlockReceivedVotes {
	totals := make(map[types.Hash]*BlockReceivedVotes)
	var hash types.Hash

	el.vote.repVotes.Range(func(key, value interface{}) bool {
		hash = value.(*voteInfo).hash

		if _, ok := totals[hash]; !ok {
			if block, ok := el.blocks.Load(hash); ok {
				totals[hash] = &BlockReceivedVotes{
					block:   block.(*types.StateBlock),
					balance: types.ZeroBalance,
				}
			} else {
				return true
			}
		}

		var weight types.Balance
		repAddress := key.(types.Address)
		if !isSync {
			weight = el.dps.ledger.Weight(repAddress)
		} else {
			if w, ok := el.dps.totalVote[repAddress]; ok {
				weight = w
			} else {
				weight = el.dps.ledger.Weight(repAddress)
			}
		}

		totals[hash].balance = totals[hash].balance.Add(weight)
		el.dps.logger.Infof("rep[%s] ack block[%s] weight[%s]", repAddress, hash, weight)
		return true
	})

	return totals
}

func (el *Election) getOnlineRepresentativesBalance() types.Balance {
	b := types.ZeroBalance
	reps := el.dps.getOnlineRepresentatives()

	for _, addr := range reps {
		if b1, _ := el.dps.ledger.GetRepresentation(addr); b1 != nil {
			b = b.Add(b1.Total)
		}
	}

	return b
}

func (el *Election) getGenesisBalance() (types.Balance, error) {
	genesis := common.GenesisBlock()
	return genesis.Balance, nil
}

func (el *Election) ifValidAndSetInvalid() bool {
	return atomic.CompareAndSwapInt32(&el.valid, 1, 0)
}

func (el *Election) isValid() bool {
	return atomic.LoadInt32(&el.valid) == 1
}

func (el *Election) cleanBlockInfo() {
	el.blocks.Range(func(key, value interface{}) bool {
		h := key.(types.Hash)
		el.dps.hash2el.Delete(h)
		el.dps.unsubAckDo(h)
		return true
	})
}
