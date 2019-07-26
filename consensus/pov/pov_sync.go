package pov

import (
	"github.com/qlcchain/go-qlc/ledger"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"go.uber.org/zap"
)

const (
	minForcePovSyncPeerCount  = 1
	minEnoughPovSyncPeerCount = 3
	checkPeerStatusTime       = 30
	waitEnoughPeerTime        = 60

	maxSyncBlockPerReq = 100
	maxPullBlockPerReq = 100
	maxPullTxPerReq    = 100

	peerStatusInit = 0
	peerStatusGood = 1
	peerStatusDead = 2
	peerStatusSlow = 3
)

type PovSyncPeer struct {
	peerID         string
	currentHeight  uint64
	currentTD      *big.Int
	timestamp      int64
	lastStatusTime time.Time
	status         int

	lastSyncRspTime time.Time
}

// PeerSetByTD is in descend order
type PovSyncPeerSetByTD []*PovSyncPeer

func (s PovSyncPeerSetByTD) Len() int           { return len(s) }
func (s PovSyncPeerSetByTD) Less(i, j int) bool { return s[i].currentTD.Cmp(s[j].currentTD) > 0 }
func (s PovSyncPeerSetByTD) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type PovSyncPeerSetByHeight []*PovSyncPeer

func (s PovSyncPeerSetByHeight) Len() int           { return len(s) }
func (s PovSyncPeerSetByHeight) Less(i, j int) bool { return s[i].currentHeight < s[j].currentHeight }
func (s PovSyncPeerSetByHeight) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type PovSyncer struct {
	povEngine *PoVEngine
	logger    *zap.SugaredLogger
	allPeers  sync.Map // map[string]*PovSyncPeer

	state         common.SyncState
	syncStartTime time.Time
	syncEndTime   time.Time
	lastCheckTime time.Time

	syncCurHeight uint64
	syncToHeight  uint64
	syncPeerID    string

	syncPeerLostTime time.Time

	messageCh chan *PovSyncMessage
	eventCh   chan *PovSyncEvent
	quitCh    chan struct{}
}

type PovSyncMessage struct {
	msgValue interface{}
	msgHash  types.Hash
	msgPeer  string
}

type PovSyncEvent struct {
	eventType common.TopicType
	eventData interface{}
}

func NewPovSyncer(povEngine *PoVEngine) *PovSyncer {
	ss := &PovSyncer{
		povEngine:     povEngine,
		state:         common.SyncNotStart,
		lastCheckTime: time.Now(),
		messageCh:     make(chan *PovSyncMessage, 2000),
		eventCh:       make(chan *PovSyncEvent, 200),
		quitCh:        make(chan struct{}),
		logger:        log.NewLogger("pov_sync"),
	}
	return ss
}

func (ss *PovSyncer) Start() {
	eb := ss.povEngine.GetEventBus()
	if eb != nil {
		err := eb.SubscribeSync(common.EventAddP2PStream, ss.onAddP2PStream)
		if err != nil {
			return
		}
		err = eb.SubscribeSync(common.EventDeleteP2PStream, ss.onDeleteP2PStream)
		if err != nil {
			return
		}
		err = eb.SubscribeSync(common.EventPovPeerStatus, ss.onPovStatus)
		if err != nil {
			return
		}
		err = eb.Subscribe(common.EventPovBulkPullReq, ss.onPovBulkPullReq)
		if err != nil {
			return
		}
		err = eb.Subscribe(common.EventPovBulkPullRsp, ss.onPovBulkPullRsp)
		if err != nil {
			return
		}
	}

	common.Go(ss.mainLoop)
	common.Go(ss.syncLoop)
}

func (ss *PovSyncer) Stop() {
	eb := ss.povEngine.GetEventBus()
	if eb != nil {
		err := eb.Unsubscribe(common.EventAddP2PStream, ss.onAddP2PStream)
		if err != nil {
			return
		}
		err = eb.Unsubscribe(common.EventDeleteP2PStream, ss.onDeleteP2PStream)
		if err != nil {
			return
		}
		err = eb.Unsubscribe(common.EventPovPeerStatus, ss.onPovStatus)
		if err != nil {
			return
		}
		err = eb.Unsubscribe(common.EventPovBulkPullReq, ss.onPovBulkPullReq)
		if err != nil {
			return
		}
		err = eb.Unsubscribe(common.EventPovBulkPullRsp, ss.onPovBulkPullRsp)
		if err != nil {
			return
		}
	}

	close(ss.quitCh)
}

