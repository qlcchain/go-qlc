package miner

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
	"runtime"
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
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for {
		select {
		case <-w.quitCh:
			w.logger.Info("Exiting PoV miner worker loop")
			return
		default:
			if w.checkValidMiner() {
				w.genNextBlock()
			} else {
				time.Sleep(time.Minute)
			}
		}
	}
}

func (w *PovWorker) checkValidMiner() bool {
	if w.miner.GetSyncState() != common.Syncdone {
		w.logger.Infof("miner pausing for sync state %s", w.miner.GetSyncState())
		return false
	}

	cbAccount := w.GetCoinbaseAccount()
	if cbAccount == nil {
		w.logger.Warnf("miner pausing for coinbase account not exist")
		return false
	}

	latestBlock := w.GetChain().LatestBlock()

	tmNow := time.Now()
	if tmNow.Add(time.Hour).Unix() < latestBlock.GetTimestamp() {
		w.logger.Warnf("miner pausing for time now %d is older than latest block %d", tmNow.Unix(), latestBlock.GetTimestamp())
		return false
	}

	if latestBlock.GetHeight() >= (common.PovMinerVerifyHeightStart - 1) {
		prevStateHash := latestBlock.GetStateHash()
		stateTrie := w.GetChain().GetStateTrie(&prevStateHash)
		as := w.GetChain().GetAccountState(stateTrie, cbAccount.Address())
		if as == nil || as.RepState == nil {
			w.logger.Warnf("miner pausing for account state not exist")
			return false
		}
		rs := as.RepState
		if rs.Vote.Compare(common.PovMinerPledgeAmountMin) == types.BalanceCompSmaller {
			w.logger.Warnf("miner pausing for vote pledge not enough")
			return false
		}
	}

	return true
}

func (w *PovWorker) genNextBlock() *types.PovBlock {
	latestBlock := w.GetChain().LatestBlock()

	target, err := w.GetChain().CalcNextRequiredTarget(latestBlock)
	if err != nil {
		return nil
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	current := &types.PovBlock{
		Previous:  latestBlock.GetHash(),
		Height:    latestBlock.GetHeight() + 1,
		Target:    target,
		Timestamp: time.Now().Unix(),
	}

	prevStateHash := latestBlock.GetStateHash()
	prevStateTrie := w.GetChain().GetStateTrie(&prevStateHash)
	if prevStateTrie == nil {
		w.logger.Errorf("failed to get prev state trie, err %s", prevStateHash, err)
		return nil
	}

	var mklTxHashList []*types.Hash
	accBlocks := w.GetTxPool().SelectPendingTxs(prevStateTrie, w.maxTxPerBlock)

	//w.logger.Debugf("current block %d select pending txs %d", len(accBlocks))

	for _, accBlock := range accBlocks {
		txPov := &types.PovTransaction{
			Hash:  accBlock.GetHash(),
			Block: accBlock,
		}
		mklTxHashList = append(mklTxHashList, &txPov.Hash)
		current.Transactions = append(current.Transactions, txPov)
	}
	current.TxNum = uint32(len(current.Transactions))

	mklHash := merkle.CalcMerkleTreeRootHash(mklTxHashList)
	current.MerkleRoot = mklHash

	stateTrie, err := w.GetChain().GenStateTrie(prevStateHash, current.Transactions)
	if err != nil {
		w.logger.Errorf("failed to generate state trie, err %s", err)
		return nil
	}
	if stateTrie == nil {
		w.logger.Errorf("failed to generate state trie, err nil")
		return nil
	}
	current.StateHash = *stateTrie.Hash()

	if w.solveBlock(current, ticker, w.quitCh) {
		w.submitBlock(current)
		return current
	} else {
		return nil
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
			if lastTxUpdateBegin != lastTxUpdateNow && time.Now().After(loopBeginTime.Add(time.Minute)) {
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

	block.Timestamp = time.Now().Unix()
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
