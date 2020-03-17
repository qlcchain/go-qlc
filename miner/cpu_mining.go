package miner

import (
	"runtime"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
)

func (w *PovWorker) cpuMiningLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	w.logger.Infof("running cpu mining loop, miner:%s, algo:%s", w.GetMinerAddress(), w.GetAlgoType())

	nextBlkTicker := time.NewTicker(time.Second)
	defer nextBlkTicker.Stop()

	checkMinerTicker := time.NewTicker(time.Minute)
	defer checkMinerTicker.Stop()

	isMinerValid := w.checkValidMiner()

	for {
		select {
		case <-w.quitCh:
			w.logger.Info("exiting cpu mining loop")
			return

		case <-nextBlkTicker.C:
			if !w.cpuMining {
				w.logger.Info("stopping cpu mining loop")
				return
			}

			if isMinerValid {
				w.mineNextBlock()
			}

		case <-checkMinerTicker.C:
			isMinerValid = w.checkValidMiner()
		}
	}
}

func (w *PovWorker) checkValidMiner() bool {
	if w.miner.GetSyncState() != topic.SyncDone {
		w.logger.Infof("miner pausing for sync state %s", w.miner.GetSyncState())
		return false
	}

	minerAddr := w.GetMinerAddress()
	if minerAddr.IsZero() {
		w.logger.Warnf("miner pausing for miner account not exist")
		return false
	}

	latestBlock := w.miner.GetChain().LatestBlock()

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
	lastTxUpdateBegin := w.miner.GetTxPool().LastUpdated()

	sealResultCh := make(chan *types.PovHeader)
	sealQuitCh := make(chan struct{})

	//w.logger.Debugf("before seal header %+v", genBlock.Header)

	err := w.sealHeader(mineBlock.Header, sealQuitCh, sealResultCh)
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
			if !w.cpuMining {
				w.logger.Debugf("abort generate block because stopping cpu mining")
				break Loop
			}

			tmNow := time.Now()
			latestBlock := w.miner.GetChain().LatestBlock()
			if latestBlock.GetHash() != mineBlock.Header.GetPrevious() {
				w.logger.Debugf("abort generate block because latest block changed")
				break Loop
			}

			lastTxUpdateNow := w.miner.GetTxPool().LastUpdated()
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

func (w *PovWorker) sealHeader(header *types.PovHeader, quitCh chan struct{}, resultCh chan<- *types.PovHeader) error {
	var wgMine sync.WaitGroup

	abortCh := make(chan struct{})
	localCh := make(chan *types.PovHeader)

	for id := 1; id <= w.mineWorkerNum; id++ {
		wgMine.Add(1)
		go func(id int, gap int) {
			defer wgMine.Done()
			w.mineWorker(id, gap, header, abortCh, localCh)
		}(id, w.mineWorkerNum)
	}

	go func() {
		select {
		case <-quitCh:
			close(abortCh)
		case result := <-localCh:
			select {
			case resultCh <- result:
			default:
				// there's no hash
				w.logger.Warnf("failed to send sealing result to miner")
			}
			close(abortCh)
		}
		wgMine.Wait()
	}()

	return nil
}

func (w *PovWorker) mineWorker(id int, gap int, header *types.PovHeader, abortCh chan struct{}, localCh chan *types.PovHeader) {
	copyHdr := header.Copy()
	targetIntAlgo := copyHdr.GetAlgoTargetInt()

	tryCnt := 0
	for nonce := uint32(gap); nonce < uint32(0xFFFFFFFF); nonce += uint32(gap) {
		tryCnt++
		if tryCnt >= 100 {
			tryCnt = 0
			select {
			case <-abortCh:
				w.logger.Debugf("mine worker %d abort search nonce", id)
				localCh <- nil
				return
			default:
				//Non-blocking select to fall through
			}
		}

		copyHdr.BasHdr.Nonce = nonce

		powHash := copyHdr.ComputePowHash()
		powInt := powHash.ToBigInt()
		if powInt.Cmp(targetIntAlgo) <= 0 {
			w.logger.Debugf("mine worker %d found nonce %d", id, nonce)
			localCh <- copyHdr
			return
		}
	}

	w.logger.Debugf("mine worker %d exhaust nonce", id)
	localCh <- nil
}
