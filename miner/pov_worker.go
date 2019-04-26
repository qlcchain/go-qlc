package miner

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
	"time"
)

const (
	maxNonce = ^uint64(0) // 2^64 - 1
)

type PovWorker struct {
	miner  *Miner
	logger *zap.SugaredLogger

	maxTxPerBlock   int
	coinbaseAddress *types.Address

	quitCh chan struct{}
}

func NewPovWorker(miner *Miner) *PovWorker {
	worker := &PovWorker{
		miner:  miner,
		logger: log.NewLogger("pov_miner"),

		quitCh: make(chan struct{}),
	}

	return worker
}

func (w *PovWorker) Init() error {
	var err error

	blkHeader := &types.PovHeader{}
	tx := &types.PovTransaction{}
	w.maxTxPerBlock = (w.GetConfig().PoV.BlockSize - blkHeader.Msgsize()) / tx.Msgsize()

	cbAddress, err := types.HexToAddress(w.GetConfig().PoV.Coinbase)
	if err != nil {
		w.logger.Errorf("invalid coinbase address %s", w.GetConfig().PoV.Coinbase)
	} else if cbAddress.IsZero() {
		w.logger.Errorf("coinbase address is zero")
	} else {
		w.coinbaseAddress = &cbAddress
	}

	return nil
}

func (w *PovWorker) Start() error {
	if w.coinbaseAddress != nil {
		cbAccount := w.GetCoinbaseAccount()
		if cbAccount == nil {
			w.logger.Errorf("coinbase %s account not exist", w.coinbaseAddress)
		} else {
			common.Go(w.loop)
		}
	}

	return nil
}

func (w *PovWorker) Stop() error {
	if w.quitCh != nil {
		close(w.quitCh)
	}

	return nil
}

func (w *PovWorker) GetConfig() *config.Config {
	return w.miner.GetConfig()
}

func (w *PovWorker) GetTxPool() *consensus.PovTxPool {
	return w.miner.GetPovEngine().GetTxPool()
}

func (w *PovWorker) GetChain() *consensus.PovBlockChain {
	return w.miner.GetPovEngine().GetChain()
}

func (w *PovWorker) GetCoinbaseAccount() *types.Account {
	accounts := w.miner.GetPovEngine().GetAccounts()
	for _, account := range accounts {
		if account.Address() == *w.coinbaseAddress {
			return account
		}
	}
	return nil
}

func (w *PovWorker) loop() {
	for {
		select {
		case <-w.quitCh:
			w.logger.Errorf("Exiting PoV miner worker loop")
		default:
			w.genNextBlock()
		}
	}
}

func (w *PovWorker) genNextBlock() {
	if w.miner.GetSyncState() != common.Syncdone {
		time.Sleep(time.Second)
		return
	}

	cbAccount := w.GetCoinbaseAccount()
	if cbAccount == nil {
		time.Sleep(time.Minute)
		return
	}

	ticker := time.NewTicker(time.Second)

	parent := w.GetChain().LatestBlock()

	target, err := w.GetChain().CalcNextRequiredTarget(parent)
	if err != nil {
		return
	}

	current := &types.PovBlock{
		Previous:  parent.GetHash(),
		Height:    parent.GetHeight() + 1,
		Timestamp: time.Now().Unix(),
		Target:    target,
	}

	var mklTxHashList []*types.Hash
	accBlocks := w.GetTxPool().SelectPendingTxs(w.maxTxPerBlock)
	for _, accBlock := range accBlocks {
		txPov := &types.PovTransaction{
			Address: accBlock.Address,
			Hash:    accBlock.GetHash(),
		}
		mklTxHashList = append(mklTxHashList, &txPov.Hash)
		current.Transactions = append(current.Transactions, txPov)
	}
	current.TxNum = uint32(len(current.Transactions))

	mklHash := merkle.CalcMerkleTreeRootHash(mklTxHashList)
	current.MerkleRoot = mklHash

	if w.solveBlock(current, ticker, w.quitCh) {
		w.submitBlock(current)
	}
}

func (w *PovWorker) solveBlock(block *types.PovBlock, ticker *time.Ticker, quitCh chan struct{}) bool {
	cbAccount := w.GetCoinbaseAccount()
	targetInt := block.Target.ToBigInt()

	block.Coinbase = cbAccount.Address()

	loopBeginTime := time.Now()
	lastTxUpdateBegin := w.GetTxPool().LastUpdated()

	foundNonce := false
	for nonce := uint64(0); nonce < maxNonce; nonce++ {
		select {
		case <-w.quitCh:
			return false

		case <-ticker.C:
			latestBlock := w.GetChain().LatestBlock()
			if latestBlock.Hash != block.GetPrevious() {
				return false
			}

			lastTxUpdateNow := w.GetTxPool().LastUpdated()
			if lastTxUpdateBegin != lastTxUpdateNow && time.Now().After(loopBeginTime) {
				return false
			}

		default:
			//Non-blocking select to fall through
		}

		block.Nonce = nonce

		voteHash := block.ComputeVoteHash()
		voteSignature := cbAccount.Sign(voteHash)
		voteSigInt := voteSignature.ToBigInt()
		if voteSigInt.Cmp(targetInt) <= 0 {
			foundNonce = true
			block.VoteSignature = voteSignature
			break
		}
	}

	if !foundNonce {
		return false
	}

	block.Hash = block.ComputeHash()
	block.Signature = cbAccount.Sign(block.Hash)

	return true
}

func (w *PovWorker) submitBlock(block *types.PovBlock) {
	w.logger.Infof("submit block %d/%s", block.GetHeight(), block.GetHash())

	err := w.miner.GetPovEngine().AddMinedBlock(block)
	if err != nil {
		w.logger.Infof("failed to submit block %d/%s", block.GetHeight(), block.GetHash())
	}
}