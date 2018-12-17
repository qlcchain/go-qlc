package p2p

import (
	"math"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	messagepb "github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	SyncInterval = time.Minute * 5
)

var zeroHash types.Hash

var headerBlockHash types.Hash
var openBlockHash types.Hash
var bulkPush, bulkPull []*messagepb.Bulk

// Service manage sync tasks
type ServiceSync struct {
	netService *QlcService
	qlcLedger  *ledger.Ledger
	frontiers  []*types.Frontier
	quitCh     chan bool
}

// NewService return new Service.
func NewSyncService(netService *QlcService, ledger *ledger.Ledger) *ServiceSync {
	frontiers, err := getLocalFrontier(ledger)
	if err != nil {
		logger.Error("New Sync Service error")
	}
	ss := &ServiceSync{
		netService: netService,
		qlcLedger:  ledger,
		frontiers:  frontiers,
		quitCh:     make(chan bool, 1),
	}
	ss.next()
	return ss
}
func (ss *ServiceSync) Start() {
	logger.Info("started sync loop")
	address := types.Address{}
	Req := messagepb.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	data, err := messagepb.FrontierReqToProto(Req)
	if err != nil {
		logger.Error("New FrontierReq error")
		return
	}
	ticker := time.NewTicker(SyncInterval)
	for {
		peerID, err := ss.netService.node.StreamManager().RandomPeer()
		if err != nil {
			continue
		}
		select {
		case <-ss.quitCh:
			logger.Info("Stopped Sync Loop.")
			return
		case <-ticker.C:
			ss.netService.node.SendMessageToPeer(FrontierRequest, data, peerID)
		}
	}
}

