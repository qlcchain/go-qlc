package pov

import (
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	checkPeerStatusTime = 30
	forceSyncTimeInSec  = 60

	maxSyncBlockPerReq = 1000
	maxPullBlockPerReq = 1000
	maxSyncBlockInQue  = 3000
)

type PovSyncerChainReader interface {
	GenesisBlock() *types.PovBlock
	LatestBlock() *types.PovBlock
	GetBlockLocator(hash types.Hash) []*types.Hash
	LocateBestBlock(locator []*types.Hash) *types.PovBlock
	GetBlockTDByHash(hash types.Hash) *types.PovTD
}

type PovSyncer struct {
	eb         event.EventBus
	handlerIds map[common.TopicType]string // topic->handler id
	ledger     ledger.Store
	chain      PovSyncerChainReader

	logger   *zap.SugaredLogger
	allPeers sync.Map // map[string]*PovSyncPeer

	initSyncOver  *atomic.Bool
	initSyncState common.SyncState
	inSyncing     *atomic.Bool
	syncStartTime time.Time
	syncEndTime   time.Time

	syncSeqID     *atomic.Uint32
	syncPeerID    string
	syncToHeight  uint64
	syncCurHeight uint64
	syncRcvHeight uint64
	syncReqHeight uint64
	syncBlocks    map[uint64]*PovSyncBlock
	syncBlocksMux sync.RWMutex

	lastReqTxTime *atomic.Int64 // time.Time.Unix()

	messageCh chan *PovSyncMessage
	eventCh   chan *PovSyncEvent
	quitCh    chan struct{}
}

type PovSyncMessage struct {
	msgValue interface{}
	msgPeer  string
}

type PovSyncEvent struct {
	eventType common.TopicType
	eventData interface{}
}

func NewPovSyncer(eb event.EventBus, l ledger.Store, chain PovSyncerChainReader) *PovSyncer {
	ss := &PovSyncer{
		eb:            eb,
		ledger:        l,
		chain:         chain,
		initSyncState: common.SyncNotStart,
		messageCh:     make(chan *PovSyncMessage, 2000),
		eventCh:       make(chan *PovSyncEvent, 200),
		quitCh:        make(chan struct{}),
		logger:        log.NewLogger("pov_sync"),
		handlerIds:    make(map[common.TopicType]string),
		lastReqTxTime: atomic.NewInt64(0),
		syncSeqID:     atomic.NewUint32(0),
		initSyncOver:  atomic.NewBool(false),
		inSyncing:     atomic.NewBool(false),
		syncBlocks:    make(map[uint64]*PovSyncBlock),
	}
	return ss
}

func (ss *PovSyncer) Start() error {
	eb := ss.eb
	if eb != nil {
		if id, err := eb.SubscribeSync(common.EventAddP2PStream, ss.onAddP2PStream); err != nil {
			ss.logger.Error(err)
			return err
		} else {
			ss.handlerIds[common.EventAddP2PStream] = id
		}
		if id, err := eb.SubscribeSync(common.EventDeleteP2PStream, ss.onDeleteP2PStream); err != nil {
			ss.logger.Error(err)
			return err
		} else {
			ss.handlerIds[common.EventDeleteP2PStream] = id
		}
		if id, err := eb.SubscribeSync(common.EventPovPeerStatus, ss.onPovStatus); err != nil {
			ss.logger.Error(err)
			return err
		} else {
			ss.handlerIds[common.EventPovPeerStatus] = id
		}
		if id, err := eb.Subscribe(common.EventPovBulkPullReq, ss.onPovBulkPullReq); err != nil {
			ss.logger.Error(err)
			return err
		} else {
			ss.handlerIds[common.EventPovBulkPullReq] = id
		}
		if id, err := eb.Subscribe(common.EventPovBulkPullRsp, ss.onPovBulkPullRsp); err != nil {
			ss.logger.Error(err)
			return err
		} else {
			ss.handlerIds[common.EventPovBulkPullRsp] = id
		}
	}

	common.Go(ss.mainLoop)
	common.Go(ss.syncLoop)

	return nil
}

func (ss *PovSyncer) Stop() {
	eb := ss.eb
	if eb != nil {
		for k, v := range ss.handlerIds {
			if err := eb.Unsubscribe(k, v); err != nil {
				ss.logger.Error(err)
			}
		}
	}

	close(ss.quitCh)
}

