package p2p

import (
	"errors"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"

	"github.com/qlcchain/go-qlc/common/topic"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	maxResendTime  = 3
	pullRspTimeOut = 10 * time.Minute
	pullReqTimeOut = 60 * time.Second
)

// Service manage sync tasks
type ServiceSync struct {
	netService *QlcService
	qlcLedger  *ledger.Ledger
	frontiers  []*types.Frontier
	quitCh     chan bool
	logger     *zap.SugaredLogger
	//	syncState          atomic.Value
	syncTicker         *time.Ticker
	pullTimer          *time.Timer
	pullRequestStartCh chan bool
	pullStartHash      types.Hash
	pullEndHash        types.Hash
	headerBlockHash    types.Hash
	openBlockHash      types.Hash
	bulkPull           []*protos.Bulk
	lastSyncHash       types.Hash
	quitChanForSync    chan bool
	mu                 *sync.Mutex
}

// NewService return new Service.
func NewSyncService(netService *QlcService, ledger *ledger.Ledger) *ServiceSync {
	ss := &ServiceSync{
		netService:         netService,
		qlcLedger:          ledger,
		quitCh:             make(chan bool, 1),
		logger:             log.NewLogger("sync"),
		pullTimer:          time.NewTimer(pullReqTimeOut),
		pullRequestStartCh: make(chan bool, 1),
		quitChanForSync:    make(chan bool, 1),
		mu:                 &sync.Mutex{},
	}
	return ss
}

func (ss *ServiceSync) Start() {
	ss.logger.Info("started sync loop")
	address := types.Address{}
	Req := protos.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	ss.syncTicker = time.NewTicker(time.Duration(ss.netService.node.cfg.P2P.SyncInterval) * time.Second)

	for {
		select {
		case <-ss.quitCh:
			ss.logger.Info("Stopped Sync Loop.")
			return
		case <-ss.syncTicker.C:
			syncState := ss.netService.cc.P2PSyncState()
			if syncState == topic.SyncFinish || syncState == topic.SyncNotStart {
				peerID, err := ss.netService.node.StreamManager().randomLowerLatencyPeer()
				if err != nil {
					ss.logger.Error(err)
					continue
				}
				ss.logger.Warnf("begin sync block from [%s]", peerID)
				err = ss.netService.node.SendMessageToPeer(FrontierRequest, Req, peerID)
				if err != nil {
					ss.logger.Errorf("err [%s] when send FrontierRequest", err)
				}
			}
		}
	}
}

// Stop sync service
func (ss *ServiceSync) Stop() {
	//ss.logger.VInfo("Stop Qlc sync...")

	ss.quitCh <- true
}

func (ss *ServiceSync) onConsensusSyncFinished() {
	select {
	case <-ss.quitChanForSync:
	default:
	}
	select {
	case ss.quitChanForSync <- true:
	default:
	}
}
func (ss *ServiceSync) syncInit() {
	select {
	case <-ss.quitChanForSync:
	default:
	}

	ss.lastSyncHash = types.ZeroHash
	ss.netService.msgEvent.Publish(topic.EventSyncStateChange, &topic.EventP2PSyncStateMsg{P2pSyncState: topic.Syncing})
}

func (ss *ServiceSync) onFrontierReq(message *Message) error {
	ss.netService.node.logger.Debug("receive FrontierReq")
	var fs []*types.Frontier
	fs, err := ss.qlcLedger.GetFrontiers()
	if err != nil {
		return err
	}
	frs := make([]*types.FrontierBlock, 0)
	for _, f := range fs {
		b, err := ss.qlcLedger.GetStateBlockConfirmed(f.HeaderBlock)
		if err != nil {
			ss.logger.Error(err)
			continue
		}
		fb := &types.FrontierBlock{
			Fr:        f,
			HeaderBlk: b,
		}
		frs = append(frs, fb)
	}
	rsp := &protos.FrontierResponse{
		Fs: frs,
	}
	err = ss.netService.SendMessageToPeer(FrontierRsp, rsp, message.MessageFrom())
	if err != nil {
		ss.logger.Errorf("send FrontierRsp err [%s]", err)
	}
	return nil
}

