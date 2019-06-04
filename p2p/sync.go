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
	netService      *QlcService
	qlcLedger       *ledger.Ledger
	frontiers       []*types.Frontier
	remoteFrontiers []*types.Frontier
	quitCh          chan bool
	logger          *zap.SugaredLogger
	lastSyncTime    int64
}

// NewService return new Service.
func NewSyncService(netService *QlcService, ledger *ledger.Ledger) *ServiceSync {
	ss := &ServiceSync{
		netService:   netService,
		qlcLedger:    ledger,
		quitCh:       make(chan bool, 1),
		logger:       log.NewLogger("sync"),
		lastSyncTime: 0,
	}
	return ss
}

func (ss *ServiceSync) Start() {
	ss.logger.Info("started sync loop")
	address := types.Address{}
	Req := protos.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	ticker := time.NewTicker(time.Duration(ss.netService.node.cfg.P2P.SyncInterval) * time.Second)
	for {
		select {
		case <-ss.quitCh:
			ss.logger.Info("Stopped Sync Loop.")
			return
		case <-ticker.C:
			now := time.Now().UTC().Unix()
			v := atomic.LoadInt64(&ss.lastSyncTime)
			if v < now {
				peerID, err := ss.netService.node.StreamManager().RandomPeer()
				if err != nil {
					continue
				}
				ss.frontiers, err = getLocalFrontier(ss.qlcLedger)
				if err != nil {
					continue
				}
				ss.logger.Infof("begin sync block from [%s]", peerID)
				ss.remoteFrontiers = ss.remoteFrontiers[:0:0]
				ss.next()
				bulkPull = bulkPull[:0:0]
				bulkPush = bulkPush[:0:0]
				err = ss.netService.node.SendMessageToPeer(FrontierRequest, Req, peerID)
				if err != nil {
					ss.logger.Errorf("err [%s] when send FrontierRequest", err)
				}
			}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (ss *ServiceSync) LastSyncTime(t time.Time) {
	atomic.StoreInt64(&ss.lastSyncTime, t.Add(syncTimeout).UTC().Unix())
}

// Stop sync service
func (ss *ServiceSync) Stop() {
	//ss.logger.Info("Stop Qlc sync...")

	ss.quitCh <- true
}

func (ss *ServiceSync) onFrontierReq(message *Message) error {
	ss.netService.node.logger.Debug("receive FrontierReq")
	now := time.Now().UTC().Unix()
	v := atomic.LoadInt64(&ss.lastSyncTime)
	if v < now {
		var fs []*types.Frontier
		fs, err := ss.qlcLedger.GetFrontiers()
		if err != nil {
			return err
		}
		num := len(fs)
		var rsp *protos.FrontierResponse
		for _, f := range fs {
			rsp = protos.NewFrontierRsp(f, uint32(num))
			err = ss.netService.SendMessageToPeer(FrontierRsp, rsp, message.MessageFrom())
			if err != nil {
				ss.logger.Errorf("send FrontierRsp err [%s]", err)
			}
		}
	}
	//send frontier finished,last frontier is all zero,tell remote peer send finished
	//zeroFrontier := new(types.Frontier)
	//err = ss.netService.SendMessageToPeer(FrontierRsp, zeroFrontier, message.MessageFrom())
	//if err != nil {
	//	ss.logger.Errorf("send FrontierRsp err [%s] for zeroFrontier", err)
	//}
	return nil
}

func (ss *ServiceSync) checkFrontier(message *Message) {
	rsp, err := protos.FrontierResponseFromProto(message.Data())
	if err != nil {
		ss.logger.Error(err)
		return
	}
	if uint32(len(ss.remoteFrontiers)) < rsp.TotalFrontierNum {
		ss.remoteFrontiers = append(ss.remoteFrontiers, rsp.Frontier)
		if uint32(len(ss.remoteFrontiers)) == rsp.TotalFrontierNum {
			var remoteFrontiers []*types.Frontier
			remoteFrontiers = append(remoteFrontiers, ss.remoteFrontiers...)
			sort.Sort(types.Frontiers(remoteFrontiers))
			zeroFrontier := new(types.Frontier)
			remoteFrontiers = append(remoteFrontiers, zeroFrontier)
			ss.remoteFrontiers = ss.remoteFrontiers[:0:0]
			go func() {
				err := ss.processFrontiers(remoteFrontiers, message.MessageFrom())
				if err != nil {
					ss.logger.Error("process frontiers error")
				}
			}()
		}
		return
	}

}

func (ss *ServiceSync) processFrontiers(fsRemotes []*types.Frontier, peerID string) error {
	//ss.netService.node.logger.Info("receive FrontierRsp")
	//fsRemote, err := protos.FrontierResponseFromProto(message.Data())
	//if err != nil {
	//	return err
	//}
	//fr := fsRemote.Frontier
	//ss.logger.Info(fr.HeaderBlock, fr.OpenBlock)
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
						exit, _ := ss.qlcLedger.HasStateBlock(fsRemotes[i].HeaderBlock)
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
									return err
								}
								bulkBlk = append(bulkBlk, blk)
								endHash = blk.GetPrevious()
								if endHash.IsZero() == true {
									break
								}
							}
							for i := len(bulkBlk) - 1; i >= 0; i-- {
								err = ss.netService.SendMessageToPeer(BulkPushBlock, bulkBlk[i], peerID)
								if err != nil {
									ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
								}
							}
						} else {
							//ss.logger.Info("need to send some blocks of this account")
							var blk *types.StateBlock
							var bulkBlk []*types.StateBlock
							for {
								blk, err = ss.qlcLedger.GetStateBlock(endHash)
								if err != nil {
									return err
								}
								bulkBlk = append(bulkBlk, blk)

								endHash = blk.GetPrevious()
								if endHash == startHash {
									break
								}
							}
							for i := len(bulkBlk) - 1; i >= 0; i-- {
								err = ss.netService.SendMessageToPeer(BulkPushBlock, bulkBlk[i], peerID)
								if err != nil {
									ss.logger.Errorf("err [%s] when send BulkPushBlock", err)
								}
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
	ss.netService.node.logger.Debugf("receive BulkPullRequest, start %s end %s", pullRemote.StartHash, pullRemote.EndHash)

	startHash := pullRemote.StartHash
	endHash := pullRemote.EndHash
	if startHash.IsZero() {
		var blk *types.StateBlock
		var bulkBlk []*types.StateBlock
		//ss.logger.Info("need to send all the blocks of this account")
		for {
			blk, err = ss.qlcLedger.GetStateBlock(endHash)
			if err != nil {
				return err
			}
			bulkBlk = append(bulkBlk, blk)
			endHash = blk.GetPrevious()
			if endHash.IsZero() == true {
				break
			}
		}
		for i := len(bulkBlk) - 1; i >= 0; i-- {
			err = ss.netService.SendMessageToPeer(BulkPullRsp, bulkBlk[i], message.MessageFrom())
			if err != nil {
				ss.logger.Errorf("err [%s] when send BulkPullRsp", err)
			}
		}
	} else {
		var blk *types.StateBlock
		var bulkBlk []*types.StateBlock
		//ss.logger.Info("need to send some blocks of this account")
		for {
			blk, err = ss.qlcLedger.GetStateBlock(endHash)
			if err != nil {
				return err
			}
			bulkBlk = append(bulkBlk, blk)
			if endHash == startHash {
				break
			}
			endHash = blk.GetPrevious()
		}
		for i := len(bulkBlk) - 1; i >= 0; i-- {
			err = ss.netService.SendMessageToPeer(BulkPullRsp, bulkBlk[i], message.MessageFrom())
			if err != nil {
				ss.logger.Errorf("err [%s] when send BulkPullRsp", err)
			}
		}
	}
	return nil
}

func (ss *ServiceSync) onBulkPullRsp(message *Message) error {
	blkPacket, err := protos.BulkPushBlockFromProto(message.Data())
	if err != nil {
		return err
	}

	block := blkPacket.Blk
	if block == nil {
		return nil
	}

	ss.netService.node.logger.Debugf("receive BulkPullRsp, hash %s", block.GetHash())

	if ss.netService.node.cfg.PerformanceEnabled {
		hash := block.GetHash()
		ss.netService.msgService.addPerformanceTime(hash)
	}
	ss.netService.msgEvent.Publish(string(common.EventSyncBlock), block)
	return nil
}

func (ss *ServiceSync) onBulkPushBlock(message *Message) error {
	ss.netService.node.logger.Debug("receive BulkPushBlock")
	blkPacket, err := protos.BulkPushBlockFromProto(message.Data())
	if err != nil {
		return err
	}
	block := blkPacket.Blk

	if ss.netService.node.cfg.PerformanceEnabled {
		hash := block.GetHash()
		ss.netService.msgService.addPerformanceTime(hash)
	}
	ss.netService.msgEvent.Publish(string(common.EventSyncBlock), block)
	return nil
}

func (ss *ServiceSync) next() {
	if len(ss.frontiers) > 0 {
		openBlockHash = ss.frontiers[0].OpenBlock
		headerBlockHash = ss.frontiers[0].HeaderBlock
		ss.frontiers = ss.frontiers[1:]
	}
}
