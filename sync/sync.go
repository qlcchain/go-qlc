package sync

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
)

var logger = log.NewLogger("sync")
var zeroHash = types.Hash{}

//  sync Message Type
const (
	FrontierRequest = "frontierreq" //frontierreq
	FrontierRsp     = "frontierrsp" //frontierrsp
	BulkPullRequest = "bulkpull"    //bulkpull
	BulkPullRsp     = "bulkpullrsp" //bulkpullrsp
	BulkPushBlock   = "bulkpush"    //bulkpushblock
)

var frontiers []*types.Frontier
var headerBlockHash types.Hash
var openBlockHash types.Hash
var bulkPush, bulkPull []*Bulk

// Service manage sync tasks
type ServiceSync struct {
	netService p2p.Service
	quitCh     chan bool
	messageCh  chan p2p.Message
	qlcLedger  *ledger.Ledger
}

// NewService return new Service.
func NewSyncService(netService p2p.Service) *ServiceSync {
	return &ServiceSync{
		quitCh:     make(chan bool, 1),
		messageCh:  make(chan p2p.Message, 128),
		netService: netService,
	}
}

// SetQlcService set ledger
func (ss *ServiceSync) SetLedger(ledger *ledger.Ledger) {
	ss.qlcLedger = ledger
}

// Start start sync service.
func (ss *ServiceSync) Start() {
	logger.Info("Started sync Service.")

	if len(frontiers) == 0 {
		ss.getLocalFrontier()
		next()
	}
	// register the network handler.
	netService := ss.netService
	netService.Register(p2p.NewSubscriber(ss, ss.messageCh, false, FrontierRequest))
	netService.Register(p2p.NewSubscriber(ss, ss.messageCh, false, FrontierRsp))
	netService.Register(p2p.NewSubscriber(ss, ss.messageCh, false, BulkPullRequest))
	netService.Register(p2p.NewSubscriber(ss, ss.messageCh, false, BulkPullRsp))
	netService.Register(p2p.NewSubscriber(ss, ss.messageCh, false, BulkPushBlock))
	// start loop().
	go ss.startLoop()
}
func (ss *ServiceSync) startLoop() {
	for {
		select {
		case <-ss.quitCh:
			logger.Info("Stopped sync Service.")
			return
		case message := <-ss.messageCh:
			switch message.MessageType() {
			case FrontierRequest:
				logger.Info("receive FrontierReq")
				ss.onFrontierReq(message)
			case FrontierRsp:
				logger.Info("receive FrontierRsp")
				ss.onFrontierRsp(message)
			case BulkPullRequest:
				logger.Info("receive BulkPullRequest")
				ss.onBulkPullRequest(message)
			case BulkPullRsp:
				logger.Info("receive BulkPullRsp")
				ss.onBulkPullRsp(message)
			case BulkPushBlock:
				logger.Info("receive BulkPushBlock")
				ss.onBulkPushBlock(message)
			default:
				logger.Error("Received unknown message.")
			}
		}
	}
}
func (ss *ServiceSync) onFrontierReq(message p2p.Message) error {
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
		qlcfrs := NewFrontierRsp(f)
		frsbytes, err := FrontierResponseToProto(qlcfrs)
		if err != nil {
			return err
		}
		ss.netService.SendMessageToPeer(FrontierRsp, frsbytes, message.MessageFrom())
	}
	//send frontier finished,last frontier is all zero,tell remote peer send finished
	rsp := new(types.Frontier)
	frsp := NewFrontierRsp(rsp)
	bytes, err := FrontierResponseToProto(frsp)
	if err != nil {
		return err
	}
	ss.netService.SendMessageToPeer(FrontierRsp, bytes, message.MessageFrom())
	return nil
}
func (ss *ServiceSync) onFrontierRsp(message p2p.Message) error {
	fsremote, err := FrontierResponseFromProto(message.Data())
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
				push := &Bulk{
					StartHash: zeroHash,
					EndHash:   headerBlockHash,
				}
				bulkPush = append(bulkPush, push)
				next()
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
						push := &Bulk{
							StartHash: fr.HeaderBlock,
							EndHash:   headerBlockHash,
						}
						bulkPush = append(bulkPush, push)
					} else {
						pull := &Bulk{
							StartHash: headerBlockHash,
							EndHash:   fr.HeaderBlock,
						}
						bulkPull = append(bulkPull, pull)
					}
				}
				next()
			} else {
				if fr.OpenBlock.String() > openBlockHash.String() {
					return nil
				}
				pull := &Bulk{
					StartHash: zeroHash,
					EndHash:   fr.HeaderBlock,
				}
				bulkPull = append(bulkPull, pull)
			}
		} else {
			pull := &Bulk{
				StartHash: zeroHash,
				EndHash:   fr.HeaderBlock,
			}
			bulkPull = append(bulkPull, pull)
		}
	} else {
		for {
			if !openBlockHash.IsZero() {
				// We have an account but remote peer have not.
				push := &Bulk{
					StartHash: zeroHash,
					EndHash:   headerBlockHash,
				}
				bulkPush = append(bulkPush, push)
				next()
			} else {
				if len(frontiers) == 0 {
					ss.getLocalFrontier()
					next()
				}
				for _, value := range bulkPull {
					blkreqpk := &BulkPullReqPacket{
						StartHash: value.StartHash,
						EndHash:   value.EndHash,
					}
					bytes, err := BulkPullReqPacketToProto(blkreqpk)
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
							push := &BulkPush{
								blk: bulkblk[i],
							}
							blockBytes, err := BulkPushBlockToProto(push)
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
							push := &BulkPush{
								blk: bulkblk[i],
							}
							blockBytes, err := BulkPushBlockToProto(push)
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

func (ss *ServiceSync) getLocalFrontier() (err error) {
	session := ss.qlcLedger.NewLedgerSession(false)
	defer session.Close()
	frontiers, err = session.GetFrontiers()
	if err != nil {
		return err
	}
	fsback := new(types.Frontier)
	frontiers = append(frontiers, fsback)
	return nil
}
func (ss *ServiceSync) onBulkPullRequest(message p2p.Message) error {
	pullremote, err := BulkPullReqPacketFromProto(message.Data())
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
			PullRsp := &BulkPullRspPacket{
				blk: bulkblk[i],
			}
			blockBytes, err := BulkPullRspPacketToProto(PullRsp)
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
			PullRsp := &BulkPullRspPacket{
				blk: bulkblk[i],
			}
			blockBytes, err := BulkPullRspPacketToProto(PullRsp)
			if err != nil {
				return err
			}
			ss.netService.SendMessageToPeer(BulkPullRsp, blockBytes, message.MessageFrom())
		}
	}
	return nil
}
func (ss *ServiceSync) onBulkPullRsp(message p2p.Message) error {
	blkpacket, err := BulkPushBlockFromProto(message.Data())
	if err != nil {
		return err
	}
	session := ss.qlcLedger.NewLedgerSession(false)
	defer session.Close()

	block := blkpacket.blk
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
func (ss *ServiceSync) onBulkPushBlock(message p2p.Message) error {
	blkpacket, err := BulkPushBlockFromProto(message.Data())
	if err != nil {
		return err
	}
	block := blkpacket.blk
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
func (ss *ServiceSync) Stop() {
	logger.Info("stopped sync service")
	// quit.
	ss.quitCh <- true
	ss.netService.Deregister(p2p.NewSubscriber(ss, ss.messageCh, false, FrontierRequest))
	ss.netService.Deregister(p2p.NewSubscriber(ss, ss.messageCh, false, FrontierRsp))
	ss.netService.Deregister(p2p.NewSubscriber(ss, ss.messageCh, false, BulkPullRequest))
	ss.netService.Deregister(p2p.NewSubscriber(ss, ss.messageCh, false, BulkPullRsp))
	ss.netService.Deregister(p2p.NewSubscriber(ss, ss.messageCh, false, BulkPushBlock))
}
func next() {
	if len(frontiers) > 0 {
		openBlockHash = frontiers[0].OpenBlock
		headerBlockHash = frontiers[0].HeaderBlock
		frontiers = frontiers[1:]
	}
}
