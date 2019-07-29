package dpos

import (
	"sync"

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
	id       voteKey   //Previous block of fork
	repVotes *sync.Map // All votes received by account
}

func newVotes(blk *types.StateBlock) *Votes {
	return &Votes{
		id:       getVoteKey(blk),
		repVotes: new(sync.Map),
	}
}

func (vs *Votes) voteExit(address types.Address) (bool, *protos.ConfirmAckBlock) {
	if v, ok := vs.repVotes.Load(address); !ok {
		return false, nil
	} else {
		return true, v.(*protos.ConfirmAckBlock)
	}
}

func (vs *Votes) voteStatus(va *protos.ConfirmAckBlock) tallyResult {
	var result tallyResult

	if v, ok := vs.repVotes.Load(va.Account); !ok {
		result = vote
		vs.repVotes.Store(va.Account, va)
	} else {
		if v.(*protos.ConfirmAckBlock).Blk.GetHash() != va.Blk.GetHash() {
			//Rep changed their vote
			result = changed
			vs.repVotes.Delete(va.Account)
			vs.repVotes.Store(va.Account, va)
		} else {
			// Rep vote remained the same
			result = confirm
		}
	}

	return result
}