func (ss *PovSyncer) getChain() *PovBlockChain {
	return ss.povEngine.GetChain()
}

func (ss *PovSyncer) getLedger() ledger.Store {
	return ss.povEngine.GetLedger()
}

func (ss *PovSyncer) getEventBus() event.EventBus {
	return ss.povEngine.eb
}

func (ss *PovSyncer) getState() common.SyncState {
	return ss.state
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

		case event := <-ss.eventCh:
			ss.processEvent(event)
		}
	}
}

func (ss *PovSyncer) syncLoop() {
	waitTimer := time.NewTimer(waitEnoughPeerTime * time.Second)

wait:
	for {
		select {
		case <-ss.quitCh:
			return
		case <-waitTimer.C:
			peerCnt := ss.PeerCountWithStatus(peerStatusGood)
			if peerCnt >= minForcePovSyncPeerCount {
				ss.logger.Infof("prepare sync after timeout with peers %d", peerCnt)
				break wait
			} else {
				ss.logger.Warnf("can not sync after timeout with peers %d", peerCnt)
				waitTimer.Reset(waitEnoughPeerTime * time.Second)
			}
		default:
			peerCnt := ss.PeerCountWithStatus(peerStatusGood)
			if peerCnt >= minEnoughPovSyncPeerCount {
				ss.logger.Infof("prepare sync after got enough peers %d", peerCnt)
				break wait
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}

	waitTimer.Stop()

	checkSyncTicker := time.NewTicker(30 * time.Second)
	checkChainTicker := time.NewTicker(10 * time.Second)
	ss.lastCheckTime = time.Now()

loop:
	for {
		select {
		case <-ss.quitCh:
			return

		case <-checkSyncTicker.C:
			ss.checkSyncPeer()
			if ss.isFinished() {
				break loop
			}

		case <-checkChainTicker.C:
			ss.checkChain()
			if ss.isFinished() {
				break loop
			}
		}
	}

	ss.logger.Infof("exit pov sync loop")

	checkSyncTicker.Stop()
	checkChainTicker.Stop()
}

func (ss *PovSyncer) onAddP2PStream(peerID string) {
	ss.logger.Infof("add peer %s", peerID)

	peer := &PovSyncPeer{
		peerID:         peerID,
		currentHeight:  0,
		currentTD:      big.NewInt(0),
		lastStatusTime: time.Now(),
		status:         peerStatusInit,
	}

	ss.allPeers.Store(peerID, peer)

	ss.eventCh <- &PovSyncEvent{eventType: common.EventAddP2PStream, eventData: peerID}
}

func (ss *PovSyncer) onDeleteP2PStream(peerID string) {
	ss.logger.Infof("delete peer %s", peerID)
	ss.allPeers.Delete(peerID)

	ss.eventCh <- &PovSyncEvent{eventType: common.EventDeleteP2PStream, eventData: peerID}
}

func (ss *PovSyncer) onPovStatus(status *protos.PovStatus, msgHash types.Hash, msgPeer string) {
	if v, ok := ss.allPeers.Load(msgPeer); ok {
		peer := v.(*PovSyncPeer)

		td := new(big.Int).SetBytes(status.CurrentTD)
		ss.logger.Infof("recv PovStatus from peer %s, head %d/%s, td %d/%s",
			msgPeer, status.CurrentHeight, status.CurrentHash, td.BitLen(), td.Text(16))
		if status.GenesisHash != ss.getChain().GenesisBlock().GetHash() {
			ss.logger.Warnf("peer %s genesis hash %s is invalid", msgPeer, status.GenesisHash)
			return
		}

		peer.currentHeight = status.CurrentHeight
		peer.currentTD = td
		peer.timestamp = status.Timestamp
		peer.lastStatusTime = time.Now()
		if peer.status != peerStatusSlow {
			peer.status = peerStatusGood
		}
	}
}

func (ss *PovSyncer) onPovBulkPullReq(req *protos.PovBulkPullReq, msgHash types.Hash, msgPeer string) {
	ss.messageCh <- &PovSyncMessage{msgValue: req, msgHash: msgHash, msgPeer: msgPeer}
}

func (ss *PovSyncer) onPovBulkPullRsp(rsp *protos.PovBulkPullRsp, msgHash types.Hash, msgPeer string) {
	ss.messageCh <- &PovSyncMessage{msgValue: rsp, msgHash: msgHash, msgPeer: msgPeer}
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
			ss.logger.Infof("recv PovBulkPullReq by forward from peer %s, reason %d locator %s count %d", msg.msgPeer, req.Reason, req.Locators[0], req.Count)
		} else if !req.StartHash.IsZero() {
			ss.logger.Infof("recv PovBulkPullReq by forward from peer %s, reason %d hash %s count %d", msg.msgPeer, req.Reason, req.StartHash, req.Count)
		} else {
			ss.logger.Infof("recv PovBulkPullReq by forward from peer %s, reason %d height %d count %d", msg.msgPeer, req.Reason, req.StartHeight, req.Count)
		}
	} else {
		if len(req.Locators) > 0 {
			ss.logger.Debugf("recv PovBulkPullReq by forward from peer %s, reason %d locator %s count %d", msg.msgPeer, req.Reason, req.Locators[0], req.Count)
		} else if !req.StartHash.IsZero() {
			ss.logger.Debugf("recv PovBulkPullReq by forward from peer %s, reason %d hash %s count %d", msg.msgPeer, req.Reason, req.StartHash, req.Count)
		} else {
			ss.logger.Debugf("recv PovBulkPullReq by forward from peer %s, reason %d height %d count %d", msg.msgPeer, req.Reason, req.StartHeight, req.Count)
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
		block := ss.getChain().LocateBestBlock(req.Locators)
		if block == nil {
			ss.logger.Debugf("failed to locate best block %s", req.Locators[0])
			return
		}
		rsp.Blocks = append(rsp.Blocks, block)
		startHeight = block.GetHeight() + 1
		blockCount = blockCount - 1
	} else if !req.StartHash.IsZero() {
		block, _ := ss.getLedger().GetPovBlockByHash(req.StartHash)
		if block == nil {
			ss.logger.Debugf("failed to get block by hash %s", req.StartHash)
			return
		}
		rsp.Blocks = append(rsp.Blocks, block)
		startHeight = block.GetHeight() + 1
		blockCount = blockCount - 1
	}

	maxBlockSize := common.PovChainBlockSize
	curBlkMsgSize := 0

	endHeight := startHeight + uint64(blockCount)
	for height := startHeight; height < endHeight; height++ {
		block, _ := ss.getLedger().GetPovBlockByHeight(height)
		if block == nil {
			ss.logger.Debugf("failed to get block by height %d", height)
			break
		}
		rsp.Blocks = append(rsp.Blocks, block)

		curBlkMsgSize = curBlkMsgSize + block.Msgsize()
		if curBlkMsgSize >= maxBlockSize {
			break
		}
	}

	rsp.Count = uint32(len(rsp.Blocks))

	ss.getEventBus().Publish(common.EventSendMsgToSingle, p2p.PovBulkPullRsp, rsp, msg.msgPeer)
}

func (ss *PovSyncer) processPovBulkPullReqByBackward(msg *PovSyncMessage) {
	req := msg.msgValue.(*protos.PovBulkPullReq)

	if len(req.Locators) > 0 {
		ss.logger.Debugf("recv PovBulkPullReq by backward from peer %s, reason %d locator %s count %d", msg.msgPeer, req.Reason, req.Locators[0], req.Count)
	} else if !req.StartHash.IsZero() {
		ss.logger.Debugf("recv PovBulkPullReq by backward from peer %s, reason %d hash %s count %d", msg.msgPeer, req.Reason, req.StartHash, req.Count)
	} else {
		ss.logger.Debugf("recv PovBulkPullReq by backward from peer %s, reason %d height %d count %d", msg.msgPeer, req.Reason, req.StartHeight, req.Count)
	}

	rsp := new(protos.PovBulkPullRsp)
	rsp.Reason = req.Reason

	startHeight := req.StartHeight
	blockCount := req.Count
	if blockCount == 0 {
		blockCount = maxSyncBlockPerReq
	}
	if len(req.Locators) > 0 {
		block := ss.getChain().LocateBestBlock(req.Locators)
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

		blockCount = blockCount - 1
	} else if !req.StartHash.IsZero() {
		block, _ := ss.getLedger().GetPovBlockByHash(req.StartHash)
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

		blockCount = blockCount - 1
	}

	maxBlockSize := common.PovChainBlockSize
	curBlkMsgSize := 0

	endHeight := uint64(0)
	if startHeight > uint64(blockCount) {
		endHeight = startHeight - uint64(blockCount)
	}
	for height := startHeight; height > endHeight; height-- {
		block, err := ss.getLedger().GetPovBlockByHeight(height)
		if err != nil {
			ss.logger.Debugf("failed to get block by height %d, err %s", height, err)
			break
		}
		rsp.Blocks = append(rsp.Blocks, block)

		curBlkMsgSize = curBlkMsgSize + block.Msgsize()
		if curBlkMsgSize >= maxBlockSize {
			break
		}
	}

	rsp.Count = uint32(len(rsp.Blocks))

	ss.getEventBus().Publish(common.EventSendMsgToSingle, p2p.PovBulkPullRsp, rsp, msg.msgPeer)
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
		block, _ := ss.getLedger().GetPovBlockByHash(blockHash)
		if block == nil {
			ss.logger.Debugf("failed to get block by hash %s", blockHash)
			continue
		}

		rsp.Blocks = append(rsp.Blocks, block)

		curBlkMsgSize = curBlkMsgSize + block.Msgsize()
		if curBlkMsgSize >= maxBlockSize {
			break
		}
	}

	rsp.Count = uint32(len(rsp.Blocks))

	ss.getEventBus().Publish(common.EventSendMsgToSingle, p2p.PovBulkPullRsp, rsp, msg.msgPeer)
}

