package consensus

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

type tallyResult byte

const (
	vote tallyResult = iota
	changed
	confirm
)

type Votes struct {
	id        types.Hash                                //Previous block of fork
	rep_votes map[types.Address]*protos.ConfirmAckBlock // All votes received by account
}

func NewVotes(blk types.Block) *Votes {
	return &Votes{
		id:        blk.Root(),
		rep_votes: make(map[types.Address]*protos.ConfirmAckBlock),
	}
}
func (vs *Votes) voteExit(address types.Address) (bool, *protos.ConfirmAckBlock) {
	if v, ok := vs.rep_votes[address]; !ok {
		// Vote on this block hasn't been seen from rep before
		return false, nil
	} else {
		return true, v
	}
}
func (vs *Votes) voteStatus(vote_a *protos.ConfirmAckBlock) tallyResult {
	var result tallyResult
	if v, ok := vs.rep_votes[vote_a.Account]; !ok {
		result = vote
		vs.rep_votes[vote_a.Account] = vote_a
	} else {
		if v.Blk.GetHash() != vote_a.Blk.GetHash() {
			//Rep changed their vote
			result = changed
			vs.rep_votes[vote_a.Account] = vote_a
		} else {
			// Rep vote remained the same
			result = confirm
		}
	}
	return result
}
func (vs *Votes) uncontested() bool {
	var result bool
	var block types.Block
	if len(vs.rep_votes) != 0 {
		for _, v := range vs.rep_votes {
			block = v.Blk
			break
		}
		for _, v := range vs.rep_votes {
			result = v.Blk == block
		}
	}
	return result
}