func (ss *PovSyncer) mainLoop() {
	checkPeerTicker := time.NewTicker(checkPeerStatusTime * time.Second)

	for {
		select {
		case <-ss.quitCh:
			return

		case <-checkPeerTicker.C:
			ss.checkAllPeers()

		case msg := <-ss.messageCh:
			ss.processMessage(msg)

		case evt := <-ss.eventCh:
			ss.processEvent(evt)
		}
	}
}

func (ss *PovSyncer) onPovBulkPullReq(req *protos.PovBulkPullReq, msgPeer string) {
	ss.messageCh <- &PovSyncMessage{msgValue: req, msgPeer: msgPeer}
}

func (ss *PovSyncer) onPovBulkPullRsp(rsp *protos.PovBulkPullRsp, msgPeer string) {
	ss.messageCh <- &PovSyncMessage{msgValue: rsp, msgPeer: msgPeer}
}

func (ss *PovSyncer) processMessage(msg *PovSyncMessage) {
	switch v := msg.msgValue.(type) {
	case *protos.PovBulkPullReq:
		ss.processPovBulkPullReq(msg)
	case *protos.PovBulkPullRsp:
		ss.processPovBulkPullRsp(msg)
	default:
		ss.logger.Infof("unknown message value type %T!\n", v)
	}
}

func (ss *PovSyncer) processPovBulkPullReq(msg *PovSyncMessage) {
	req := msg.msgValue.(*protos.PovBulkPullReq)

	if req.PullType == protos.PovPullTypeForward {
		ss.processPovBulkPullReqByForward(msg)
		return
	} else if req.PullType == protos.PovPullTypeBackward {
		ss.processPovBulkPullReqByBackward(msg)
		return
	} else if req.PullType == protos.PovPullTypeBatch {
		ss.processPovBulkPullReqByBatch(msg)
		return
	} else {
		ss.logger.Infof("recv PovBulkPullReq by unknown type %d", req.PullType)
	}
}

func (ss *PovSyncer) processPovBulkPullReqByForward(msg *PovSyncMessage) {
	req := msg.msgValue.(*protos.PovBulkPullReq)

	if req.Reason == protos.PovReasonSync {
		if len(req.Locators) > 0 {
			ss.logger.Infof("recv PovBulkPullReq by forward from peer %s, reason %d locator %s count %d",
				msg.msgPeer, req.Reason, req.Locators[0], req.Count)
		} else if !req.StartHash.IsZero() {
			ss.logger.Infof("recv PovBulkPullReq by forward from peer %s, reason %d hash %s count %d",
				msg.msgPeer, req.Reason, req.StartHash, req.Count)
		} else {
			ss.logger.Infof("recv PovBulkPullReq by forward from peer %s, reason %d height %d count %d",
				msg.msgPeer, req.Reason, req.StartHeight, req.Count)
		}
	} else {
		if len(req.Locators) > 0 {
			ss.logger.Debugf("recv PovBulkPullReq by forward from peer %s, reason %d locator %s count %d",
				msg.msgPeer, req.Reason, req.Locators[0], req.Count)
		} else if !req.StartHash.IsZero() {
			ss.logger.Debugf("recv PovBulkPullReq by forward from peer %s, reason %d hash %s count %d",
				msg.msgPeer, req.Reason, req.StartHash, req.Count)
		} else {
			ss.logger.Debugf("recv PovBulkPullReq by forward from peer %s, reason %d height %d count %d",
				msg.msgPeer, req.Reason, req.StartHeight, req.Count)
		}
	}

	rsp := new(protos.PovBulkPullRsp)
	rsp.Reason = req.Reason

	startHeight := req.StartHeight
	blockCount := req.Count
	if blockCount == 0 {
		blockCount = maxSyncBlockPerReq
	}
	if len(req.Locators) > 0 {
		block := ss.chain.LocateBestBlock(req.Locators)
		if block == nil {
			ss.logger.Debugf("failed to locate best block %s", req.Locators[0])
			return
		}
		rsp.Blocks = append(rsp.Blocks, block)
		startHeight = block.GetHeight() + 1
		blockCount--
	} else if !req.StartHash.IsZero() {
		block, _ := ss.ledger.GetPovBlockByHash(req.StartHash)
		if block == nil {
			ss.logger.Debugf("failed to get block by hash %s", req.StartHash)
			return
		}
		rsp.Blocks = append(rsp.Blocks, block)
		startHeight = block.GetHeight() + 1
		blockCount--
	}

	maxBlockSize := common.PovChainBlockSize
	curBlkMsgSize := 0

	endHeight := startHeight + uint64(blockCount)
	for height := startHeight; height < endHeight; height++ {
		block, _ := ss.ledger.GetPovBlockByHeight(height)
		if block == nil {
			ss.logger.Debugf("failed to get block by height %d", height)
			break
		}
		rsp.Blocks = append(rsp.Blocks, block)

		curBlkMsgSize += block.Msgsize()
		if curBlkMsgSize >= maxBlockSize {
			break
		}
	}

	rsp.Count = uint32(len(rsp.Blocks))

	ss.eb.Publish(common.EventSendMsgToSingle, p2p.PovBulkPullRsp, rsp, msg.msgPeer)
}