func (ss *PovSyncer) processPovBulkPullRsp(msg *PovSyncMessage) {
	rsp := msg.msgValue.(*protos.PovBulkPullRsp)

	if rsp.Reason == protos.PovReasonSync {
		ss.logger.Infof("recv PovBulkPullRsp from peer %s, reason %d count %d", msg.msgPeer, rsp.Reason, rsp.Count)
	} else {
		ss.logger.Debugf("recv PovBulkPullRsp from peer %s, reason %d count %d", msg.msgPeer, rsp.Reason, rsp.Count)
	}

	if rsp.Count == 0 {
		return
	}

	if rsp.Reason == protos.PovReasonSync {
		if ss.getState() != common.Syncing {
			ss.logger.Infof("recv PovBulkPullRsp but state not in syncing")
			return
		}

		if ss.syncPeerID != msg.msgPeer {
			ss.logger.Infof("recv PovBulkPullRsp from peer %s is not sync peer", msg.msgPeer)
			return
		}

		syncPeer := ss.FindPeerWithStatus(msg.msgPeer, peerStatusGood)
		if syncPeer == nil {
			ss.logger.Infof("recv PovBulkPullRsp from peer %s is not exist", msg.msgPeer)
			return
		}

		syncPeer.lastSyncRspTime = time.Now()
	}

	fromType := types.PovBlockFromRemoteFetch
	if rsp.Reason == protos.PovReasonSync {
		fromType = types.PovBlockFromRemoteFetch
	}

	lastBlockHeight := uint64(0)
	for _, block := range rsp.Blocks {
		_ = ss.povEngine.AddBlock(block, fromType, msg.msgPeer)

		lastBlockHeight = block.GetHeight()
	}

	if rsp.Reason == protos.PovReasonSync {
		ss.requestSyncingBlocks(false, lastBlockHeight)
	}
}

