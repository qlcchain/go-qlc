package mock

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"math/big"
	"sync"
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

	prevTD := prevBlock.GetTargetInt()

	block := prevBlock.Clone()
	block.Header.BasHdr.Timestamp = prevBlock.GetTimestamp() + 1
	block.Header.BasHdr.Previous = prevBlock.GetHash()
	block.Header.BasHdr.Height = prevBlock.GetHeight() + 1

	cb := GeneratePovCoinbase()
	block.Header.CbTx.TxNum = txNum + 1
	block.Header.CbTx.StateHash = prevBlock.GetStateHash()
	block.Header.CbTx.Reward = common.PovMinerRewardPerBlockBalance
	block.Header.CbTx.CoinBase = cb.Address()
	block.Header.CbTx.Hash = block.Header.CbTx.ComputeHash()
	block.Header.CbTx.Signature = cb.Sign(block.Header.CbTx.Hash)

	txHashes := make([]*types.Hash, 0, txNum)

	txHash := block.Header.CbTx.Hash
	txHashes = append(txHashes, &txHash)
	tx := &types.PovTransaction{Hash: block.Header.CbTx.Hash, CbTx: &block.Header.CbTx}
	block.Body.Txs = append(block.Body.Txs, tx)

	if txNum > 0 {
		for txIdx := uint32(0); txIdx < txNum; txIdx++ {
			txBlk := StateBlockWithoutWork()
			txHash := txBlk.GetHash()
			txHashes = append(txHashes, &txHash)
			tx := &types.PovTransaction{Hash: txHash, Block: txBlk}
			block.Body.Txs = append(block.Body.Txs, tx)
		}
	}

	block.Header.BasHdr.MerkleRoot = merkle.CalcMerkleTreeRootHash(txHashes)
	block.Header.BasHdr.Hash = block.ComputeHash()

	nextTD := new(big.Int).Add(prevTD, prevTD)

	return block, nextTD
}