// Stop sync service
func (ss *ServiceSync) Stop() {
	logger.Info("Stop Qlc sync...")

	ss.quitCh <- true
}
func (ss *ServiceSync) onFrontierReq(message Message) error {
	var fs []*types.Frontier
	session := ss.qlcLedger.NewLedgerSession(false)
	defer session.Close()
	fs, err := session.GetFrontiers()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	for _, f := range fs {
		qlcfrs := messagepb.NewFrontierRsp(f)
		frsbytes, err := messagepb.FrontierResponseToProto(qlcfrs)
		if err != nil {
			return err
		}
		ss.netService.SendMessageToPeer(FrontierRsp, frsbytes, message.MessageFrom())
	}
	//send frontier finished,last frontier is all zero,tell remote peer send finished
	rsp := new(types.Frontier)
	frsp := messagepb.NewFrontierRsp(rsp)
	bytes, err := messagepb.FrontierResponseToProto(frsp)
	if err != nil {
		return err
	}
	ss.netService.SendMessageToPeer(FrontierRsp, bytes, message.MessageFrom())
	return nil
}
func (ss *ServiceSync) onFrontierRsp(message Message) error {
	fsremote, err := messagepb.FrontierResponseFromProto(message.Data())
	if err != nil {
		return err
	}
	fr := fsremote.Frontier
	logger.Info(fr.HeaderBlock, fr.OpenBlock)
	session := ss.qlcLedger.NewLedgerSession(false)
	defer session.Close()

	if !fr.OpenBlock.IsZero() {
		for {
			if !openBlockHash.IsZero() && (openBlockHash.String() < fr.OpenBlock.String()) {
				// We have an account but remote peer have not.
				push := &messagepb.Bulk{
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
					logger.Infof("this token %s have the same block", openBlockHash)
				} else {
					exit, _ := session.HasBlock(fr.HeaderBlock)
					if exit == true {
						push := &messagepb.Bulk{
							StartHash: fr.HeaderBlock,
							EndHash:   headerBlockHash,
						}
						bulkPush = append(bulkPush, push)
					} else {
						pull := &messagepb.Bulk{
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
				pull := &messagepb.Bulk{
					StartHash: zeroHash,
					EndHash:   fr.HeaderBlock,
				}
				bulkPull = append(bulkPull, pull)
			}
		} else {
			pull := &messagepb.Bulk{
				StartHash: zeroHash,
				EndHash:   fr.HeaderBlock,
			}
			bulkPull = append(bulkPull, pull)
		}
	} else {
		for {
			if !openBlockHash.IsZero() {
				// We have an account but remote peer have not.
				push := &messagepb.Bulk{
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
					blkreqpk := &messagepb.BulkPullReqPacket{
						StartHash: value.StartHash,
						EndHash:   value.EndHash,
					}
					bytes, err := messagepb.BulkPullReqPacketToProto(blkreqpk)
					if err != nil {
						return err
					}
					ss.netService.SendMessageToPeer(BulkPullRequest, bytes, message.MessageFrom())
				}
				for _, value := range bulkPush {
					startHash := value.StartHash
					endHash := value.EndHash
					if startHash.IsZero() {
						logger.Infof("need to send all the blocks of this account")
						var blk types.Block
						var bulkblk []types.Block
						for {
							blk, err = session.GetBlock(endHash)
							if err != nil {
								return err
							}
							bulkblk = append(bulkblk, blk)
							endHash = blk.GetPrevious()
							if endHash.IsZero() == true {
								break
							}
						}
						for i := (len(bulkblk) - 1); i >= 0; i-- {
							push := &messagepb.BulkPush{
								Blk: bulkblk[i],
							}
							blockBytes, err := messagepb.BulkPushBlockToProto(push)
							if err != nil {
								return err
							}
							ss.netService.SendMessageToPeer(BulkPushBlock, blockBytes, message.MessageFrom())
						}
					} else {
						logger.Info("need to send some blocks of this account")
						var blk types.Block
						var bulkblk []types.Block
						for {
							blk, err = session.GetBlock(endHash)
							if err != nil {
								return err
							}
							bulkblk = append(bulkblk, blk)

							endHash = blk.GetPrevious()
							if endHash == startHash {
								break
							}
						}
						for i := (len(bulkblk) - 1); i >= 0; i-- {
							push := &messagepb.BulkPush{
								Blk: bulkblk[i],
							}
							blockBytes, err := messagepb.BulkPushBlockToProto(push)
							if err != nil {
								return err
							}
							ss.netService.SendMessageToPeer(BulkPushBlock, blockBytes, message.MessageFrom())
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
	session := ledger.NewLedgerSession(false)
	defer session.Close()
	frontiers, err := session.GetFrontiers()
	if err != nil {
		return nil, err
	}
	fsback := new(types.Frontier)
	frontiers = append(frontiers, fsback)
	return frontiers, nil
}
func (ss *ServiceSync) onBulkPullRequest(message Message) error {
	pullremote, err := messagepb.BulkPullReqPacketFromProto(message.Data())
	if err != nil {
		return err
	}
	session := ss.qlcLedger.NewLedgerSession(false)
	defer session.Close()
	startHash := pullremote.StartHash
	endHash := pullremote.EndHash
	if startHash.IsZero() {
		var blk types.Block
		var bulkblk []types.Block
		logger.Info("need to send all the blocks of this account")
		for {
			blk, err = session.GetBlock(endHash)
			if err != nil {
				return err
			}
			bulkblk = append(bulkblk, blk)
			endHash = blk.GetPrevious()
			if endHash.IsZero() == true {
				break
			}
		}
		for i := (len(bulkblk) - 1); i >= 0; i-- {
			PullRsp := &messagepb.BulkPullRspPacket{
				Blk: bulkblk[i],
			}
			blockBytes, err := messagepb.BulkPullRspPacketToProto(PullRsp)
			if err != nil {
				return err
			}
			ss.netService.SendMessageToPeer(BulkPullRsp, blockBytes, message.MessageFrom())
		}
	} else {
		var blk types.Block
		var bulkblk []types.Block
		logger.Info("need to send some blocks of this account")
		for {
			blk, err = session.GetBlock(endHash)
			if err != nil {
				return err
			}
			bulkblk = append(bulkblk, blk)
			endHash = blk.GetPrevious()
			if endHash == startHash {
				break
			}
		}
		for i := (len(bulkblk) - 1); i >= 0; i-- {
			PullRsp := &messagepb.BulkPullRspPacket{
				Blk: bulkblk[i],
			}
			blockBytes, err := messagepb.BulkPullRspPacketToProto(PullRsp)
			if err != nil {
				return err
			}
			ss.netService.SendMessageToPeer(BulkPullRsp, blockBytes, message.MessageFrom())
		}
	}
	return nil
}
func (ss *ServiceSync) onBulkPullRsp(message Message) error {
	blkpacket, err := messagepb.BulkPushBlockFromProto(message.Data())
	if err != nil {
		return err
	}
	session := ss.qlcLedger.NewLedgerSession(false)
	defer session.Close()

	block := blkpacket.Blk
	ss.netService.msgEvent.GetEvent("consensus").Notify(EventSyncBlock, block)
	err = session.AddBlock(block)
	if err != nil {
		return err
	}
	previousHash := block.GetPrevious()
	if previousHash.IsZero() == false {
		currentfr, err := session.GetFrontier(block.GetPrevious())
		if err != nil {
			return err
		}
		updatefr := &types.Frontier{
			HeaderBlock: block.GetHash(),
			OpenBlock:   currentfr.OpenBlock,
		}
		err = session.DeleteFrontier(block.GetPrevious())
		if err != nil {
			return err
		}
		err = session.AddFrontier(updatefr)
		if err != nil {
			return err
		}
	} else {
		fr := &types.Frontier{
			HeaderBlock: block.GetHash(),
			OpenBlock:   block.GetHash(),
		}
		err = session.AddFrontier(fr)
		if err != nil {
			return err
		}
	}
	return nil
}
func (ss *ServiceSync) onBulkPushBlock(message Message) error {
	blkpacket, err := messagepb.BulkPushBlockFromProto(message.Data())
	if err != nil {
		return err
	}
	block := blkpacket.Blk
	ss.netService.msgEvent.GetEvent("consensus").Notify(EventSyncBlock, block)
	session := ss.qlcLedger.NewLedgerSession(false)
	defer session.Close()

	err = session.AddBlock(block)
	if err != nil {
		return err
	}
	previousHash := block.GetPrevious()
	if previousHash.IsZero() == false {
		currentfr, err := session.GetFrontier(block.GetPrevious())
		if err != nil {
			return err
		}
		updatefr := &types.Frontier{
			HeaderBlock: block.GetHash(),
			OpenBlock:   currentfr.OpenBlock,
		}
		err = session.DeleteFrontier(block.GetPrevious())
		if err != nil {
			return err
		}
		err = session.AddFrontier(updatefr)
		if err != nil {
			return err
		}
	} else {
		fr := &types.Frontier{
			HeaderBlock: block.GetHash(),
			OpenBlock:   block.GetHash(),
		}
		err = session.AddFrontier(fr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ss *ServiceSync) next() {
	if len(ss.frontiers) > 0 {
		openBlockHash = ss.frontiers[0].OpenBlock
		headerBlockHash = ss.frontiers[0].HeaderBlock
		ss.frontiers = ss.frontiers[1:]
	}
}