func (ss *PovSyncer) processEvent(event *PovSyncEvent) {
	switch event.eventType {
	case common.EventAddP2PStream:
		ss.processStreamEvent(event)
	case common.EventDeleteP2PStream:
		break
	default:
		ss.logger.Infof("unknown event type %T!\n", event.eventType)
	}
}

func (ss *PovSyncer) processStreamEvent(event *PovSyncEvent) {
	peerID := event.eventData.(string)

	genesisBlock := ss.povEngine.chain.GenesisBlock()
	latestBlock := ss.povEngine.chain.LatestBlock()
	latestTD := ss.povEngine.chain.GetBlockTDByHash(latestBlock.GetHash())

	status := &protos.PovStatus{
		CurrentHeight: latestBlock.GetHeight(),
		CurrentTD:     latestTD.Bytes(),
		CurrentHash:   latestBlock.GetHash(),
		GenesisHash:   genesisBlock.GetHash(),
		Timestamp:     time.Now().Unix(),
	}
	ss.logger.Debugf("send PovStatus to peer %s", peerID)
	ss.povEngine.eb.Publish(common.EventSendMsgToSingle, p2p.PovStatus, status, peerID)
}

func (ss *PovSyncer) checkAllPeers() {
	peerCount := ss.PeerCount()
	if peerCount <= 0 {
		return
	}

	genesisBlock := ss.povEngine.chain.GenesisBlock()
	latestBlock := ss.povEngine.chain.LatestBlock()
	latestTD := ss.povEngine.chain.GetBlockTDByHash(latestBlock.GetHash())

	status := &protos.PovStatus{
		CurrentHeight: latestBlock.GetHeight(),
		CurrentTD:     latestTD.Bytes(),
		CurrentHash:   latestBlock.GetHash(),
		GenesisHash:   genesisBlock.GetHash(),
		Timestamp:     time.Now().Unix(),
	}
	ss.logger.Infof("broadcast PovStatus to %d peers", peerCount)
	ss.povEngine.eb.Publish(common.EventBroadcast, p2p.PovStatus, status)

	now := time.Now()
	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if now.Sub(peer.lastStatusTime) >= 10*time.Minute {
			if peer.status != peerStatusDead {
				ss.logger.Infof("peer %s may be dead", peer.peerID)
				peer.status = peerStatusDead
			}
		}
		return true
	})
}