func (ss *PovSyncer) processPovBulkPullReqByBackward(msg *PovSyncMessage) {
	req := msg.msgValue.(*protos.PovBulkPullReq)

	if len(req.Locators) > 0 {
		ss.logger.Debugf("recv PovBulkPullReq by backward from peer %s, reason %d locator %s count %d",
			msg.msgPeer, req.Reason, req.Locators[0], req.Count)
	} else if !req.StartHash.IsZero() {
		ss.logger.Debugf("recv PovBulkPullReq by backward from peer %s, reason %d hash %s count %d",
			msg.msgPeer, req.Reason, req.StartHash, req.Count)
	} else {
		ss.logger.Debugf("recv PovBulkPullReq by backward from peer %s, reason %d height %d count %d",
			msg.msgPeer, req.Reason, req.StartHeight, req.Count)
	}

	rsp := new(protos.PovBulkPullRsp)
	rsp.Reason = req.Reason

	startHeight := req.StartHeight
	blockCount := req.Count
	if blockCount == 0 {
		blockCount = maxSyncBlockPerReq
	}
	if len(req.Locators) > 0 {
		block := ss.chain.LocateBestBlock(req.Locators)
		if block == nil {
			ss.logger.Debugf("failed to locate best block %s", req.Locators[0])
			return
		}
		rsp.Blocks = append(rsp.Blocks, block)

		if block.GetHeight() > 1 {
			startHeight = block.GetHeight() - 1
		} else {
			startHeight = 0
		}

		blockCount--
	} else if !req.StartHash.IsZero() {
		block, _ := ss.ledger.GetPovBlockByHash(req.StartHash)
		if block == nil {
			ss.logger.Debugf("failed to get block by hash %s", req.StartHash)
			return
		}
		rsp.Blocks = append(rsp.Blocks, block)

		if block.GetHeight() > 1 {
			startHeight = block.GetHeight() - 1
		} else {
			startHeight = 0
		}

		blockCount--
	}

	maxBlockSize := common.PovChainBlockSize
	curBlkMsgSize := 0

	endHeight := uint64(0)
	if startHeight > uint64(blockCount) {
		endHeight = startHeight - uint64(blockCount)
	}
	for height := startHeight; height > endHeight; height-- {
		block, err := ss.ledger.GetPovBlockByHeight(height)
		if err != nil {
			ss.logger.Debugf("failed to get block by height %d, err %s", height, err)
			break
		}
		rsp.Blocks = append(rsp.Blocks, block)

		curBlkMsgSize += block.Msgsize()
		if curBlkMsgSize >= maxBlockSize {
			break
		}
	}

	rsp.Count = uint32(len(rsp.Blocks))

	ss.eb.Publish(common.EventSendMsgToSingle, p2p.PovBulkPullRsp, rsp, msg.msgPeer)
}

func (ss *PovSyncer) processPovBulkPullReqByBatch(msg *PovSyncMessage) {
	req := msg.msgValue.(*protos.PovBulkPullReq)

	ss.logger.Debugf("recv PovBulkPullReq by batch from peer %s, locators %d", msg.msgPeer, len(req.Locators))

	rsp := new(protos.PovBulkPullRsp)
	rsp.Reason = req.Reason

	maxBlockSize := common.PovChainBlockSize
	curBlkMsgSize := 0

	for _, locHash := range req.Locators {
		if locHash == nil {
			continue
		}

		blockHash := *locHash
		block, _ := ss.ledger.GetPovBlockByHash(blockHash)
		if block == nil {
			ss.logger.Debugf("failed to get block by hash %s", blockHash)
			continue
		}

		rsp.Blocks = append(rsp.Blocks, block)

		curBlkMsgSize += block.Msgsize()
		if curBlkMsgSize >= maxBlockSize {
			break
		}
	}

	rsp.Count = uint32(len(rsp.Blocks))

	ss.eb.Publish(common.EventSendMsgToSingle, p2p.PovBulkPullRsp, rsp, msg.msgPeer)
}

