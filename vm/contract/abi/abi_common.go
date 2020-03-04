package abi

import (
	"errors"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

// height begin from 0, so height + 1 == blocks count
func PovHeightToCount(height uint64) uint64 {
	return height + 1
}

func PovHeightRound(height uint64, round uint64) uint64 {
	roundCount := (PovHeightToCount(height) / round) * round
	if roundCount == 0 {
		return 0
	}
	return roundCount - 1
}

func PovGetNodeRewardHeightByDay(ctx *vmstore.VMContext) (uint64, error) {
	latestBlock, err := ctx.Ledger.GetLatestPovBlock()
	if err != nil || latestBlock == nil {
		return 0, errors.New("failed to get latest block")
	}

	nodeHeight := latestBlock.GetHeight()
	if nodeHeight < common.PovMinerRewardHeightStart {
		return 0, nil
	}
	if nodeHeight < common.PovMinerRewardHeightGapToLatest {
		return 0, nil
	}
	nodeHeight = nodeHeight - common.PovMinerRewardHeightGapToLatest

	nodeHeight = PovHeightRound(nodeHeight, uint64(common.POVChainBlocksPerDay))

	return nodeHeight, nil
}