func (ss *ServiceSync) checkFrontier(message *Message) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	var err error
	syncState := ss.netService.cc.P2PSyncState()
	if syncState == topic.SyncFinish || syncState == topic.SyncNotStart {
		ss.frontiers, err = ss.getLocalFrontier(ss.qlcLedger)
		if err != nil {
			ss.logger.Error(err)
			return
		}
		ss.next()
		ss.bulkPull = ss.bulkPull[:0:0]
		rsp, err := protos.FrontierResponseFromProto(message.Data())
		if err != nil {
			ss.logger.Error(err)
			return
		}

		sv, err := ss.netService.cc.Service(context.ConsensusService)
		if err != nil {
			ss.logger.Error(err)
			return
		}

		ss.syncInit()
		sv.RpcCall(common.RpcDPoSOnSyncStateChange, topic.Syncing, nil)
		ss.logger.Warn("sync start")

		var remoteFrontiers []*types.Frontier
		var blks types.StateBlockList

		for _, f := range rsp.Fs {
			remoteFrontiers = append(remoteFrontiers, f.Fr)
			blks = append(blks, f.HeaderBlk)
		}
		remoteFrontiersLen := len(remoteFrontiers)

		if remoteFrontiersLen > 0 {
			sv.RpcCall(common.RpcDPoSProcessFrontier, blks, nil)
			sort.Sort(types.Frontiers(remoteFrontiers))
			zeroFrontier := new(types.Frontier)
			remoteFrontiers = append(remoteFrontiers, zeroFrontier)
			state := ss.processFrontiers(remoteFrontiers, message.MessageFrom())
			ss.netService.msgEvent.Publish(topic.EventSyncStateChange, &topic.EventP2PSyncStateMsg{P2pSyncState: state})
			sv.RpcCall(common.RpcDPoSOnSyncStateChange, state, nil)
		}
		ss.logger.Warn("sync pull all blocks done")
	}
}