func (ss *PovSyncer) checkSyncPeer() {
	if ss.state == common.SyncNotStart {
		ss.setState(common.Syncing)
	} else if ss.state != common.Syncing {
		return
	}

	topPeers := ss.GetBestPeers(10)
	ss.logger.Infof("topPeers: %d", len(topPeers))
	for _, peer := range topPeers {
		ss.logger.Infof("%s-%d-%s", peer.peerID, peer.currentHeight, peer.currentTD)
	}

	bestPeer := ss.GetBestPeer(ss.syncPeerID)
	if bestPeer == nil {
		if ss.syncPeerLostTime.Unix() > 0 {
			if time.Now().Unix() >= ss.syncPeerLostTime.Add(10*time.Minute).Unix() {
				ss.logger.Errorf("sync err, because no peers in 10 minutes")
				ss.setState(common.Syncerr)
			}
			return
		} else {
			ss.logger.Warnf("there is no best peer for sync, last peer %s", ss.syncPeerID)
			ss.syncPeerLostTime = time.Now()
			return
		}
	} else {
		ss.syncPeerLostTime = time.Time{}

		lastSyncPeer := ss.FindPeerWithStatus(ss.syncPeerID, peerStatusGood)
		if lastSyncPeer != nil && lastSyncPeer.peerID != bestPeer.peerID {
			if bestPeer.currentHeight < (lastSyncPeer.currentHeight + 20) {
				ss.logger.Infof("no need switch sync peer, best %s height %d, last %s height %d",
					bestPeer.peerID, bestPeer.currentHeight, lastSyncPeer.peerID, lastSyncPeer.currentHeight)
				return
			}
		}
	}

	ss.syncWithPeer(bestPeer)
}

func (ss *PovSyncer) checkChain() {
	timeNow := time.Now()

	if ss.state != common.Syncing {
		ss.lastCheckTime = timeNow
		return
	}
	syncPeer := ss.FindPeer(ss.syncPeerID)
	if syncPeer == nil {
		ss.logger.Warnf("sync peer %s is gone", ss.syncPeerID)
		return
	}

	latestBlock := ss.getChain().LatestBlock()
	if latestBlock == nil {
		ss.logger.Errorf("failed to get latest block")
		return
	}
	latestTD := ss.getChain().GetBlockTDByHash(latestBlock.GetHash())
	if latestTD == nil {
		ss.logger.Errorf("failed to latest block td")
		return
	}

	if latestTD.Cmp(syncPeer.currentTD) >= 0 {
		ss.logger.Infof("sync done, current height: %d", latestBlock.Height)
		ss.setState(common.Syncdone)
		return
	}

	ss.logger.Infof("syncCurHeight: %d, syncToHeight: %d, chainHeight: %d",
		ss.syncCurHeight, ss.syncToHeight, latestBlock.Height)

	ss.lastCheckTime = timeNow
}

