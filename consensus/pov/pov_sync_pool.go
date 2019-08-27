package pov

import (
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	maxSyncBlockInQue = 500
)

type PovSyncBlock struct {
	PeerID   string
	Height   uint64
	Block    *types.PovBlock
	TxExists map[types.Hash]struct{}
}

func (ss *PovSyncer) syncLoop() {
	forceTicker := time.NewTicker(forceSyncTimeInSec * time.Second)
	checkSyncTicker := time.NewTicker(1 * time.Second)
	checkChainTicker := time.NewTicker(10 * time.Second)
	requestSyncTicker := time.NewTicker(5 * time.Second)
	checkSyncPeerTicker := time.NewTicker(10 * time.Second)

	defer forceTicker.Stop()
	defer checkSyncTicker.Stop()
	defer checkChainTicker.Stop()

	for {
		select {
		case <-ss.quitCh:
			return

		case <-forceTicker.C:
			ss.onPeriodicSyncTimer()

		case <-checkSyncTicker.C:
			ss.onCheckSyncBlockTimer()

		case <-checkChainTicker.C:
			ss.onCheckChainTimer()

		case <-requestSyncTicker.C:
			ss.onRequestSyncTimer()

		case <-checkSyncPeerTicker.C:
			ss.onSyncPeerTimer()
		}
	}

	ss.logger.Infof("exit pov sync loop")
}

func (ss *PovSyncer) onPeriodicSyncTimer() {
	if ss.inSyncing.Load() == true {
		return
	}

	latestBlock := ss.chain.LatestBlock()
	if latestBlock == nil {
		ss.logger.Errorf("failed to get latest block")
		return
	}
	latestTD := ss.chain.GetBlockTDByHash(latestBlock.GetHash())
	if latestTD == nil {
		ss.logger.Errorf("failed to get latest td")
		return
	}

	bestPeer := ss.GetBestPeer("")
	if bestPeer == nil {
		ss.logger.Warnf("all peers are gone")
		return
	}

	syncOver := false
	if latestTD.Cmp(bestPeer.currentTD) >= 0 {
		syncOver = true
	} else if ss.absDiffHeight(latestBlock.GetHeight(), bestPeer.currentHeight) <= 3 {
		syncOver = true
	}
	if syncOver {
		ss.setInitState(common.Syncing)
		ss.setInitState(common.Syncdone)
		return
	}

	topPeers := ss.GetBestPeers(5)
	ss.logger.Infof("TopPeers: %d", len(topPeers))
	for _, peer := range topPeers {
		ss.logger.Infof("%s-%d-%s", peer.peerID, peer.currentHeight, peer.currentTD)
	}

	ss.inSyncing.Store(true)

	ss.setInitState(common.Syncing)

	ss.syncWithPeer(bestPeer)
}

func (ss *PovSyncer) onSyncPeerTimer() {
	if ss.inSyncing.Load() != true {
		return
	}

	syncErr := false

	syncPeer := ss.FindPeerWithStatus(ss.syncPeerID, peerStatusGood)
	if syncPeer == nil {
		ss.logger.Infof("sync peer %s is lost", ss.syncPeerID)

		syncErr = true
	} else if syncPeer.waitSyncRspMsg {
		if syncPeer.lastSyncReqTime.Add(time.Minute).Before(time.Now()) {
			ss.logger.Infof("sync peer %s may be too slow", ss.syncPeerID)
			syncPeer.status = peerStatusSlow

			syncErr = true
		}
	}

	if syncErr {
		ss.logger.Warnf("sync terminated with peer %s", ss.syncPeerID)
		ss.inSyncing.Store(false)
		ss.resetSyncPeer(syncPeer)
	}
}

func (ss *PovSyncer) onRequestSyncTimer() {
	if ss.inSyncing.Load() != true {
		return
	}

	syncPeer := ss.FindPeerWithStatus(ss.syncPeerID, peerStatusGood)
	if syncPeer == nil {
		ss.logger.Warnf("request syncing blocks but peer %s is gone", ss.syncPeerID)
		return
	}

	if syncPeer.waitLocatorRsp {
		ss.requestSyncingBlocks(syncPeer, true)
	} else {
		ss.requestSyncingBlocks(syncPeer, false)
	}
}

func (ss *PovSyncer) onCheckChainTimer() {
	if ss.inSyncing.Load() != true {
		return
	}

	latestBlock := ss.chain.LatestBlock()
	if latestBlock == nil {
		ss.logger.Errorf("failed to get latest block")
		return
	}

	if ss.syncCurHeight >= ss.syncToHeight && latestBlock.GetHeight() >= ss.syncToHeight {
		ss.logger.Infof("sync done, current height:%d", latestBlock.Height)
		ss.inSyncing.Store(false)
		ss.setInitState(common.Syncdone)
		return
	}

	ss.logger.Infof("syncCurHeight:%d, syncRcvHeight:%d, syncToHeight:%d, chainHeight:%d",
		ss.syncCurHeight, ss.syncRcvHeight, ss.syncToHeight, latestBlock.Height)
}