func (ss *ServiceSync) processFrontiers(fsRemotes []*types.Frontier, peerID string) topic.SyncState {
	for i := 0; i < len(fsRemotes); i++ {
		if !fsRemotes[i].OpenBlock.IsZero() {
			for {
				if !ss.openBlockHash.IsZero() && (ss.openBlockHash.String() < fsRemotes[i].OpenBlock.String()) {
					// We have an account but remote peer have not.
					ss.next()
				} else {
					break
				}
			}
			if !ss.openBlockHash.IsZero() {
				if fsRemotes[i].OpenBlock == ss.openBlockHash {
					if ss.headerBlockHash == fsRemotes[i].HeaderBlock {
						//ss.logger.Infof("this token %s have the same block", openBlockHash)
					} else {
						exit, _ := ss.qlcLedger.HasStateBlockConfirmed(fsRemotes[i].HeaderBlock)
						if !exit {
							pull := &protos.Bulk{
								StartHash: ss.headerBlockHash,
								EndHash:   fsRemotes[i].HeaderBlock,
							}
							ss.bulkPull = append(ss.bulkPull, pull)
						}
					}
					ss.next()
				} else {
					if fsRemotes[i].OpenBlock.String() > ss.openBlockHash.String() {
						return topic.SyncDone
					}
					pull := &protos.Bulk{
						StartHash: types.ZeroHash,
						EndHash:   fsRemotes[i].HeaderBlock,
					}
					ss.bulkPull = append(ss.bulkPull, pull)
				}
			} else {
				pull := &protos.Bulk{
					StartHash: types.ZeroHash,
					EndHash:   fsRemotes[i].HeaderBlock,
				}
				ss.bulkPull = append(ss.bulkPull, pull)
			}
		} else {
			for {
				if !ss.openBlockHash.IsZero() {
					// We have an account but remote peer have not.
					ss.next()
				} else {
					if len(ss.frontiers) == 0 {
						var err error
						ss.frontiers, err = ss.getLocalFrontier(ss.qlcLedger)
						if err != nil {
							ss.logger.Error("get local frontier error")
						}
						ss.next()
					}

					if len(ss.bulkPull) > 0 {
						var index int
						var resend int
						if len(ss.pullTimer.C) > 0 {
							<-ss.pullTimer.C
						}
						ss.pullTimer.Reset(pullReqTimeOut)
						select {
						case ss.pullRequestStartCh <- true:
						default:
						}

						for {
							select {
							case <-ss.quitChanForSync:
								ss.logger.Warn("sync already finish,exit pull blocks loop")
								return topic.SyncFinish
							case <-ss.pullTimer.C:
								blkReq := &protos.BulkPullReqPacket{
									StartHash: ss.pullStartHash,
									EndHash:   ss.pullEndHash,
								}
								err := ss.netService.SendMessageToPeer(BulkPullRequest, blkReq, peerID)
								if err != nil {
									ss.logger.Errorf("err [%s] when send BulkPullRequest", err)
								}
								resend++
								ss.logger.Warnf("resend pull request startHash is [%s],endHash is [%s]", ss.pullStartHash, ss.pullEndHash)
								if resend == maxResendTime {
									ss.logger.Warn("resend pull request timeout")
									return topic.SyncDone
								}
								ss.pullTimer.Reset(pullReqTimeOut)
							case <-ss.pullRequestStartCh:
								if index == len(ss.bulkPull) {
									ss.pullStartHash = types.ZeroHash
									ss.pullEndHash = types.ZeroHash
									return topic.SyncDone
								}

								ss.pullStartHash = ss.bulkPull[index].StartHash
								ss.pullEndHash = ss.bulkPull[index].EndHash

								blkReq := &protos.BulkPullReqPacket{
									StartHash: ss.bulkPull[index].StartHash,
									EndHash:   ss.bulkPull[index].EndHash,
								}
								ss.logger.Warnf("pull request startHash is [%s],endHash is [%s]", ss.pullStartHash, ss.pullEndHash)
								err := ss.netService.SendMessageToPeer(BulkPullRequest, blkReq, peerID)
								if err != nil {
									ss.logger.Errorf("err [%s] when send BulkPullRequest", err)
								}
								index++
							}
						}
					} else {
						return topic.SyncFinish
					}
				}
			}
		}
	}
	return topic.SyncDone
}

func (ss *ServiceSync) getLocalFrontier(ledger *ledger.Ledger) ([]*types.Frontier, error) {
	frontiers, err := ledger.GetFrontiers()
	if err != nil {
		return nil, err
	}
	fsBack := new(types.Frontier)
	frontiers = append(frontiers, fsBack)
	return frontiers, nil
}

