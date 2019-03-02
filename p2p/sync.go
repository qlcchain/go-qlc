package p2p

import (
	"math"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"go.uber.org/zap"
)

const (
	SyncInterval = time.Minute * 2
)

var zeroHash = types.Hash{}
var headerBlockHash types.Hash
var openBlockHash types.Hash
var bulkPush, bulkPull []*protos.Bulk

// Service manage sync tasks
type ServiceSync struct {
	netService *QlcService
	qlcLedger  *ledger.Ledger
	frontiers  []*types.Frontier
	quitCh     chan bool
	logger     *zap.SugaredLogger
}

// NewService return new Service.
func NewSyncService(netService *QlcService, ledger *ledger.Ledger) *ServiceSync {
	ss := &ServiceSync{
		netService: netService,
		qlcLedger:  ledger,
		quitCh:     make(chan bool, 1),
		logger:     log.NewLogger("sync"),
	}
	return ss
}

func (ss *ServiceSync) Start() {
	ss.logger.Info("started sync loop")
	address := types.Address{}
	Req := protos.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	ticker := time.NewTicker(SyncInterval)
	for {
		select {
		case <-ss.quitCh:
			ss.logger.Info("Stopped Sync Loop.")
			return
		case <-ticker.C:
			peerID, err := ss.netService.node.StreamManager().RandomPeer()
			if err != nil {
				continue
			}
			ss.frontiers, err = getLocalFrontier(ss.qlcLedger)
			if err != nil {
				continue
			}
			ss.logger.Infof("begin sync block from [%s]", peerID)
			//ss.logger.Info("begin print fr info")
			//for k, v := range ss.frontiers {
			//	ss.logger.Info(k, v)
			//}
			ss.next()
			bulkPull = bulkPull[:0:0]
			bulkPush = bulkPush[:0:0]
			err = ss.netService.node.SendMessageToPeer(FrontierRequest, Req, peerID)
			if err != nil {
				ss.logger.Errorf("err [%s] when send FrontierRequest", err)
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Stop sync service
func (ss *ServiceSync) Stop() {
	//ss.logger.Info("Stop Qlc sync...")

	ss.quitCh <- true
}

func (ss *ServiceSync) onFrontierReq(message Message) error {
	ss.netService.node.logger.Info("receive FrontierReq")
	var fs []*types.Frontier
	fs, err := ss.qlcLedger.GetFrontiers()
	if err != nil {
		return err
	}

	for _, f := range fs {
		err = ss.netService.SendMessageToPeer(FrontierRsp, f, message.MessageFrom())
		if err != nil {
			ss.logger.Errorf("send FrontierRsp err [%s]", err)
		}
	}
	//send frontier finished,last frontier is all zero,tell remote peer send finished
	zeroFrontier := new(types.Frontier)
	err = ss.netService.SendMessageToPeer(FrontierRsp, zeroFrontier, message.MessageFrom())
	if err != nil {
		ss.logger.Errorf("send FrontierRsp err [%s] for zeroFrontier", err)
	}
	return nil
}

func (ss *ServiceSync) onFrontierRsp(message Message) error {
	ss.netService.node.logger.Info("receive FrontierRsp")
	fsRemote, err := protos.FrontierResponseFromProto(message.Data())
	if err != nil {
		return err
	}
	fr := fsRemote.Frontier
	//ss.logger.Info(fr.HeaderBlock, fr.OpenBlock)

	if !fr.OpenBlock.IsZero() {
		for {
			if !openBlockHash.IsZero() && (openBlockHash.String() < fr.OpenBlock.String()) {
				// We have an account but remote peer have not.
				push := &protos.Bulk{
					StartHash: zeroHash,
					EndHash:   headerBlockHash,
				}
				bulkPush = append(bulkPush, push)
				ss.next()
			} else {
				break
			}
		}
		if !openBlockHash.IsZero() {
			if fr.OpenBlock == openBlockHash {
				if headerBlockHash == fr.HeaderBlock {
					//ss.logger.Infof("this token %s have the same block", openBlockHash)
				} else {
					exit, _ := ss.qlcLedger.HasStateBlock(fr.HeaderBlock)
					if exit == true {
						push := &protos.Bulk{
							StartHash: fr.HeaderBlock,
							EndHash:   headerBlockHash,
						}
						bulkPush = append(bulkPush, push)
					} else {
						pull := &protos.Bulk{
							StartHash: headerBlockHash,
							EndHash:   fr.HeaderBlock,
						}
						bulkPull = append(bulkPull, pull)
					}
				}
				ss.next()
			} else {
				if fr.OpenBlock.String() > openBlockHash.String() {
					return nil
				}
				pull := &protos.Bulk{
					StartHash: zeroHash,
					EndHash:   fr.HeaderBlock,
				}
				bulkPull = append(bulkPull, pull)
			}
		} else {
			pull := &protos.Bulk{
				StartHash: zeroHash,
				EndHash:   fr.HeaderBlock,
			}
			bulkPull = append(bulkPull, pull)
		}
	} else {
		for {
			if !openBlockHash.IsZero() {
				// We have an account but remote peer have not.
				push := &protos.Bulk{
					StartHash: zeroHash,
					EndHash:   headerBlockHash,
				}
				bulkPush = append(bulkPush, push)
				ss.next()
			} else {
				if len(ss.frontiers) == 0 {
					getLocalFrontier(ss.qlcLedger)
					ss.next()
				}
				for _, value := range bulkPull {
					blkReq := &protos.BulkPullReqPacket{
						StartHash: value.StartHash,
						EndHash:   value.EndHash,
					}
					err = ss.netService.SendMessageToPeer(BulkPullRequest, blkReq, message.MessageFrom())
					if err != nil {
						ss.logger.Errorf("err [%s] when send BulkPullRequest", err)
					}
				}
				for _, value := range bulkPush {
					startHash := value.StartHash
					endHash := value.EndHash
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
							err = ss.netService.SendMessageToPeer(BulkPushBlock, bulkBlk[i], message.MessageFrom())
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
							err = ss.netService.SendMessageToPeer(BulkPushBlock, bulkBlk[i], message.MessageFrom())
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

func (ss *ServiceSync) onBulkPullRequest(message Message) error {
	ss.netService.node.logger.Info("receive BulkPullRequest")
	pullRemote, err := protos.BulkPullReqPacketFromProto(message.Data())
	if err != nil {
		return err
	}

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
			endHash = blk.GetPrevious()
			if endHash == startHash {
				break
			}
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

func (ss *ServiceSync) onBulkPullRsp(message Message) error {
	ss.netService.node.logger.Info("receive BulkPullRsp")
	blkPacket, err := protos.BulkPushBlockFromProto(message.Data())
	if err != nil {
		return err
	}

	block := blkPacket.Blk
	if ss.netService.node.cfg.PerformanceTest.Enabled {
		hash := block.GetHash()
		if exit, err := ss.qlcLedger.IsPerformanceTimeExist(hash); !exit && err == nil {
			if b, err := ss.qlcLedger.HasStateBlock(hash); !b && err == nil {
				t := &types.PerformanceTime{
					Hash: hash,
					T0:   time.Now().UnixNano(),
					T1:   0,
					T2:   0,
					T3:   0,
				}
				err = ss.qlcLedger.AddOrUpdatePerformance(t)
				if err != nil {
					ss.netService.node.logger.Error("error when run AddOrUpdatePerformance in onBulkPullRsp func")
				}
			}
		}
	}
	ss.netService.msgEvent.GetEvent("consensus").Notify(EventSyncBlock, block)
	return nil
}

func (ss *ServiceSync) onBulkPushBlock(message Message) error {
	ss.netService.node.logger.Info("receive BulkPushBlock")
	blkPacket, err := protos.BulkPushBlockFromProto(message.Data())
	if err != nil {
		return err
	}
	block := blkPacket.Blk

	if ss.netService.node.cfg.PerformanceTest.Enabled {
		hash := block.GetHash()
		if exit, err := ss.qlcLedger.IsPerformanceTimeExist(hash); !exit && err == nil {
			if b, err := ss.qlcLedger.HasStateBlock(hash); !b && err == nil {
				t := &types.PerformanceTime{
					Hash: hash,
					T0:   time.Now().UnixNano(),
					T1:   0,
					T2:   0,
					T3:   0,
				}
				err = ss.qlcLedger.AddOrUpdatePerformance(t)
				if err != nil {
					ss.netService.node.logger.Error("error when run AddOrUpdatePerformance in onBulkPushBlock func")
				}
			}
		}
	}
	ss.netService.msgEvent.GetEvent("consensus").Notify(EventSyncBlock, block)
	return nil
}

func (ss *ServiceSync) next() {
	if len(ss.frontiers) > 0 {
		openBlockHash = ss.frontiers[0].OpenBlock
		headerBlockHash = ss.frontiers[0].HeaderBlock
		ss.frontiers = ss.frontiers[1:]
	}
}
