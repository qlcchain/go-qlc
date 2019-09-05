package mock

import (
	"math/big"
	"sync"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
)

var povCoinbaseOnce sync.Once
var povCoinbaseAcc *types.Account

func GeneratePovCoinbase() *types.Account {
	if povCoinbaseAcc == nil {
		povCoinbaseOnce.Do(func() {
			povCoinbaseAcc = Account()
		})
	}

	return povCoinbaseAcc
}

func GeneratePovBlock(prevBlock *types.PovBlock, txNum uint32) (*types.PovBlock, *big.Int) {
	if prevBlock == nil {
		genesis := common.GenesisPovBlock()
		prevBlock = &genesis
	}

	prevTD := prevBlock.Target.ToBigInt()

	block := prevBlock.Clone()
	block.Timestamp = prevBlock.Timestamp + 1
	block.Previous = prevBlock.GetHash()
	block.Height = prevBlock.GetHeight() + 1

	if txNum > 0 {
		txHashes := make([]*types.Hash, 0, txNum)
		for txIdx := uint32(0); txIdx < txNum; txIdx++ {
			txBlk := StateBlockWithoutWork()
			txHash := txBlk.GetHash()
			txHashes = append(txHashes, &txHash)
			tx := &types.PovTransaction{Hash: txHash, Block: txBlk}
			block.Transactions = append(block.Transactions, tx)
		}
		block.TxNum = txNum
		block.MerkleRoot = merkle.CalcMerkleTreeRootHash(txHashes)
	}

	cb := GeneratePovCoinbase()
	block.Coinbase = cb.Address()
	block.VoteSignature = cb.Sign(block.ComputeVoteHash())

	block.Hash = block.ComputeHash()
	block.Signature = cb.Sign(block.Hash)

	nextTD := new(big.Int).Add(prevTD, prevTD)

	return block, nextTD
}
