package p2p

import (
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
	headerBlockHash    types.Hash
	openBlockHash      types.Hash
	bulkPush, bulkPull []*protos.Bulk
)

const syncTimeout = 10 * time.Second

// Service manage sync tasks
type ServiceSync struct {
	netService   *QlcService
	qlcLedger    *ledger.Ledger
	frontiers    []*types.Frontier
	quitCh       chan bool
	logger       *zap.SugaredLogger
	lastSyncTime int64
	syncCount    uint32
	syncState    common.SyncState
	syncTimer	 *time.Timer
}

// NewService return new Service.
func NewSyncService(netService *QlcService, ledger *ledger.Ledger) *ServiceSync {
	ss := &ServiceSync{
		netService:   netService,
		qlcLedger:    ledger,
		quitCh:       make(chan bool, 1),
		logger:       log.NewLogger("sync"),
		lastSyncTime: 0,
		syncCount:    0,
		syncState:    common.SyncNotStart,
	}
	return ss
}

func (ss *ServiceSync) Start() {
	ss.logger.Info("started sync loop")
	address := types.Address{}
	Req := protos.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	ss.syncTimer = time.NewTimer(time.Duration(ss.netService.node.cfg.P2P.SyncInterval) * time.Second)

	err := ss.netService.msgEvent.SubscribeSync(common.EventConsensusSyncFinished, ss.onConsensusSyncFinished)
	if err != nil {
		ss.logger.Errorf("subscribe consensus sync finished event err")
	}

	for {
		select {
		case <-ss.quitCh:
			ss.logger.Info("Stopped Sync Loop.")
			return
		case <-ss.syncTimer.C:
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
			bulkPush = bulkPush[:0:0]
			err = ss.netService.node.SendMessageToPeer(FrontierRequest, Req, peerID)
			if err != nil {
				ss.logger.Errorf("err [%s] when send FrontierRequest", err)
			}
			ss.syncCount++
		}
	}
}

func (ss *ServiceSync) LastSyncTime(t time.Time) {
	atomic.StoreInt64(&ss.lastSyncTime, t.Add(syncTimeout).Unix())
}

// Stop sync service
func (ss *ServiceSync) Stop() {
	//ss.logger.Info("Stop Qlc sync...")

	ss.quitCh <- true
}

func (ss *ServiceSync) onConsensusSyncFinished() {
	ss.syncState = common.SyncFinish
	ss.syncTimer.Reset(time.Duration(ss.netService.node.cfg.P2P.SyncInterval) * time.Second)
}

func (ss *ServiceSync) onFrontierReq(message *Message) error {
	ss.netService.node.logger.Debug("receive FrontierReq")
	now := time.Now().Unix()
	v := atomic.LoadInt64(&ss.lastSyncTime)
	if v < now {
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
	}
	return nil
}

func (ss *ServiceSync) checkFrontier(message *Message) {
	rsp, err := protos.FrontierResponseFromProto(message.Data())
	if err != nil {
		ss.logger.Error(err)
		return
	}

	go func() {
		var remoteFrontiers []*types.Frontier
		var blks types.StateBlockList
		var confirmed bool
		ss.syncState = common.Syncing
		ss.netService.msgEvent.Publish(common.EventSyncStateChange, common.Syncing)
		ss.logger.Info("sync start")

		for {
			remoteFrontiers = remoteFrontiers[0:0]
			blks = blks[0:0]

			for _, f := range rsp.Fs {
				ss.netService.msgEvent.Publish(common.EventFrontierConfirmed, f.HeaderBlk.GetHash(), &confirmed)
				if confirmed {
					continue
				}

				remoteFrontiers = append(remoteFrontiers, f.Fr)
				blks = append(blks, f.HeaderBlk)
			}

			remoteFrontiersLen := len(remoteFrontiers)
			ss.logger.Infof("req %d frontiers", remoteFrontiersLen)

			if remoteFrontiersLen > 0 {
				ss.netService.msgEvent.Publish(common.EventFrontierConsensus, blks)
				sort.Sort(types.Frontiers(remoteFrontiers))
				zeroFrontier := new(types.Frontier)
				remoteFrontiers = append(remoteFrontiers, zeroFrontier)
				//go func() {
				//	err := ss.processFrontiers(remoteFrontiers, message.MessageFrom())
				//	if err != nil {
				//		ss.logger.Errorf("process frontiers error:[%s]", err)
				//	}
				//}()
				err := ss.processFrontiers(remoteFrontiers, message.MessageFrom())
				if err != nil {
					ss.logger.Errorf("process frontiers error:[%s]", err)
				}

				time.Sleep(time.Duration(ss.netService.node.cfg.P2P.SyncInterval) * time.Second)
			} else {
				ss.syncState = common.SyncDone
				ss.netService.msgEvent.Publish(common.EventSyncStateChange, common.SyncDone)
				ss.logger.Infof("sync pull all blocks done")
				break
			}
		}
	}()
}

