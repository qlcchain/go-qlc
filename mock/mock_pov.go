package mock

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"math/big"
)

func GeneratePovBlock(prevBlock *types.PovBlock) (*types.PovBlock, *big.Int) {
	if prevBlock == nil {
		genesis := common.GenesisPovBlock()
		prevBlock = &genesis
	}

	prevTD := prevBlock.Target.ToBigInt()

	block := prevBlock.Clone()
	block.Previous = prevBlock.GetHash()
	block.Height = prevBlock.GetHeight()
	block.Hash = block.ComputeHash()

	nextTD := new(big.Int).Add(prevTD, prevTD)

	return block, nextTD
}
