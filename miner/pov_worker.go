package miner

import (
	"runtime"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus/pov"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

const (
	maxNonce = ^uint64(0) // 2^64 - 1
)

type PovWorker struct {
	miner  *Miner
	logger *zap.SugaredLogger

	maxTxPerBlock   int
	coinbaseAddress *types.Address
	algo            types.PovAlgoType

	quitCh chan struct{}
}

type PovMineBlock struct {
	Header *types.PovHeader
	Body   *types.PovBody
	Block  *types.PovBlock
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
	cbtx := &types.PovCoinBaseTx{}

	tx := &types.PovTransaction{}
	w.maxTxPerBlock = (common.PovChainBlockSize - blkHeader.Msgsize() - cbtx.Msgsize()) / tx.Msgsize()

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

	w.algo = w.GetAlgo()

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

func (w *PovWorker) GetTxPool() *pov.PovTxPool {
	return w.miner.GetPovEngine().GetTxPool()
}

func (w *PovWorker) GetChain() *pov.PovBlockChain {
	return w.miner.GetPovEngine().GetChain()
}

func (w *PovWorker) GetPovConsensus() pov.ConsensusPov {
	return w.miner.GetPovEngine().GetConsensus()
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

func (w *PovWorker) GetAlgo() types.PovAlgoType {
	algo := types.NewPoVHashAlgoFromStr(w.miner.cfg.PoV.MinerAlgo)
	if algo != types.ALGO_UNKNOWN {
		return algo
	}
	return types.ALGO_SHA256D
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
	if tmNow.Add(time.Hour).Unix() < int64(latestBlock.GetTimestamp()) {
		w.logger.Warnf("miner pausing for time now %d is older than latest block %d", tmNow.Unix(), latestBlock.GetTimestamp())
		return false
	}

	if latestBlock.GetHeight() >= (common.PovMinerVerifyHeightStart - 1) {
		prevStateHash := latestBlock.GetStateHash()
		stateTrie := w.GetChain().GetStateTrie(&prevStateHash)
		rs := w.GetChain().GetRepState(stateTrie, cbAccount.Address())
		if rs == nil {
			w.logger.Warnf("miner pausing for account state not exist")
			return false
		}
		if rs.Vote.Compare(common.PovMinerPledgeAmountMin) == types.BalanceCompSmaller {
			w.logger.Warnf("miner pausing for vote pledge not enough")
			return false
		}
	}

	return true
}

func (w *PovWorker) genNextBlock() *PovMineBlock {
	latestHeader := w.GetChain().LatestHeader()

	w.logger.Debugf("try to generate block after latest %d/%s", latestHeader.GetHeight(), latestHeader.GetPrevious())

	mineBlock := &PovMineBlock{}

	mineBlock.Header = types.NewPovHeader()
	mineBlock.Header.BasHdr.Version = 0 | uint32(w.algo)
	mineBlock.Header.BasHdr.Previous = latestHeader.GetHash()
	mineBlock.Header.BasHdr.Height = latestHeader.GetHeight() + 1
	mineBlock.Header.BasHdr.Timestamp = uint32(time.Now().Unix())

	mineBlock.Body = types.NewPovBody()

	prevStateHash := latestHeader.GetStateHash()
	prevStateTrie := w.GetChain().GetStateTrie(&prevStateHash)
	if prevStateTrie == nil {
		w.logger.Errorf("failed to get prev state trie", prevStateHash)
		return nil
	}

	// coinbase tx
	cbtx := mineBlock.Header.CbTx

	// pack account block txs
	accBlocks := w.GetTxPool().SelectPendingTxs(prevStateTrie, w.maxTxPerBlock)

	//w.logger.Debugf("current block %d select pending txs %d", len(accBlocks))

	var accTxHashes []*types.Hash
	var accTxs []*types.PovTransaction
	for _, accBlock := range accBlocks {
		accTx := &types.PovTransaction{
			Hash:  accBlock.GetHash(),
			Block: accBlock,
		}
		accTxHashes = append(accTxHashes, &accTx.Hash)
		accTxs = append(accTxs, accTx)
	}

	err := w.GetPovConsensus().PrepareHeader(mineBlock.Header)
	if err != nil {
		w.logger.Errorf("failed to prepare header, err %s", err)
		return nil
	}

	stateTrie, err := w.GetChain().GenStateTrie(prevStateHash, accTxs)
	if err != nil {
		w.logger.Errorf("failed to generate state trie, err %s", err)
		return nil
	}
	if stateTrie == nil {
		w.logger.Errorf("failed to generate state trie, err nil")
		return nil
	}

	// build coinbase tx
	cbtx.StateHash = *stateTrie.Hash()
	cbtx.TxNum = uint32(len(accTxs) + 1)

	minerRwd, repRwd := w.GetChain().CalcBlockReward(mineBlock.Header)

	minerTxOut := cbtx.GetMinerTxOut()
	minerTxOut.Address = w.GetCoinbaseAccount().Address()
	minerTxOut.Value = minerRwd

	repTxOut := cbtx.GetRepTxOut()
	repTxOut.Address = types.MinerAddress
	repTxOut.Value = repRwd

	cbtxHash := cbtx.ComputeHash()

	// append all txs to body
	mineBlock.Body.Txs = append(mineBlock.Body.Txs, []*types.PovTransaction{{Hash: cbtxHash, CbTx: cbtx}}...)
	mineBlock.Body.Txs = append(mineBlock.Body.Txs, accTxs...)

	// calc merkle root
	var mklTxHashList []*types.Hash
	mklTxHashList = append(mklTxHashList, []*types.Hash{&cbtxHash}...)
	mklTxHashList = append(mklTxHashList, accTxHashes...)
	mklHash := merkle.CalcMerkleTreeRootHash(mklTxHashList)
	mineBlock.Header.BasHdr.MerkleRoot = mklHash

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	if w.solveBlock(mineBlock, ticker, w.quitCh) {
		w.submitBlock(mineBlock)
		return mineBlock
	} else {
		return nil
	}
}

func (w *PovWorker) solveBlock(mineBlock *PovMineBlock, ticker *time.Ticker, quitCh chan struct{}) bool {
	cbAccount := w.GetCoinbaseAccount()

	loopBeginTime := time.Now()
	lastTxUpdateBegin := w.GetTxPool().LastUpdated()

	sealResultCh := make(chan *types.PovHeader)
	sealQuitCh := make(chan struct{})

	//w.logger.Debugf("before seal header %+v", genBlock.Header)

	err := w.GetPovConsensus().SealHeader(mineBlock.Header, cbAccount, sealQuitCh, sealResultCh)
	if err != nil {
		w.logger.Errorf("failed to seal header, err %s", err)
		return false
	}

	foundNonce := false
Loop:
	for {
		select {
		case <-w.quitCh:
			break Loop

		case resultHeader := <-sealResultCh:
			if resultHeader != nil {
				foundNonce = true

				// fill coinbase tx
				mineBlock.Header.CbTx.Extra = make([]byte, len(resultHeader.CbTx.Extra))
				copy(mineBlock.Header.CbTx.Extra, resultHeader.CbTx.Extra)
				mineBlock.Header.CbTx.Hash = mineBlock.Header.CbTx.ComputeHash()
				mineBlock.Header.CbTx.Signature = cbAccount.Sign(mineBlock.Header.CbTx.Hash)

				mineBlock.Body.Txs[0].Hash = mineBlock.Header.CbTx.Hash

				// fill block header
				mineBlock.Header.BasHdr.Timestamp = resultHeader.BasHdr.Timestamp
				mineBlock.Header.BasHdr.MerkleRoot = resultHeader.BasHdr.MerkleRoot
				mineBlock.Header.BasHdr.Nonce = resultHeader.BasHdr.Nonce
				mineBlock.Header.BasHdr.Hash = mineBlock.Header.ComputeHash()
			}
			break Loop

		case <-ticker.C:
			tmNow := time.Now()
			latestBlock := w.GetChain().LatestBlock()
			if latestBlock.GetHash() != mineBlock.Header.GetPrevious() {
				w.logger.Debugf("abort generate block because latest block changed")
				break Loop
			}

			lastTxUpdateNow := w.GetTxPool().LastUpdated()
			if lastTxUpdateBegin != lastTxUpdateNow && tmNow.After(loopBeginTime.Add(time.Minute)) {
				w.logger.Debugf("abort generate block because tx pool changed")
				break Loop
			}

			if tmNow.After(loopBeginTime.Add(time.Duration(common.PovMinerMaxFindNonceTimeSec) * time.Second)) {
				w.logger.Debugf("abort generate block because exceed max timeout")
				break Loop
			}
		}
	}

	//w.logger.Debugf("after seal header %+v", genBlock.Header)

	close(sealQuitCh)

	if !foundNonce {
		return false
	}

	return true
}

func (w *PovWorker) submitBlock(genBlock *PovMineBlock) {
	genBlock.Block = types.NewPovBlockWithBody(genBlock.Header, genBlock.Body)

	w.logger.Infof("submit block %d/%s", genBlock.Block.GetHeight(), genBlock.Block.GetHash())

	err := w.miner.GetPovEngine().AddMinedBlock(genBlock.Block)
	if err != nil {
		w.logger.Infof("failed to submit block %d/%s", genBlock.Block.GetHeight(), genBlock.Block.GetHash())
	}
}
