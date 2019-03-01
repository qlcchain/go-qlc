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
	id       types.Hash                                //Previous block of fork
	repVotes map[types.Address]*protos.ConfirmAckBlock // All votes received by account
}

func NewVotes(blk *types.StateBlock) *Votes {
	return &Votes{
		id:       blk.Root(),
		repVotes: make(map[types.Address]*protos.ConfirmAckBlock),
	}
}

func (vs *Votes) voteExit(address types.Address) (bool, *protos.ConfirmAckBlock) {
	if v, ok := vs.repVotes[address]; !ok {
		// Vote on this block hasn't been seen from rep before
		return false, nil
	} else {
		return true, v
	}
}

func (vs *Votes) voteStatus(va *protos.ConfirmAckBlock) tallyResult {
	var result tallyResult
	if v, ok := vs.repVotes[va.Account]; !ok {
		result = vote
		vs.repVotes[va.Account] = va
	} else {
		if v.Blk.GetHash() != va.Blk.GetHash() {
			//Rep changed their vote
			result = changed
			vs.repVotes[va.Account] = va
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
	if len(vs.repVotes) != 0 {
		for _, v := range vs.repVotes {
			block = v.Blk
			break
		}
		for _, v := range vs.repVotes {
			result = v.Blk == block
		}
	}
	return result
}