func (ss *PovSyncer) setState(st common.SyncState) {
	ss.state = st
	if st == common.Syncing {
		ss.syncStartTime = time.Now()
	} else if st == common.Syncdone || st == common.Syncerr {
		ss.syncEndTime = time.Now()
		usedTime := ss.syncEndTime.Sub(ss.syncStartTime)
		ss.logger.Infof("pov sync used time: %s", usedTime)
	}
	ss.povEngine.GetEventBus().Publish(common.EventPovSyncState, ss.state)
}

func (ss *PovSyncer) isFinished() bool {
	if ss.state == common.SyncNotStart || ss.state == common.Syncing {
		return false
	}

	return true
}

func (ss *PovSyncer) FindPeer(peerID string) *PovSyncPeer {
	if peerID == "" {
		return nil
	}

	if v, ok := ss.allPeers.Load(peerID); ok {
		peer := v.(*PovSyncPeer)
		return peer
	}

	return nil
}

func (ss *PovSyncer) FindPeerWithStatus(peerID string, status int) *PovSyncPeer {
	peer := ss.FindPeer(peerID)
	if peer != nil {
		if peer.status == status {
			return peer
		}
	}

	return nil
}

func (ss *PovSyncer) GetBestPeer(lastPeerID string) *PovSyncPeer {
	bestPeer := ss.FindPeerWithStatus(lastPeerID, peerStatusGood)

	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if peer.status != peerStatusGood {
			return true
		}
		if bestPeer == nil {
			bestPeer = peer
		} else if peer.currentTD.Cmp(bestPeer.currentTD) > 0 {
			bestPeer = peer
		}
		return true
	})

	return bestPeer
}

func (ss *PovSyncer) GetBestPeers(limit int) []*PovSyncPeer {
	var allPeers PovSyncPeerSetByTD

	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if peer.status != peerStatusGood {
			return true
		}
		allPeers = append(allPeers, peer)
		return true
	})
	sort.Sort(allPeers)

	if len(allPeers) <= limit {
		return allPeers
	}

	return allPeers[:limit]
}

// GetRandomTopPeer select one peer from top peers
func (ss *PovSyncer) GetRandomTopPeer(top int) *PovSyncPeer {
	peers := ss.GetBestPeers(top)
	if len(peers) <= 0 {
		return nil
	}
	if len(peers) == 1 {
		return peers[0]
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := rd.Intn(len(peers))
	if idx >= len(peers) {
		idx = idx - 1
	}
	if idx < 0 {
		idx = 0
	}
	return peers[idx]
}

func (ss *PovSyncer) GetRandomPeers(limit int) []*PovSyncPeer {
	var allPeers []*PovSyncPeer
	var selectPeers []*PovSyncPeer

	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if peer.status != peerStatusGood {
			return true
		}
		allPeers = append(allPeers, peer)
		return true
	})

	if len(allPeers) <= limit {
		return allPeers
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	idxSeqs := rd.Perm(len(allPeers))

	for i := 0; i < limit; i++ {
		selectPeers = append(selectPeers, allPeers[idxSeqs[i]])
	}

	return selectPeers
}

func (ss *PovSyncer) GetPeerLocators() []*PovSyncPeer {
	var allPeers PovSyncPeerSetByTD

	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if peer.status != peerStatusGood {
			return true
		}
		allPeers = append(allPeers, peer)
		return true
	})
	sort.Sort(allPeers)

	if len(allPeers) <= 3 {
		return allPeers
	}

	var selectPeers []*PovSyncPeer
	selectPeers = append(selectPeers, allPeers[0])
	selectPeers = append(selectPeers, allPeers[len(allPeers)/2])
	selectPeers = append(selectPeers, allPeers[len(allPeers)-1])
	return selectPeers
}

func (ss *PovSyncer) PeerCount() int {
	peerCount := 0
	ss.allPeers.Range(func(key, value interface{}) bool {
		peerCount++
		return true
	})

	return peerCount
}

func (ss *PovSyncer) PeerCountWithStatus(status int) int {
	peerCount := 0
	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if peer.status == status {
			peerCount++
		}
		return true
	})

	return peerCount
}

