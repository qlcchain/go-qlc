package mock

import (
	"math"
	"math/big"
	"sync"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
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

func GenerateGenesisPovBlock() (*types.PovBlock, *types.PovTD) {
	genBlk := common.GenesisPovBlock()

	curWorkAlgo := types.CalcWorkIntToBigNum(genBlk.Header.GetAlgoTargetInt())
	genTD := new(types.PovTD)
	genTD.Chain = *curWorkAlgo

	return &genBlk, genTD
}

func GeneratePovBlockByFakePow(prevBlock *types.PovBlock, txNum uint32) (*types.PovBlock, *types.PovTD) {
	return doGeneratePovBlock(prevBlock, txNum, true)
}

func GeneratePovBlock(prevBlock *types.PovBlock, txNum uint32) (*types.PovBlock, *types.PovTD) {
	return doGeneratePovBlock(prevBlock, txNum, false)
}

func doGeneratePovBlock(prevBlock *types.PovBlock, txNum uint32, fakePow bool) (*types.PovBlock, *types.PovTD) {
	genBlk, genTD := GenerateGenesisPovBlock()
	if prevBlock == nil {
		prevBlock = genBlk
	}

	block := prevBlock.Clone()
	block.Body.Txs = nil

	block.Header.BasHdr.Timestamp = prevBlock.GetTimestamp() + 1
	block.Header.BasHdr.Previous = prevBlock.GetHash()
	block.Header.BasHdr.Height = prevBlock.GetHeight() + 1

	if fakePow {
		block.Header.BasHdr.Nonce = block.Header.BasHdr.Timestamp
	}

	cb := GeneratePovCoinbase()
	block.Header.CbTx.TxNum = txNum + 1
	block.Header.CbTx.StateHash = prevBlock.GetStateHash()
	block.Header.CbTx.TxOuts[0].Address = cb.Address()
	block.Header.CbTx.TxOuts[0].Value = types.NewBalance(int64(common.PovMinerRewardPerBlock * uint64(common.PovMinerRewardRatioMiner) / uint64(100)))
	block.Header.CbTx.TxOuts[1].Address = types.RepAddress
	block.Header.CbTx.TxOuts[1].Value = types.NewBalance(int64(common.PovMinerRewardPerBlock * uint64(common.PovMinerRewardRatioRep) / uint64(100)))
	block.Header.CbTx.Hash = block.Header.CbTx.ComputeHash()

	txHashes := make([]*types.Hash, 0, txNum)

	txHash := block.Header.CbTx.Hash
	txHashes = append(txHashes, &txHash)
	tx := &types.PovTransaction{Hash: block.Header.CbTx.Hash, CbTx: block.Header.CbTx}
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

	nextTD := types.NewPovTD()
	tdUint := genTD.Chain.Int.Uint64() * block.GetHeight()
	nextTD.Chain.AddBigInt(big.NewInt(0), big.NewInt(int64(tdUint)))

	return block, nextTD
}

func UpdatePovHash(povBlk *types.PovBlock) {
	povBlk.Header.CbTx.Hash = povBlk.Header.CbTx.ComputeHash()
	povBlk.Body.Txs[0].Hash = povBlk.Header.CbTx.Hash

	allTxs := povBlk.GetAllTxs()
	txHashes := make([]*types.Hash, 0, len(allTxs))

	for _, tx := range povBlk.GetAllTxs() {
		txHash := tx.GetHash()
		txHashes = append(txHashes, &txHash)
	}

	povBlk.Header.BasHdr.MerkleRoot = merkle.CalcMerkleTreeRootHash(txHashes)
	povBlk.Header.BasHdr.Hash = povBlk.ComputeHash()
}

func PovHeader() *types.PovHeader {
	i, _ := random.Intn(math.MaxInt16)
	return &types.PovHeader{
		BasHdr: types.PovBaseHeader{
			Hash:       Hash(),
			Previous:   Hash(),
			MerkleRoot: Hash(),
			Height:     uint64(i),
			//Coinbase:   Address(),
			//StateHash:  Hash(),
		},
	}
}