func (ss *PovSyncer) processPovBulkPullRsp(msg *PovSyncMessage) {
	rsp := msg.msgValue.(*protos.PovBulkPullRsp)

	if rsp.Reason == protos.PovReasonSync {
		ss.logger.Infof("recv PovBulkPullRsp from peer %s, reason %d count %d",
			msg.msgPeer, rsp.Reason, rsp.Count)
	} else {
		ss.logger.Debugf("recv PovBulkPullRsp from peer %s, reason %d count %d",
			msg.msgPeer, rsp.Reason, rsp.Count)
	}

	if rsp.Count == 0 {
		return
	}

	if rsp.Reason == protos.PovReasonSync {
		if !ss.inSyncing.Load() {
			ss.logger.Infof("recv PovBulkPullRsp but state not in syncing")
			return
		}

		if ss.syncPeerID != msg.msgPeer {
			ss.logger.Infof("recv PovBulkPullRsp but peer %s is not sync peer", msg.msgPeer)
			return
		}

		syncPeer := ss.FindPeerWithStatus(msg.msgPeer, peerStatusGood)
		if syncPeer == nil {
			ss.logger.Infof("recv PovBulkPullRsp but peer %s is not exist", msg.msgPeer)
			return
		}
		if syncPeer.syncSeqID != ss.syncSeqID.Load() {
			ss.logger.Infof("recv PovBulkPullRsp but syncSeqID is not equal, %d, %d",
				syncPeer.syncSeqID, ss.syncSeqID.Load())
			return
		}

		syncPeer.waitSyncRspMsg = false

		for _, block := range rsp.Blocks {
			ss.addSyncBlock(block, syncPeer)
		}
	} else {
		for _, block := range rsp.Blocks {
			ss.eb.Publish(common.EventPovRecvBlock, block, types.PovBlockFromRemoteFetch, msg.msgPeer)
		}
	}
}

func (ss *PovSyncer) processEvent(evt *PovSyncEvent) {
	switch evt.eventType {
	case common.EventAddP2PStream:
		ss.processStreamEvent(evt)
	case common.EventDeleteP2PStream:
		break
	default:
		ss.logger.Infof("unknown event type %T!\n", evt.eventType)
	}
}

func (ss *PovSyncer) setInitState(st common.SyncState) {
	if ss.initSyncOver.Load() {
		return
	}

	if st == ss.initSyncState {
		return
	}

	ss.initSyncState = st
	if st == common.Syncing {
		ss.syncStartTime = time.Now()
	} else if st == common.SyncDone {
		ss.syncEndTime = time.Now()
		usedTime := ss.syncEndTime.Sub(ss.syncStartTime)
		ss.logger.Infof("pov init sync used time: %s", usedTime)
		ss.initSyncOver.Store(true)
	}
	ss.eb.Publish(common.EventPovSyncState, ss.initSyncState)
}

func (ss *PovSyncer) requestBlocksByHashes(reqBlkHashes []*types.Hash, peerID string) {
	if len(reqBlkHashes) == 0 {
		return
	}

	var peer *PovSyncPeer
	if peerID != "" {
		peer = ss.FindPeerWithStatus(peerID, peerStatusGood)
	}
	if peer == nil {
		peer = ss.GetRandomTopPeer(3)
	}
	if peer == nil {
		return
	}

	for len(reqBlkHashes) > 0 {
		sendHashNum := 0
		if len(reqBlkHashes) > maxPullBlockPerReq {
			sendHashNum = maxPullBlockPerReq
		} else {
			sendHashNum = len(reqBlkHashes)
		}

		sendBlkHashes := reqBlkHashes[0:sendHashNum]

		req := new(protos.PovBulkPullReq)
		req.PullType = protos.PovPullTypeBatch
		req.Locators = sendBlkHashes
		req.Count = uint32(len(sendBlkHashes))
		req.Reason = protos.PovReasonFetch

		ss.logger.Debugf("request blocks %d from peer %s", len(sendBlkHashes), peer.peerID)
		ss.eb.Publish(common.EventSendMsgToSingle, p2p.PovBulkPullReq, req, peer.peerID)

		reqBlkHashes = reqBlkHashes[sendHashNum:]
	}
}

