package p2p

import (
	"errors"
	"math"
	"sort"
	"sync/atomic"
	"time"

	"github.com/qlcchain/go-qlc/common"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"go.uber.org/zap"
)

var (
	headerBlockHash types.Hash
	openBlockHash   types.Hash
	bulkPull        []*protos.Bulk
)

const (
	maxResendTime  = 5
	pullRspTimeOut = 30 * time.Second
	pullReqTimeOut = 60 * time.Second
)

var (
	ErrSyncTimeOut = errors.New("sync time out")
)

// Service manage sync tasks
type ServiceSync struct {
	netService         *QlcService
	qlcLedger          *ledger.Ledger
	frontiers          []*types.Frontier
	quitCh             chan bool
	logger             *zap.SugaredLogger
	syncState          atomic.Value
	syncTicker         *time.Ticker
	pullTimer          *time.Timer
	pullRequestStartCh chan bool
	pullStartHash      types.Hash
	pullEndHash        types.Hash
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
	}
	ss.syncState.Store(common.SyncNotStart)
	return ss
}

func (ss *ServiceSync) Start() {
	ss.logger.Info("started sync loop")
	address := types.Address{}
	Req := protos.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	ss.syncTicker = time.NewTicker(time.Duration(ss.netService.node.cfg.P2P.SyncInterval) * time.Second)

	err := ss.netService.msgEvent.SubscribeSync(common.EventConsensusSyncFinished, ss.onConsensusSyncFinished)
	if err != nil {
		ss.logger.Errorf("subscribe consensus sync finished event err")
	}

	for {
		select {
		case <-ss.quitCh:
			ss.logger.Info("Stopped Sync Loop.")
			return
		case <-ss.syncTicker.C:
			syncState := ss.syncState.Load()
			if syncState == common.SyncFinish || syncState == common.SyncNotStart {
				peerID, err := ss.netService.node.StreamManager().RandomPeer()
				if err != nil {
					continue
				}
				ss.frontiers, err = getLocalFrontier(ss.qlcLedger)
				if err != nil {
					continue
				}
				ss.logger.Infof("begin sync block from [%s]", peerID)
				ss.next()
				bulkPull = bulkPull[:0:0]
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
	//ss.logger.Info("Stop Qlc sync...")

	ss.quitCh <- true
}

func (ss *ServiceSync) onConsensusSyncFinished() {
	ss.syncState.Store(common.SyncFinish)
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
	syncState := ss.syncState.Load()
	if syncState != common.Syncing {
		rsp, err := protos.FrontierResponseFromProto(message.Data())
		if err != nil {
			ss.logger.Error(err)
			return
		}

		go func() {
			var remoteFrontiers []*types.Frontier
			var blks types.StateBlockList
			ss.syncState.Store(common.Syncing)
			ss.netService.msgEvent.Publish(common.EventSyncStateChange, common.Syncing)
			ss.logger.Info("sync start")

			for _, f := range rsp.Fs {
				remoteFrontiers = append(remoteFrontiers, f.Fr)
				blks = append(blks, f.HeaderBlk)
			}
			remoteFrontiersLen := len(remoteFrontiers)

			if remoteFrontiersLen > 0 {
				ss.netService.msgEvent.Publish(common.EventFrontierConsensus, blks)
				sort.Sort(types.Frontiers(remoteFrontiers))
				zeroFrontier := new(types.Frontier)
				remoteFrontiers = append(remoteFrontiers, zeroFrontier)
				err := ss.processFrontiers(remoteFrontiers, message.MessageFrom())
				if err == ErrSyncTimeOut {
					ss.logger.Errorf("process frontiers error:[%s]", err)
					ss.syncState.Store(common.SyncFinish)
					ss.netService.msgEvent.Publish(common.EventSyncStateChange, common.SyncFinish)
				} else {
					ss.syncState.Store(common.SyncDone)
					ss.netService.msgEvent.Publish(common.EventSyncStateChange, common.SyncDone)
				}
			}
			ss.logger.Infof("sync pull all blocks done")
		}()
	}
}

func (ss *ServiceSync) processFrontiers(fsRemotes []*types.Frontier, peerID string) error {
	for i := 0; i < len(fsRemotes); i++ {
		if !fsRemotes[i].OpenBlock.IsZero() {
			for {
				if !openBlockHash.IsZero() && (openBlockHash.String() < fsRemotes[i].OpenBlock.String()) {
					// We have an account but remote peer have not.
					ss.next()
				} else {
					break
				}
			}
			if !openBlockHash.IsZero() {
				if fsRemotes[i].OpenBlock == openBlockHash {
					if headerBlockHash == fsRemotes[i].HeaderBlock {
						//ss.logger.Infof("this token %s have the same block", openBlockHash)
					} else {
						exit, _ := ss.qlcLedger.HasStateBlockConfirmed(fsRemotes[i].HeaderBlock)
						if !exit {
							pull := &protos.Bulk{
								StartHash: headerBlockHash,
								EndHash:   fsRemotes[i].HeaderBlock,
							}
							bulkPull = append(bulkPull, pull)
						}
					}
					ss.next()
				} else {
					if fsRemotes[i].OpenBlock.String() > openBlockHash.String() {
						return nil
					}
					pull := &protos.Bulk{
						StartHash: types.ZeroHash,
						EndHash:   fsRemotes[i].HeaderBlock,
					}
					bulkPull = append(bulkPull, pull)
				}
			} else {
				pull := &protos.Bulk{
					StartHash: types.ZeroHash,
					EndHash:   fsRemotes[i].HeaderBlock,
				}
				bulkPull = append(bulkPull, pull)
			}
		} else {
			for {
				if !openBlockHash.IsZero() {
					// We have an account but remote peer have not.
					ss.next()
				} else {
					if len(ss.frontiers) == 0 {
						var err error
						ss.frontiers, err = getLocalFrontier(ss.qlcLedger)
						if err != nil {
							ss.logger.Error("get local frontier error")
						}
						ss.next()
					}

					if len(bulkPull) > 0 {
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
								ss.logger.Infof("resend pull request startHash is [%s],endHash is [%s]\n", ss.pullStartHash, ss.pullEndHash)
								if resend == maxResendTime {
									ss.logger.Infof("resend timeout......")
									return ErrSyncTimeOut
								}
								ss.pullTimer.Reset(pullReqTimeOut)
							case <-ss.pullRequestStartCh:
								ss.pullStartHash = bulkPull[index].StartHash
								ss.pullEndHash = bulkPull[index].EndHash
								blkReq := &protos.BulkPullReqPacket{
									StartHash: bulkPull[index].StartHash,
									EndHash:   bulkPull[index].EndHash,
								}
								ss.logger.Infof("pull request startHash is [%s],endHash is [%s]\n", ss.pullStartHash, ss.pullEndHash)
								err := ss.netService.SendMessageToPeer(BulkPullRequest, blkReq, peerID)
								if err != nil {
									ss.logger.Errorf("err [%s] when send BulkPullRequest", err)
								}
								index++
							}
							if index == len(bulkPull) {
								ss.pullStartHash = types.ZeroHash
								ss.pullEndHash = types.ZeroHash
								break
							}
						}
					}
					break
				}
			}
		}
	}
	return nil
}

func getLocalFrontier(ledger *ledger.Ledger) ([]*types.Frontier, error) {
	frontiers, err := ledger.GetFrontiers()
	if err != nil {
		return nil, err
	}
	fsBack := new(types.Frontier)
	frontiers = append(frontiers, fsBack)
	return frontiers, nil
}

func (ss *ServiceSync) onBulkPullRequest(message *Message) error {
	pullRemote, err := protos.BulkPullReqPacketFromProto(message.Data())
	if err != nil {
		return err
	}
	ss.netService.node.logger.Debugf("receive BulkPullRequest, type %d start %s end %s count %d",
		pullRemote.PullType, pullRemote.StartHash, pullRemote.EndHash, pullRemote.Count)
	startHash := pullRemote.StartHash
	endHash := pullRemote.EndHash
	pullType := pullRemote.PullType

	if pullType != protos.PullTypeSegment {
		return ss.onBulkPullRequestExt(message, pullRemote)
	}
	if startHash.IsZero() {
		var blk *types.StateBlock
		var bulkBlk []*types.StateBlock
		for {
			blk, err = ss.qlcLedger.GetStateBlock(endHash)
			if err != nil {
				ss.logger.Errorf("err when get StateBlock:[%s]", endHash.String())
				break
			}
			bulkBlk = append(bulkBlk, blk)
			endHash = blk.GetPrevious()
			if endHash.IsZero() == true {
				break
			}
		}
		reverseBlocks := make(types.StateBlockList, 0)
		for i := len(bulkBlk) - 1; i >= 0; i-- {
			reverseBlocks = append(reverseBlocks, bulkBlk[i])
		}
		var shardingBlocks types.StateBlockList

		exitPullRsp := &peerPullRsp{}
		if v, ok := ss.netService.msgService.pullRspMap.Load(message.from); ok {
			exitPullRsp, _ = v.(*peerPullRsp)
			if len(exitPullRsp.pullRspTimer.C) > 0 {
				<-exitPullRsp.pullRspTimer.C
			}
			exitPullRsp.pullRspTimer.Reset(pullRspTimeOut)

		} else {
			exitPullRsp = &peerPullRsp{
				pullRspStartCh: make(chan bool, 1),
				pullRspTimer:   time.NewTimer(pullRspTimeOut),
			}
		}
		select {
		case exitPullRsp.pullRspStartCh <- true:
		default:
		}
		exitPullRsp.pullRspHash = types.ZeroHash
		ss.netService.msgService.pullRspMap.Store(message.from, exitPullRsp)
		for len(reverseBlocks) > 0 {
			sendBlockNum := 0
			select {
			case <-exitPullRsp.pullRspTimer.C:
				return nil
			case <-exitPullRsp.pullRspStartCh:
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
				req.PullType = protos.PullTypeSegment
				req.Blocks = shardingBlocks
				data, err := marshalMessage(BulkPullRsp, req)
				if err != nil {
					ss.logger.Error(err)
					break
				}
				qData := NewQlcMessage(data, byte(p2pVersion), BulkPullRsp)
				exitPullRsp.pullRspHash, _ = types.HashBytes(qData)
				err = ss.netService.SendMessageToPeer(BulkPullRsp, req, message.MessageFrom())
				if err != nil {
					ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
				}
				reverseBlocks = reverseBlocks[sendBlockNum:]
				exitPullRsp.pullRspTimer.Reset(pullRspTimeOut)
			default:
			}
		}

	} else {
		var blk *types.StateBlock
		var bulkBlk types.StateBlockList
		//ss.logger.Info("need to send some blocks of this account")
		for {
			blk, err = ss.qlcLedger.GetStateBlock(endHash)
			if err != nil {
				ss.logger.Errorf("err when get StateBlock:[%s]", endHash.String())
				break
			}
			bulkBlk = append(bulkBlk, blk)
			endHash = blk.GetPrevious()
			if endHash == startHash {
				break
			}
		}
		reverseBlocks := make(types.StateBlockList, 0)
		for i := len(bulkBlk) - 1; i >= 0; i-- {
			reverseBlocks = append(reverseBlocks, bulkBlk[i])
		}
		var shardingBlocks types.StateBlockList
		exitPullRsp := &peerPullRsp{}
		if v, ok := ss.netService.msgService.pullRspMap.Load(message.from); ok {
			exitPullRsp, _ = v.(*peerPullRsp)
			if len(exitPullRsp.pullRspTimer.C) > 0 {
				<-exitPullRsp.pullRspTimer.C
			}
			exitPullRsp.pullRspTimer.Reset(pullRspTimeOut)
		} else {
			exitPullRsp = &peerPullRsp{
				pullRspStartCh: make(chan bool, 1),
				pullRspTimer:   time.NewTimer(pullRspTimeOut),
			}
		}
		select {
		case exitPullRsp.pullRspStartCh <- true:
		default:
		}
		exitPullRsp.pullRspHash = types.ZeroHash
		ss.netService.msgService.pullRspMap.Store(message.from, exitPullRsp)
		for len(reverseBlocks) > 0 {
			sendBlockNum := 0
			select {
			case <-exitPullRsp.pullRspTimer.C:
				return nil
			case <-exitPullRsp.pullRspStartCh:
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
				req.PullType = protos.PullTypeSegment
				req.Blocks = shardingBlocks
				data, err := marshalMessage(BulkPullRsp, req)
				if err != nil {
					continue
				}
				qData := NewQlcMessage(data, byte(p2pVersion), BulkPullRsp)
				exitPullRsp.pullRspHash, _ = types.HashBytes(qData)
				err = ss.netService.SendMessageToPeer(BulkPullRsp, req, message.MessageFrom())
				if err != nil {
					ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
				}
				reverseBlocks = reverseBlocks[sendBlockNum:]
				exitPullRsp.pullRspTimer.Reset(pullRspTimeOut)
			default:
			}
		}
	}
	return nil
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
			blk, err = ss.qlcLedger.GetStateBlock(scanHash)
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

		blk, err = ss.qlcLedger.GetStateBlock(startHash)
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

			blk, err = ss.qlcLedger.GetStateBlock(scanHash)
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

			blk, err = ss.qlcLedger.GetStateBlock(*scanHash)
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
	if ss.netService.node.cfg.PerformanceEnabled {
		for _, b := range blocks {
			hash := b.GetHash()
			ss.netService.msgService.addPerformanceTime(hash)
		}
	}
	ss.netService.msgEvent.Publish(common.EventSyncBlock, blocks)
	if blkPacket.PullType == protos.PullTypeSegment {
		ss.pullTimer.Reset(pullReqTimeOut)
		if ss.netService.Node().streamManager.IsConnectWithPeerId(message.MessageFrom()) {
			err = ss.netService.SendMessageToPeer(MessageResponse, message.Hash(), message.MessageFrom())
			if err != nil {
				ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
			}
		}
		if blocks[len(blocks)-1].GetHash().String() == ss.pullEndHash.String() &&
			(ss.pullEndHash.String() != types.ZeroHash.String()) {
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

	if ss.netService.node.cfg.PerformanceEnabled {
		for _, b := range blocks {
			hash := b.GetHash()
			ss.netService.msgService.addPerformanceTime(hash)
		}
	}
	ss.netService.msgEvent.Publish(common.EventSyncBlock, blocks)
	return nil
}

func (ss *ServiceSync) next() {
	if len(ss.frontiers) > 0 {
		openBlockHash = ss.frontiers[0].OpenBlock
		headerBlockHash = ss.frontiers[0].HeaderBlock
		ss.frontiers = ss.frontiers[1:]
	}
}

func (ss *ServiceSync) requestFrontiersFromPov(peerID string) {
	ss.logger.Info("request frontier from pov")
	syncState := ss.syncState.Load()
	if syncState == common.SyncFinish || syncState == common.SyncNotStart {
		var err error
		address := types.Address{}
		Req := protos.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
		ss.frontiers, err = getLocalFrontier(ss.qlcLedger)
		if err != nil {
			ss.logger.Errorf("get error when process request frontiers from pov [%s]", err)
		}
		ss.logger.Infof("begin sync block from [%s]", peerID)
		ss.next()
		bulkPull = bulkPull[:0:0]
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
		ss.netService.msgEvent.Publish(common.EventSendMsgToSingle, BulkPullRequest, req, peerID)

		reqTxHashes = reqTxHashes[sendHashNum:]
	}
}