func (ss *PovSyncer) syncWithPeer(peer *PovSyncPeer) {
	ss.syncPeerID = peer.peerID
	ss.syncToHeight = peer.currentHeight
	ss.syncCurHeight = 0
	ss.syncRcvHeight = 0
	ss.syncReqHeight = 0
	ss.syncBlocks = make(map[uint64]*PovSyncBlock)

	ss.logger.Infof("sync starting with peer %s height %d", peer.peerID, peer.currentHeight)

	ss.requestSyncingBlocks(peer, true)
}

func (ss *PovSyncer) resetSyncPeer(peer *PovSyncPeer) {
	ss.syncPeerID = ""
	ss.syncToHeight = 0
	ss.syncCurHeight = 0
	ss.syncRcvHeight = 0
	ss.syncReqHeight = 0
	ss.syncBlocks = nil

	if peer != nil {
		peer.waitSyncRspMsg = false
		peer.lastSyncReqTime = time.Now()
		peer.waitLocatorRsp = false
	}
}

func (ss *PovSyncer) requestSyncingBlocks(syncPeer *PovSyncPeer, useLocator bool) {
	if ss.inSyncing.Load() != true {
		return
	}

	if len(ss.syncBlocks) >= maxSyncBlockInQue {
		ss.logger.Warnf("request syncing blocks but queue %d is full", len(ss.syncBlocks))
		return
	}

	if syncPeer.waitSyncRspMsg {
		if syncPeer.lastSyncReqTime.Add(15 * time.Second).After(time.Now()) {
			return
		}
	}

	req := new(protos.PovBulkPullReq)

	req.Reason = protos.PovReasonSync
	req.Count = maxSyncBlockPerReq
	if useLocator {
		req.Locators = ss.chain.GetBlockLocator(types.ZeroHash)
	} else {
		req.StartHeight = ss.syncRcvHeight + 1
		if req.StartHeight > ss.syncToHeight {
			return
		}
	}

	if useLocator {
		ss.logger.Infof("request syncing blocks use locators %d with peer %s", len(req.Locators), ss.syncPeerID)
	} else {
		ss.logger.Infof("request syncing blocks use height %d with peer %s", req.StartHeight, ss.syncPeerID)
	}

	ss.eb.Publish(common.EventSendMsgToSingle, p2p.PovBulkPullReq, req, ss.syncPeerID)

	syncPeer.lastSyncReqTime = time.Now()
	syncPeer.waitSyncRspMsg = true
	if useLocator {
		syncPeer.waitLocatorRsp = true
	}

	ss.syncReqHeight = req.StartHeight
}

func (ss *PovSyncer) onCheckSyncBlockTimer() {
	if ss.inSyncing.Load() != true {
		return
	}

	for height := ss.syncCurHeight; height <= ss.syncRcvHeight; height++ {
		syncBlk := ss.syncBlocks[height]
		if syncBlk == nil || syncBlk.Block == nil {
			return
		}

		hasTxPend := ss.checkSyncBlock(syncBlk)
		if hasTxPend {
			return
		}

		ss.eb.Publish(common.EventPovRecvBlock, syncBlk.Block, types.PovBlockFromRemoteSync, syncBlk.PeerID)
		ss.syncCurHeight = height + 1

		delete(ss.syncBlocks, height)
	}
}

func (ss *PovSyncer) addSyncBlock(block *types.PovBlock, peer *PovSyncPeer) {
	if ss.inSyncing.Load() != true {
		return
	}

	syncBlk := ss.syncBlocks[block.GetHeight()]
	if syncBlk == nil {
		syncBlk = &PovSyncBlock{Height: block.GetHeight(), Block: block, PeerID: peer.peerID}
		syncBlk.TxExists = make(map[types.Hash]struct{})
		ss.syncBlocks[block.GetHeight()] = syncBlk
	} else if syncBlk.Block != nil {
		if syncBlk.Block.GetHash() != block.GetHash() {
			syncBlk.Block = block
		}
	} else {
		syncBlk.Block = block
	}

	if peer.waitLocatorRsp {
		peer.waitLocatorRsp = false
		ss.syncCurHeight = block.GetHeight()
		ss.syncRcvHeight = block.GetHeight()

		ss.logger.Infof("got locator response, syncCurHeight %d", ss.syncCurHeight)
	} else {
		if block.GetHeight() == (ss.syncRcvHeight + 1) {
			ss.syncRcvHeight = block.GetHeight()
		}
	}
}

func (ss *PovSyncer) checkSyncBlock(syncBlk *PovSyncBlock) bool {
	var reqTxHashes []*types.Hash
	for _, tx := range syncBlk.Block.Transactions {
		txHash := tx.GetHash()
		ok, _ := ss.ledger.HasStateBlock(txHash)
		if ok {
			syncBlk.TxExists[txHash] = struct{}{}
		} else {
			reqTxHashes = append(reqTxHashes, &txHash)
		}
	}

	if uint32(len(syncBlk.TxExists)) >= syncBlk.Block.GetTxNum() {
		return false
	}

	if len(reqTxHashes) > 0 {
		ss.requestTxsByHashes(reqTxHashes, syncBlk.PeerID)
	}
	return true
}