func (ss *PovSyncer) requestTxsByHashes(reqTxHashes []*types.Hash, peerID string) {
	if len(reqTxHashes) == 0 {
		return
	}

	if time.Now().Unix() < (ss.lastReqTxTime.Load() + 15) {
		return
	}

	var peer *PovSyncPeer
	if peerID != "" {
		peer = ss.FindPeerWithStatus(peerID, peerStatusGood)
	}
	if peer == nil {
		peer = ss.GetRandomTopPeer(3)
	}
	if peer == nil {
		return
	}

	ss.logger.Infof("request txs %d from peer %s", len(reqTxHashes), peer.peerID)

	ss.eb.Publish(common.EventFrontiersReq, peer.peerID)
	ss.lastReqTxTime.Store(time.Now().Unix())
}

func (ss *PovSyncer) requestSyncFrontiers(peerID string) {
	if time.Now().Unix() < (ss.lastReqTxTime.Load() + 15) {
		return
	}

	var peer *PovSyncPeer
	if peerID != "" {
		peer = ss.FindPeerWithStatus(peerID, peerStatusGood)
	}
	if peer == nil {
		peer = ss.GetRandomTopPeer(3)
	}
	if peer == nil {
		return
	}

	ss.logger.Infof("request frontiers from peer %s", peer.peerID)

	ss.eb.Publish(common.EventFrontiersReq, peer.peerID)
	ss.lastReqTxTime.Store(time.Now().Unix())
}

func (ss *PovSyncer) absDiffHeight(lhs uint64, rhs uint64) uint64 {
	if lhs > rhs {
		return lhs - rhs
	}
	return rhs - lhs
}

func (ss *PovSyncer) GetDebugInfo() map[string]interface{} {
	// !!! be very careful about to map concurrent read !!!

	info := make(map[string]interface{})
	info["initSyncOver"] = ss.initSyncOver.Load()
	info["inSyncing"] = ss.inSyncing.Load()
	info["syncSeqID"] = ss.syncSeqID.Load()
	info["syncToHeight"] = ss.syncToHeight
	info["syncCurHeight"] = ss.syncCurHeight
	info["syncRcvHeight"] = ss.syncRcvHeight
	info["syncStartTime"] = ss.syncStartTime
	info["syncEndTime"] = ss.syncEndTime
	info["syncQueueNum"] = len(ss.syncBlocks)

	syncCurBlock := ss.findSyncCurBlockForDebug()
	if syncCurBlock != nil {
		blkInfo := make(map[string]interface{})
		info["syncCurBlock"] = blkInfo
		blkInfo["height"] = syncCurBlock.Height
		blkInfo["peerID"] = syncCurBlock.PeerID
		if syncCurBlock.Block != nil {
			blkInfo["hash"] = syncCurBlock.Block.GetHash()
			blkInfo["txNum"] = syncCurBlock.Block.GetTxNum()
			blkInfo["checkTxIndex"] = syncCurBlock.CheckTxIndex
			txPov := syncCurBlock.Block.GetTxByIndex(syncCurBlock.CheckTxIndex)
			if txPov != nil {
				blkInfo["checkTxHash"] = txPov.Hash
			}
		}
	}

	info["syncPeerID"] = ss.syncPeerID
	if ss.syncPeerID != "" {
		syncPeer := ss.FindPeer(ss.syncPeerID)
		if syncPeer != nil {
			peerInfo := make(map[string]interface{})
			info["syncPeerInfo"] = peerInfo
			peerInfo["status"] = syncPeer.status
			peerInfo["syncSeqID"] = syncPeer.syncSeqID
			peerInfo["waitSyncRspMsg"] = syncPeer.waitSyncRspMsg
			peerInfo["waitLocatorRsp"] = syncPeer.waitLocatorRsp
			peerInfo["lastStatusTime"] = syncPeer.lastStatusTime
			peerInfo["lastSyncReqTime"] = syncPeer.lastSyncReqTime
			peerInfo["currentHeight"] = syncPeer.currentHeight
			peerInfo["currentTD"] = syncPeer.currentTD
			peerInfo["currentHash"] = syncPeer.currentHash
		}
	}

	info["peerCount"] = ss.PeerCount()
	info["peerGoodCount"] = ss.PeerCountWithStatus(peerStatusGood)
	topPeers := ss.GetBestPeers(5)
	if len(topPeers) > 0 {
		topInfos := make([]map[string]interface{}, 0)
		for _, tp := range topPeers {
			topInfo := make(map[string]interface{})
			topInfo["peerID"] = tp.peerID
			topInfo["currentHeight"] = tp.currentHeight
			topInfo["currentTD"] = tp.currentTD
			topInfo["currentHash"] = tp.currentHash
			topInfo["status"] = tp.status
			topInfos = append(topInfos, topInfo)
		}
		info["topPeers"] = topInfos
	}

	return info
}