func (ss *ServiceSync) processFrontiers(fsRemotes []*types.Frontier, peerID string) error {
	for i := 0; i < len(fsRemotes); i++ {
		if !fsRemotes[i].OpenBlock.IsZero() {
			for {
				if !openBlockHash.IsZero() && (openBlockHash.String() < fsRemotes[i].OpenBlock.String()) {
					// We have an account but remote peer have not.
					push := &protos.Bulk{
						StartHash: types.ZeroHash,
						EndHash:   headerBlockHash,
					}
					bulkPush = append(bulkPush, push)
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
						if exit == true {
							push := &protos.Bulk{
								StartHash: fsRemotes[i].HeaderBlock,
								EndHash:   headerBlockHash,
							}
							bulkPush = append(bulkPush, push)
						} else {
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
					push := &protos.Bulk{
						StartHash: types.ZeroHash,
						EndHash:   headerBlockHash,
					}
					bulkPush = append(bulkPush, push)
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
					for _, value := range bulkPull {
						blkReq := &protos.BulkPullReqPacket{
							StartHash: value.StartHash,
							EndHash:   value.EndHash,
						}
						err := ss.netService.SendMessageToPeer(BulkPullRequest, blkReq, peerID)
						if err != nil {
							ss.logger.Errorf("err [%s] when send BulkPullRequest", err)
						}
					}
					for _, value := range bulkPush {
						startHash := value.StartHash
						endHash := value.EndHash
						var err error
						if startHash.IsZero() {
							//ss.logger.Infof("need to send all the blocks of this account")
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
							for len(reverseBlocks) > 0 {
								sendBlockNum := 0
								if len(reverseBlocks) > maxPushTxPerTime {
									sendBlockNum = maxPushTxPerTime
								} else {
									sendBlockNum = len(reverseBlocks)
								}
								sendTxBlocks := reverseBlocks[0:sendBlockNum]
								if !ss.netService.Node().streamManager.IsConnectWithPeerId(peerID) {
									break
								}
								err = ss.netService.SendMessageToPeer(BulkPushBlock, sendTxBlocks, peerID)
								if err != nil {
									ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
								}
								reverseBlocks = reverseBlocks[sendBlockNum:]
							}
							//for i := len(bulkBlk) - 1; i >= 0; i-- {
							//	if !ss.netService.Node().streamManager.IsConnectWithPeerId(peerID) {
							//		break
							//	}
							//	err = ss.netService.SendMessageToPeer(BulkPushBlock, bulkBlk[i], peerID)
							//	if err != nil {
							//		ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
							//	}
							//}
						} else {
							//ss.logger.Info("need to send some blocks of this account")
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
								if endHash == startHash {
									break
								}
							}
							reverseBlocks := make(types.StateBlockList, 0)
							for i := len(bulkBlk) - 1; i >= 0; i-- {
								reverseBlocks = append(reverseBlocks, bulkBlk[i])
							}
							for len(reverseBlocks) > 0 {
								sendBlockNum := 0
								if len(reverseBlocks) > maxPushTxPerTime {
									sendBlockNum = maxPushTxPerTime
								} else {
									sendBlockNum = len(reverseBlocks)
								}
								sendTxBlocks := reverseBlocks[0:sendBlockNum]
								if !ss.netService.Node().streamManager.IsConnectWithPeerId(peerID) {
									break
								}
								err = ss.netService.SendMessageToPeer(BulkPushBlock, sendTxBlocks, peerID)
								if err != nil {
									ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
								}
								reverseBlocks = reverseBlocks[sendBlockNum:]
							}
							//for i := len(bulkBlk) - 1; i >= 0; i-- {
							//	if !ss.netService.Node().streamManager.IsConnectWithPeerId(peerID) {
							//		break
							//	}
							//	err = ss.netService.SendMessageToPeer(BulkPushBlock, bulkBlk[i], peerID)
							//	if err != nil {
							//		ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
							//	}
							//}
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
		//ss.logger.Info("need to send all the blocks of this account")
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
			err = ss.netService.SendMessageToPeer(BulkPullRsp, shardingBlocks, message.MessageFrom())
			if err != nil {
				ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
			}
			reverseBlocks = reverseBlocks[sendBlockNum:]
		}
		//for i := len(bulkBlk) - 1; i >= 0; i-- {
		//	if !ss.netService.Node().streamManager.IsConnectWithPeerId(message.MessageFrom()) {
		//		break
		//	}
		//	err = ss.netService.SendMessageToPeer(BulkPullRsp, bulkBlk[i], message.MessageFrom())
		//	if err != nil {
		//		ss.logger.Errorf("err [%s] when send BulkPullRsp", err)
		//	}
		//}
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
		if !ss.netService.Node().streamManager.IsConnectWithPeerId(message.MessageFrom()) {
			return nil
		}
		err = ss.netService.SendMessageToPeer(BulkPullRsp, bulkBlk, message.MessageFrom())
		if err != nil {
			ss.logger.Errorf("err [%s] when send BulkPullRsp", err)
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
		err = ss.netService.SendMessageToPeer(BulkPullRsp, shardingBlocks, message.MessageFrom())
		if err != nil {
			ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
		}
		reverseBlocks = reverseBlocks[sendBlockNum:]
	}
	//for i := len(bulkBlk) - 1; i >= 0; i-- {
	//	if !ss.netService.Node().streamManager.IsConnectWithPeerId(message.MessageFrom()) {
	//		break
	//	}
	//
	//	err = ss.netService.SendMessageToPeer(BulkPullRsp, bulkBlk[i], message.MessageFrom())
	//	if err != nil {
	//		ss.logger.Errorf("err [%s] when send BulkPullRsp", err)
	//	}
	//}

	return nil
}

func (ss *ServiceSync) onBulkPullRsp(message *Message) error {
	blkPacket, err := protos.BulkPushBlockFromProto(message.Data())
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