func (ss *ServiceSync) onBulkPullRequest(message *Message) error {
	exitPullRsp := &peerPullRsp{}
	if v, ok := ss.netService.msgService.pullRspMap.Load(message.from); ok {
		exitPullRsp, _ = v.(*peerPullRsp)
		select {
		case exitPullRsp.pullRspQuitCh <- true:
		default:
		}
	} else {
		exitPullRsp = &peerPullRsp{
			pullRspStartCh: make(chan bool, 1),
			pullRspTimer:   time.NewTimer(pullRspTimeOut),
			pullRspQuitCh:  make(chan bool, 1),
			muForPullRsp:   &sync.Mutex{},
		}
	}

	exitPullRsp.muForPullRsp.Lock()
	defer exitPullRsp.muForPullRsp.Unlock()

	var blk *types.StateBlock
	var bulkBlk types.StateBlockList
	var temp types.Hash
	pullRemote, err := protos.BulkPullReqPacketFromProto(message.Data())
	if err != nil {
		return err
	}
	ss.netService.node.logger.Debugf("receive BulkPullRequest, type %d start %s end %s count %d",
		pullRemote.PullType, pullRemote.StartHash, pullRemote.EndHash, pullRemote.Count)
	startHash := pullRemote.StartHash
	endHash := pullRemote.EndHash
	pullType := pullRemote.PullType
	openBlockHash, err := ss.getOpenBlockHash(endHash)
	if err != nil {
		ss.logger.Error(err)
		return err
	}
	if pullType != protos.PullTypeSegment {
		return ss.onBulkPullRequestExt(message, pullRemote)
	}

	if exitPullRsp != nil {
		exitPullRsp.pullRspTimer.Reset(pullRspTimeOut)
		if len(exitPullRsp.pullRspQuitCh) > 0 {
			<-exitPullRsp.pullRspQuitCh
		}
		select {
		case exitPullRsp.pullRspStartCh <- true:
		default:
		}
		exitPullRsp.pullRspHash = types.ZeroHash
		ss.netService.msgService.pullRspMap.Store(message.from, exitPullRsp)
	} else {
		return errors.New("exitPullRsp is nil")
	}

	if startHash.IsZero() {
		temp = openBlockHash
		for {
			blk, err = ss.qlcLedger.GetStateBlockConfirmed(temp)
			if err != nil {
				ss.logger.Errorf("err [%s] when get StateBlock:[%s]", err, temp.String())
				break
			}
			bulkBlk = append(bulkBlk, blk)
			if temp.String() == endHash.String() || len(bulkBlk) == maxPushTxPerTime {
				break
			}
			temp, err = ss.qlcLedger.GetBlockChild(temp)
			if err != nil {
				ss.logger.Error(err)
				return err
			}
		}
	} else {
		blk, err = ss.qlcLedger.GetStateBlockConfirmed(startHash)
		if err == ledger.ErrBlockNotFound {
			temp = openBlockHash
		} else {
			temp = startHash
		}
		for {
			blk, err = ss.qlcLedger.GetStateBlockConfirmed(temp)
			if err != nil {
				ss.logger.Errorf("err [%s] when get StateBlock:[%s]", err, temp.String())
				break
			}
			bulkBlk = append(bulkBlk, blk)
			if temp.String() == endHash.String() || len(bulkBlk) == maxPushTxPerTime {
				break
			}
			temp, err = ss.qlcLedger.GetBlockChild(temp)
			if err != nil {
				ss.logger.Error(err)
				return err
			}
		}
	}
	for {
		select {
		case <-exitPullRsp.pullRspTimer.C:
			return nil
		case <-exitPullRsp.pullRspQuitCh:
			ss.logger.Info("exit pullRsp loop because received another pullReq")
			return nil
		case <-exitPullRsp.pullRspStartCh:
			if !ss.netService.Node().streamManager.IsConnectWithPeerId(message.MessageFrom()) {
				return ErrStreamIsNotConnected
			}
			req := new(protos.BulkPullRspPacket)
			req.PullType = protos.PullTypeSegment
			req.Blocks = bulkBlk
			data, err := marshalMessage(BulkPullRsp, req)
			if err != nil {
				ss.logger.Error(err)
				return err
			}
			qData := NewQlcMessage(data, byte(p2pVersion), BulkPullRsp)
			exitPullRsp.pullRspHash, _ = types.HashBytes(qData)
			err = ss.netService.SendMessageToPeer(BulkPullRsp, req, message.MessageFrom())
			if err != nil {
				ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
			}
			exitPullRsp.pullRspTimer.Reset(pullRspTimeOut)
			if temp.String() == endHash.String() {
				return nil
			} else {
				bulkBlk = bulkBlk[:0:0]
				temp, err = ss.qlcLedger.GetBlockChild(temp)
				if err != nil {
					ss.logger.Error(err)
					return err
				}
				for {
					blk, err = ss.qlcLedger.GetStateBlockConfirmed(temp)
					if err != nil {
						ss.logger.Errorf("err [%s] when get StateBlock:[%s]", err, temp.String())
						break
					}
					bulkBlk = append(bulkBlk, blk)
					if temp.String() == endHash.String() || len(bulkBlk) == maxPushTxPerTime {
						break
					}
					temp, err = ss.qlcLedger.GetBlockChild(temp)
					if err != nil {
						ss.logger.Error(err)
						return err
					}
				}
			}
		}
	}
}

