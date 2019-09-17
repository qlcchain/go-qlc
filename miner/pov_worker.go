package miner

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus/pov"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type PovWorker struct {
	miner  *Miner
	logger *zap.SugaredLogger

	maxTxPerBlock int
	minerAddr     types.Address
	minerAccount  *types.Account
	algoType      types.PovAlgoType
	cpuMining     bool

	mineBlockPool map[types.Hash]*types.PovMineBlock
	curMineBlock  *types.PovMineBlock
	preMineHeight uint64
	preMineTime   time.Time
	muxMineBlock  sync.Mutex

	quitCh chan struct{}
}

func NewPovWorker(miner *Miner) *PovWorker {
	worker := &PovWorker{
		miner:  miner,
		logger: log.NewLogger("pov_miner"),

		quitCh: make(chan struct{}),
	}
	worker.mineBlockPool = make(map[types.Hash]*types.PovMineBlock)

	return worker
}

func (w *PovWorker) Init() error {
	blkHeader := &types.PovHeader{
		AuxHdr: types.NewPovAuxHeader(),
		CbTx:   types.NewPovCoinBaseTx(1, 2),
	}
	blkHdrSize := blkHeader.Msgsize()
	blkHdrSize += 100 // cbtx extra

	hash := &types.Hash{}
	blkHdrSize += hash.Msgsize() * 32 * 2 // aux merkle branch + coinbase branch

	tx := &types.PovTransaction{}
	w.maxTxPerBlock = (common.PovChainBlockSize - blkHdrSize) / tx.Msgsize()
	w.logger.Infof("MaxBlockSize:%d, MaxHeaderSize:%d, MaxTxSize:%d, MaxTxNum:%d",
		common.PovChainBlockSize, blkHdrSize, tx.Msgsize(), w.maxTxPerBlock)

	w.algoType = types.ALGO_UNKNOWN

	return nil
}

func (w *PovWorker) Start() error {
	err := w.miner.eb.SubscribeSync(common.EventRpcSyncCall, w.OnEventRpcSyncCall)
	if err != nil {
		return err
	}

	return nil
}