func (ss *PovSyncer) syncWithPeer(peer *PovSyncPeer) {
	ss.syncToHeight = peer.currentHeight

	lastHeight := uint64(0)

	if ss.syncPeerID == peer.peerID {
		// check sync pull blocks action may be finished
		if ss.syncCurHeight >= ss.syncToHeight {
			return
		}

		// check sync pull blocks message may be lost
		if time.Now().Unix() < peer.lastSyncRspTime.Add(60*time.Second).Unix() {
			return
		}

		// check sync pull blocks too slow
		if time.Now().Unix() >= peer.lastSyncRspTime.Add(10*time.Minute).Unix() {
			ss.logger.Infof("sync peer %s may be too slow", peer.peerID)
			peer.status = peerStatusSlow
			return
		}

		lastHeight = ss.syncCurHeight - 1
	} else {
		peer.lastSyncRspTime = time.Now()
	}

	ss.logger.Infof("sync with peer %s to height %d", peer.peerID, peer.currentHeight)

	ss.syncPeerID = peer.peerID

	ss.requestSyncingBlocks(true, lastHeight)
}

func (ss *PovSyncer) requestSyncingBlocks(useLocator bool, lastHeight uint64) {
	if ss.state != common.Syncing {
		return
	}

	syncPeer := ss.FindPeerWithStatus(ss.syncPeerID, peerStatusGood)
	if syncPeer == nil {
		ss.logger.Warnf("request syncing blocks but peer %s is gone", ss.syncPeerID)
		return
	}

	ss.syncCurHeight = lastHeight + 1
	if ss.syncCurHeight >= ss.syncToHeight {
		return
	}

	req := new(protos.PovBulkPullReq)

	req.Count = maxSyncBlockPerReq
	req.StartHeight = lastHeight + 1
	if useLocator {
		req.Locators = ss.getChain().GetBlockLocator(types.ZeroHash)
	}
	req.Reason = protos.PovReasonSync

	if useLocator {
		ss.logger.Infof("request syncing blocks use locators %d with peer %s", len(req.Locators), ss.syncPeerID)
	} else {
		ss.logger.Infof("request syncing blocks use height %d with peer %s", req.StartHeight, ss.syncPeerID)
	}

	ss.povEngine.eb.Publish(common.EventSendMsgToSingle, p2p.PovBulkPullReq, req, ss.syncPeerID)
}

func (ss *PovSyncer) requestBlocksByHeight(startHeight uint64, count uint32) {
	peer := ss.GetBestPeer("")
	if peer == nil {
		return
	}

	req := new(protos.PovBulkPullReq)

	req.Count = count
	req.StartHeight = startHeight
	req.Reason = protos.PovReasonFetch

	ss.povEngine.eb.Publish(common.EventSendMsgToSingle, p2p.PovBulkPullReq, req, peer.peerID)
}

func (ss *PovSyncer) requestBlocksByHashes(reqBlkHashes []*types.Hash, peerID string) {
	if len(reqBlkHashes) <= 0 {
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
		ss.povEngine.eb.Publish(common.EventSendMsgToSingle, p2p.PovBulkPullReq, req, peer.peerID)

		reqBlkHashes = reqBlkHashes[sendHashNum:]
	}
}

func (ss *PovSyncer) requestTxsByHashes(reqTxHashes []*types.Hash, peerID string) {
	if len(reqTxHashes) <= 0 {
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

	for len(reqTxHashes) > 0 {
		sendHashNum := 0
		if len(reqTxHashes) > maxPullTxPerReq {
			sendHashNum = maxPullTxPerReq
		} else {
			sendHashNum = len(reqTxHashes)
		}

		sendTxHashes := reqTxHashes[0:sendHashNum]

		req := new(protos.BulkPullReqPacket)
		req.PullType = protos.PullTypeBatch
		req.Hashes = sendTxHashes
		req.Count = uint32(len(sendTxHashes))

		ss.logger.Debugf("request txs %d from peer %s", len(sendTxHashes), peer.peerID)
		ss.povEngine.eb.Publish(common.EventSendMsgToSingle, p2p.BulkPullRequest, req, peer.peerID)

		reqTxHashes = reqTxHashes[sendHashNum:]
	}
}