func (ss *ServiceSync) onBulkPullRequestExt(message *Message, pullRemote *protos.BulkPullReqPacket) error {
	var err error
	var blk *types.StateBlock
	var bulkBlk []*types.StateBlock

	pullType := pullRemote.PullType
	blkCnt := pullRemote.Count

	if pullType == protos.PullTypeBackward {
		scanHash := pullRemote.EndHash

		ss.logger.Debugf("need to send %d blocks by backward", blkCnt)

		for {
			blk, err = ss.qlcLedger.GetStateBlockConfirmed(scanHash)
			if err != nil {
				break
			}

			bulkBlk = append(bulkBlk, blk)
			blkCnt--
			if blkCnt <= 0 {
				break
			}

			scanHash = blk.GetPrevious()
		}
	} else if pullType == protos.PullTypeForward {
		startHash := pullRemote.StartHash
		if blkCnt == 0 {
			blkCnt = 1000
		}

		ss.logger.Debugf("need to send %d blocks by forward", blkCnt)

		blk, err = ss.qlcLedger.GetStateBlockConfirmed(startHash)
		if err != nil {
			return err
		}
		tm, err := ss.qlcLedger.GetTokenMeta(blk.GetAddress(), blk.GetToken())
		if err != nil {
			return err
		}

		scanHash := tm.Header
		for {
			if scanHash.IsZero() {
				break
			}

			blk, err = ss.qlcLedger.GetStateBlockConfirmed(scanHash)
			if err != nil {
				break
			}

			bulkBlk = append(bulkBlk, blk)
			if startHash == scanHash {
				break
			}

			scanHash = blk.GetPrevious()
		}

		if uint32(len(bulkBlk)) > blkCnt {
			bulkBlk = bulkBlk[:blkCnt]
		}
	} else if pullType == protos.PullTypeBatch {
		ss.logger.Debugf("need to send %d blocks by batch", blkCnt)

		for _, scanHash := range pullRemote.Hashes {
			if scanHash == nil {
				continue
			}

			blk, err = ss.qlcLedger.GetStateBlockConfirmed(*scanHash)
			if err != nil {
				continue
			}

			bulkBlk = append(bulkBlk, blk)
		}
	}
	reverseBlocks := make(types.StateBlockList, 0)
	for i := len(bulkBlk) - 1; i >= 0; i-- {
		reverseBlocks = append(reverseBlocks, bulkBlk[i])
	}
	var shardingBlocks types.StateBlockList
	for len(reverseBlocks) > 0 {
		sendBlockNum := 0
		if len(reverseBlocks) > maxPushTxPerTime {
			sendBlockNum = maxPushTxPerTime
		} else {
			sendBlockNum = len(reverseBlocks)
		}
		shardingBlocks = reverseBlocks[0:sendBlockNum]
		if !ss.netService.Node().streamManager.IsConnectWithPeerId(message.MessageFrom()) {
			break
		}
		req := new(protos.BulkPullRspPacket)
		req.PullType = protos.PullTypeBatch
		req.Blocks = shardingBlocks
		err = ss.netService.SendMessageToPeer(BulkPullRsp, req, message.MessageFrom())
		if err != nil {
			ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
		}
		reverseBlocks = reverseBlocks[sendBlockNum:]
	}
	return nil
}