func (w *PovWorker) Stop() error {
	_ = w.miner.eb.Unsubscribe(common.EventRpcSyncCall, w.OnEventRpcSyncCall)

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

func (w *PovWorker) GetMinerAccount() *types.Account {
	if w.minerAccount != nil {
		return w.minerAccount
	}

	if w.minerAddr.IsZero() {
		return nil
	}

	accounts := w.miner.GetPovEngine().GetAccounts()
	for _, account := range accounts {
		if account.Address() == w.minerAddr {
			w.minerAccount = account
			return w.minerAccount
		}
	}

	return nil
}

func (w *PovWorker) GetMinerAddress() types.Address {
	return w.minerAddr
}

func (w *PovWorker) GetAlgoType() types.PovAlgoType {
	return w.algoType
}

func (w *PovWorker) OnEventRpcSyncCall(name string, in interface{}, out interface{}) {
	switch name {
	case "Miner.GetWork":
		w.GetWork(in, out)
	case "Miner.SubmitWork":
		w.SubmitWork(in, out)
	case "Miner.StartMining":
		w.StartMining(in, out)
	case "Miner.StopMining":
		w.StopMining(in, out)
	case "Miner.GetMiningInfo":
		w.GetMiningInfo(in, out)
	}
}

func (w *PovWorker) GetWork(in interface{}, out interface{}) {
	inArgs := in.(map[interface{}]interface{})
	outArgs := out.(map[interface{}]interface{})

	if w.miner.GetSyncState() != common.Syncdone {
		outArgs["err"] = fmt.Errorf("miner pausing for sync state %s", w.miner.GetSyncState())
		return
	}

	minerAddr := inArgs["minerAddr"].(types.Address)
	algoName := inArgs["algoName"].(string)
	algoType := types.NewPoVHashAlgoFromStr(algoName)

	err := w.checkMinerPledge(minerAddr)
	if err != nil {
		outArgs["err"] = err
		return
	}

	mineBlock, err := w.generateBlock(minerAddr, algoType)
	if err != nil {
		outArgs["err"] = err
		return
	}

	outArgs["err"] = nil
	outArgs["mineBlock"] = mineBlock
}

func (w *PovWorker) SubmitWork(in interface{}, out interface{}) {
	inArgs := in.(map[interface{}]interface{})
	outArgs := out.(map[interface{}]interface{})

	result := inArgs["mineResult"].(*types.PovMineResult)

	mineBlock := w.findBlockInPool(result.WorkHash)
	if mineBlock == nil {
		outArgs["err"] = errors.New("failed to find block by WorkHash")
		return
	}

	err := w.checkAndFillBlockByResult(mineBlock, result)
	if err != nil {
		outArgs["err"] = err
		return
	}

	w.submitBlock(mineBlock)

	outArgs["err"] = nil
}

func (w *PovWorker) StartMining(in interface{}, out interface{}) {
	inArgs := in.(map[interface{}]interface{})
	outArgs := out.(map[interface{}]interface{})

	if w.cpuMining {
		outArgs["err"] = errors.New("cpu mining has been enabled already")
		return
	}

	minerAddr := inArgs["minerAddr"].(types.Address)
	algoName := inArgs["algoName"].(string)
	algoType := types.NewPoVHashAlgoFromStr(algoName)

	if algoType == types.ALGO_UNKNOWN {
		outArgs["err"] = errors.New("invalid algo name")
		return
	}

	err := w.checkMinerPledge(minerAddr)
	if err != nil {
		outArgs["err"] = err
		return
	}

	w.minerAddr = minerAddr
	w.algoType = algoType

	w.cpuMining = true
	common.Go(w.cpuMiningLoop)

	outArgs["err"] = nil
}

func (w *PovWorker) StopMining(in interface{}, out interface{}) {
	//inArgs := in.(map[interface{}]interface{})
	outArgs := out.(map[interface{}]interface{})

	if !w.cpuMining {
		outArgs["err"] = errors.New("cpu mining has been disabled already")
		return
	}

	w.cpuMining = false

	outArgs["err"] = nil
}

func (w *PovWorker) GetMiningInfo(in interface{}, out interface{}) {
	//inArgs := in.(map[interface{}]interface{})
	outArgs := out.(map[interface{}]interface{})

	latestBlock := w.GetChain().LatestBlock()

	outArgs["latestBlock"] = latestBlock
	outArgs["pooledTx"] = w.GetTxPool().GetPendingTxNum()

	outArgs["minerAddr"] = w.minerAddr
	outArgs["minerAlgo"] = w.algoType
	outArgs["cpuMining"] = w.cpuMining

	outArgs["err"] = nil
}

func (w *PovWorker) newBlockTemplate(minerAddr types.Address, algoType types.PovAlgoType) (*types.PovMineBlock, error) {
	latestHeader := w.GetChain().LatestHeader()

	mineBlock := types.NewPovMineBlock()

	// fill base header
	header := mineBlock.Header
	header.BasHdr.Version = types.POV_VBS_TOPBITS | uint32(algoType)
	header.BasHdr.Previous = latestHeader.GetHash()
	header.BasHdr.Height = latestHeader.GetHeight() + 1
	header.BasHdr.Timestamp = uint32(time.Now().Unix())

	prevStateHash := latestHeader.GetStateHash()
	prevStateTrie := w.GetChain().GetStateTrie(&prevStateHash)
	if prevStateTrie == nil {
		return nil, fmt.Errorf("failed to get prev state trie %s", prevStateHash)
	}

	// coinbase tx
	cbtx := header.CbTx

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

	err := w.GetPovConsensus().PrepareHeader(header)
	if err != nil {
		return nil, err
	}

	stateTrie, err := w.GetChain().GenStateTrie(header.GetHeight(), prevStateHash, accTxs)
	if err != nil {
		return nil, err
	}
	if stateTrie == nil {
		return nil, fmt.Errorf("failed to generate state trie")
	}

	// build coinbase tx
	cbtx.StateHash = *stateTrie.Hash()
	cbtx.TxNum = uint32(len(accTxs) + 1)

	minerRwd, repRwd := w.GetChain().CalcBlockReward(header)

	minerTxOut := cbtx.GetMinerTxOut()
	minerTxOut.Address = minerAddr
	minerTxOut.Value = minerRwd

	repTxOut := cbtx.GetRepTxOut()
	repTxOut.Address = types.MinerAddress
	repTxOut.Value = repRwd

	cbTxHash := cbtx.ComputeHash()

	// append all txs to body
	body := mineBlock.Body
	cbTxPov := &types.PovTransaction{Hash: cbTxHash, CbTx: cbtx}
	body.Txs = append(body.Txs, cbTxPov)
	body.Txs = append(body.Txs, accTxs...)

	// calc merkle root
	var mklTxHashList []*types.Hash
	mklTxHashList = append(mklTxHashList, &cbTxPov.Hash)
	mklTxHashList = append(mklTxHashList, accTxHashes...)
	mklHash := merkle.CalcMerkleTreeRootHash(mklTxHashList)
	header.BasHdr.MerkleRoot = mklHash

	mineBlock.AllTxHashes = mklTxHashList

	// calc merkle branch without coinbase tx
	mineBlock.CoinbaseBranch = merkle.BuildCoinbaseMerkleBranch(accTxHashes)

	mineBlock.WorkHash = mineBlock.Block.ComputeHash()
	mineBlock.MinTime = w.GetChain().CalcPastMedianTime(latestHeader)
	return mineBlock, nil
}

func (w *PovWorker) generateBlock(minerAddr types.Address, algoType types.PovAlgoType) (*types.PovMineBlock, error) {
	w.muxMineBlock.Lock()
	defer w.muxMineBlock.Unlock()

	var err error

	latestHeader := w.GetChain().LatestHeader()
	if w.curMineBlock == nil ||
		w.preMineHeight != latestHeader.GetHeight() ||
		time.Now().After(w.preMineTime.Add(30*time.Second)) {

		if w.preMineHeight != latestHeader.GetHeight() {
			w.curMineBlock = nil
			w.mineBlockPool = nil
		}

		w.curMineBlock, err = w.newBlockTemplate(minerAddr, algoType)
		if err != nil {
			return nil, err
		}

		if w.mineBlockPool == nil {
			w.mineBlockPool = make(map[types.Hash]*types.PovMineBlock)
		}
		w.mineBlockPool[w.curMineBlock.WorkHash] = w.curMineBlock

		w.preMineHeight = latestHeader.GetHeight()
		w.preMineTime = time.Now()
	}

	return w.curMineBlock, nil
}

func (w *PovWorker) findBlockInPool(workHash types.Hash) *types.PovMineBlock {
	w.muxMineBlock.Lock()
	defer w.muxMineBlock.Unlock()

	return w.mineBlockPool[workHash]
}

func (w *PovWorker) checkAndFillBlockByResult(mineBlock *types.PovMineBlock, result *types.PovMineResult) error {
	if len(result.CoinbaseExtra) < common.PovMinCoinbaseExtraSize {
		return fmt.Errorf("coinbase extra size too small, min size is %d", common.PovMinCoinbaseExtraSize)
	}

	if len(result.CoinbaseExtra) > common.PovMaxCoinbaseExtraSize {
		return fmt.Errorf("coinbase extra size too big, max size is %d", common.PovMaxCoinbaseExtraSize)
	}

	mineBlock.Header.CbTx.TxIns[0].Extra = result.CoinbaseExtra
	cbTxHash := mineBlock.Header.CbTx.ComputeHash()
	if cbTxHash.Cmp(result.CoinbaseHash) != 0 {
		return fmt.Errorf("coinbase hash not equal, %s != %s", cbTxHash, result.CoinbaseHash)
	}

	mineBlock.Header.CbTx.Hash = result.CoinbaseHash

	cbTxPov := mineBlock.Body.Txs[0]
	cbTxPov.Hash = result.CoinbaseHash
	mineBlock.AllTxHashes[0] = &cbTxPov.Hash

	calcMklRoot := merkle.CalcMerkleTreeRootHash(mineBlock.AllTxHashes)
	if calcMklRoot.Cmp(result.MerkleRoot) != 0 {
		return fmt.Errorf("merkle root not equal, %s != %s", calcMklRoot, result.MerkleRoot)
	}
	mineBlock.Header.BasHdr.MerkleRoot = result.MerkleRoot

	mineBlock.Header.BasHdr.Timestamp = result.Timestamp
	mineBlock.Header.BasHdr.Nonce = result.Nonce

	calcBlkHash := mineBlock.Header.ComputeHash()
	if calcBlkHash.Cmp(result.BlockHash) != 0 {
		return fmt.Errorf("block hash not equal, %s != %s", calcBlkHash, result.BlockHash)
	}
	mineBlock.Header.BasHdr.Hash = result.BlockHash

	if result.AuxPow != nil {
		calcParHash := result.AuxPow.ParBlockHeader.ComputeHash()
		if calcParHash != result.AuxPow.ParentHash {
			return fmt.Errorf("parent block hash not equal, %s != %s", calcParHash, result.AuxPow.ParentHash)
		}
		mineBlock.Header.AuxHdr = result.AuxPow
	}

	return nil
}

func (w *PovWorker) checkMinerPledge(minerAddr types.Address) error {
	latestBlock := w.GetChain().LatestBlock()

	if latestBlock.GetHeight() >= (common.PovMinerVerifyHeightStart - 1) {
		prevStateHash := latestBlock.GetStateHash()
		stateTrie := w.GetChain().GetStateTrie(&prevStateHash)
		rs := w.GetChain().GetRepState(stateTrie, minerAddr)
		if rs == nil {
			return errors.New("miner pausing for account state not exist")
		}
		if rs.Vote.Compare(common.PovMinerPledgeAmountMin) == types.BalanceCompSmaller {
			return errors.New("miner pausing for vote pledge not enough")
		}
	}

	return nil
}

func (w *PovWorker) cpuMiningLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	w.logger.Infof("running cpu mining loop, miner:%s, algo:%s", w.GetMinerAddress(), w.GetAlgoType())

	for {
		select {
		case <-w.quitCh:
			w.logger.Info("exiting cpu mining loop")
			return
		default:
			if !w.cpuMining {
				w.logger.Info("stopping cpu mining loop")
				return
			}

			if w.checkValidMiner() {
				w.mineNextBlock()
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

	minerAddr := w.GetMinerAddress()
	if minerAddr.IsZero() {
		w.logger.Warnf("miner pausing for miner account not exist")
		return false
	}

	latestBlock := w.GetChain().LatestBlock()

	tmNow := time.Now()
	if tmNow.Add(time.Hour).Unix() < int64(latestBlock.GetTimestamp()) {
		w.logger.Warnf("miner pausing for time now %d is older than latest block %d", tmNow.Unix(), latestBlock.GetTimestamp())
		return false
	}

	err := w.checkMinerPledge(minerAddr)
	if err != nil {
		w.logger.Warn(err)
		return false
	}

	return true
}

func (w *PovWorker) mineNextBlock() *types.PovMineBlock {
	latestHeader := w.GetChain().LatestHeader()

	w.logger.Debugf("try to generate block after latest %d/%s", latestHeader.GetHeight(), latestHeader.GetPrevious())

	mineBlock, err := w.newBlockTemplate(w.GetMinerAddress(), w.GetAlgoType())
	if err != nil {
		w.logger.Warnf("failed to generate block, err %s", err)
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	solveOK := w.solveBlock(mineBlock, ticker, w.quitCh)
	if !solveOK {
		return nil
	}

	w.submitBlock(mineBlock)
	return mineBlock
}

func (w *PovWorker) solveBlock(mineBlock *types.PovMineBlock, ticker *time.Ticker, quitCh chan struct{}) bool {
	loopBeginTime := time.Now()
	lastTxUpdateBegin := w.GetTxPool().LastUpdated()

	sealResultCh := make(chan *types.PovHeader)
	sealQuitCh := make(chan struct{})

	//w.logger.Debugf("before seal header %+v", genBlock.Header)

	err := w.GetPovConsensus().SealHeader(mineBlock.Header, sealQuitCh, sealResultCh)
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
				mineBlock.Header.CbTx.TxIns[0].Extra = make([]byte, len(resultHeader.CbTx.TxIns[0].Extra))
				copy(mineBlock.Header.CbTx.TxIns[0].Extra, resultHeader.CbTx.TxIns[0].Extra)
				mineBlock.Header.CbTx.Hash = mineBlock.Header.CbTx.ComputeHash()

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

func (w *PovWorker) submitBlock(mineBlock *types.PovMineBlock) {
	newBlock := mineBlock.Block.Copy()

	w.logger.Infof("submit block %d/%s", newBlock.GetHeight(), newBlock.GetHash())

	err := w.miner.GetPovEngine().AddMinedBlock(newBlock)
	if err != nil {
		w.logger.Infof("failed to submit block %d/%s", newBlock.GetHeight(), newBlock.GetHash())
	}
}