func (ss *ServiceSync) onBulkPullRsp(message *Message) error {
	blkPacket, err := protos.BulkPullRspPacketFromProto(message.Data())
	if err != nil {
		return err
	}
	blocks := blkPacket.Blocks
	if len(blocks) == 0 {
		return nil
	}
	//if ss.netService.node.cfg.PerformanceEnabled {
	//	for _, b := range blocks {
	//		hash := b.GetHash()
	//		ss.netService.msgService.addPerformanceTime(hash)
	//	}
	//}

	for i, b := range blocks {
		ss.logger.Debugf("sync block acc[%s]-index[%d]-hash[%s]-prev[%s]", b.Address, i, b.GetHash(), b.Previous)
	}

	if blkPacket.PullType == protos.PullTypeSegment {
		ss.pullTimer.Stop()

		if ss.lastSyncHash == types.ZeroHash {
			firstBlock := blocks[0]
			if !firstBlock.IsOpen() {
				if has, _ := ss.qlcLedger.HasStateBlockConfirmed(firstBlock.Previous); !has {
					ss.logger.Errorf("get wrong sync block[%s] prev[%s] not in ledger", firstBlock.GetHash(), firstBlock.Previous)
					ss.pullTimer.Reset(time.Second)
					ss.lastSyncHash = types.ZeroHash
					return nil
				}
				ss.lastSyncHash = firstBlock.Previous
			}
		}

		for _, block := range blocks {
			if block.Previous != ss.lastSyncHash {
				ss.logger.Errorf("get wrong sync block[%s]-prev[%s]-expect[%s]", block.GetHash(), block.Previous, ss.lastSyncHash)
				ss.pullTimer.Reset(time.Second)
				ss.lastSyncHash = types.ZeroHash
				return nil
			}
			ss.lastSyncHash = block.GetHash()
		}

		ss.netService.msgEvent.Publish(topic.EventSyncBlock, blocks)
		ss.pullTimer.Reset(pullReqTimeOut)

		if ss.netService.Node().streamManager.IsConnectWithPeerId(message.MessageFrom()) {
			err = ss.netService.SendMessageToPeer(MessageResponse, message.Hash(), message.MessageFrom())
			if err != nil {
				ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
			}
		}

		if blocks[len(blocks)-1].GetHash() == ss.pullEndHash && ss.pullEndHash != types.ZeroHash {
			ss.lastSyncHash = types.ZeroHash
			select {
			case ss.pullRequestStartCh <- true:
			default:
			}
		}
	}
	return nil
}

func (ss *ServiceSync) onBulkPushBlock(message *Message) error {
	ss.netService.node.logger.Debug("receive BulkPushBlock")
	blkPacket, err := protos.BulkPushBlockFromProto(message.Data())
	if err != nil {
		return err
	}
	blocks := blkPacket.Blocks

	//if ss.netService.node.cfg.PerformanceEnabled {
	//	for _, b := range blocks {
	//		hash := b.GetHash()
	//		ss.netService.msgService.addPerformanceTime(hash)
	//	}
	//}
	ss.netService.msgEvent.Publish(topic.EventSyncBlock, blocks)
	return nil
}

func (ss *ServiceSync) next() {
	if len(ss.frontiers) > 0 {
		ss.openBlockHash = ss.frontiers[0].OpenBlock
		ss.headerBlockHash = ss.frontiers[0].HeaderBlock
		ss.frontiers = ss.frontiers[1:]
	}
}

func (ss *ServiceSync) requestFrontiersFromPov(peerID string) {
	ss.logger.Warn("request frontier from pov")
	syncState := ss.netService.cc.P2PSyncState()
	if syncState == topic.SyncFinish || syncState == topic.SyncNotStart {
		var err error
		address := types.Address{}
		Req := protos.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
		ss.logger.Warnf("begin sync block from [%s]", peerID)
		err = ss.netService.node.SendMessageToPeer(FrontierRequest, Req, peerID)
		if err != nil {
			ss.logger.Errorf("err [%s] when send FrontierRequest", err)
		}
	}
}

func (ss *ServiceSync) requestTxsByHashes(reqTxHashes []*types.Hash, peerID string) {
	if len(reqTxHashes) <= 0 {
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

		ss.netService.node.logger.Debugf("request txs %d from peer %s", len(sendTxHashes), peerID)
		if !ss.netService.Node().streamManager.IsConnectWithPeerId(peerID) {
			break
		}
		ss.netService.msgEvent.Publish(topic.EventSendMsgToSingle,
			&EventSendMsgToSingleMsg{Type: BulkPullRequest, Message: req, PeerID: peerID})

		reqTxHashes = reqTxHashes[sendHashNum:]
	}
}

func (ss *ServiceSync) getOpenBlockHash(hash types.Hash) (types.Hash, error) {
	blk, err := ss.qlcLedger.GetStateBlockConfirmed(hash)
	if err != nil {
		return types.ZeroHash, err
	}
	tm, err := ss.qlcLedger.GetTokenMetaConfirmed(blk.Address, blk.Token)
	if err != nil {
		return types.ZeroHash, err
	}
	return tm.OpenBlock, nil
}
