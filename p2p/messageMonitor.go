package p2p

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	checkBlockCacheInterval = 60 * time.Second
	maxPullTxPerReq         = 100
	maxPushTxPerTime        = 1000
)

type peerPullRsp struct {
	pullRspTimer   *time.Timer
	pullRspHash    types.Hash
	pullRspStartCh chan bool
}

//  Message Type
const (
	PublishReq      MessageType = iota //PublishReq
	ConfirmReq                         //ConfirmReq
	ConfirmAck                         //ConfirmAck
	FrontierRequest                    //FrontierReq
	FrontierRsp                        //FrontierRsp
	BulkPullRequest                    //BulkPullRequest
	BulkPullRsp                        //BulkPullRsp
	BulkPushBlock                      //BulkPushBlock
	MessageResponse                    //MessageResponse
	PovStatus
	PovPublishReq
	PovBulkPullReq
	PovBulkPullRsp
)

type MessageService struct {
	netService          *QlcService
	ctx                 context.Context
	cancel              context.CancelFunc
	messageCh           chan *Message
	publishMessageCh    chan *Message
	confirmReqMessageCh chan *Message
	confirmAckMessageCh chan *Message
	rspMessageCh        chan *Message
	povMessageCh        chan *Message
	ledger              *ledger.Ledger
	syncService         *ServiceSync
	pullRspMap          *sync.Map
}

// NewService return new Service.
func NewMessageService(netService *QlcService, ledger *ledger.Ledger) *MessageService {
	ctx, cancel := context.WithCancel(context.Background())
	ms := &MessageService{
		ctx:                 ctx,
		cancel:              cancel,
		messageCh:           make(chan *Message, common.P2PMonitorMsgChanSize),
		publishMessageCh:    make(chan *Message, common.P2PMonitorMsgChanSize),
		confirmReqMessageCh: make(chan *Message, common.P2PMonitorMsgChanSize),
		confirmAckMessageCh: make(chan *Message, common.P2PMonitorMsgChanSize),
		rspMessageCh:        make(chan *Message, common.P2PMonitorMsgChanSize),
		povMessageCh:        make(chan *Message, common.P2PMonitorMsgChanSize),
		ledger:              ledger,
		netService:          netService,
		pullRspMap:          new(sync.Map),
	}
	ms.syncService = NewSyncService(netService, ledger)
	return ms
}

// Start start message service.
func (ms *MessageService) Start() {
	// register the network handler.
	netService := ms.netService
	netService.Register(NewSubscriber(ms.publishMessageCh, PublishReq))
	netService.Register(NewSubscriber(ms.confirmReqMessageCh, ConfirmReq))
	netService.Register(NewSubscriber(ms.confirmAckMessageCh, ConfirmAck))
	netService.Register(NewSubscriber(ms.messageCh, FrontierRequest))
	netService.Register(NewSubscriber(ms.messageCh, FrontierRsp))
	netService.Register(NewSubscriber(ms.messageCh, BulkPullRequest))
	netService.Register(NewSubscriber(ms.messageCh, BulkPullRsp))
	netService.Register(NewSubscriber(ms.messageCh, BulkPushBlock))
	netService.Register(NewSubscriber(ms.rspMessageCh, MessageResponse))
	// PoV message handlers
	netService.Register(NewSubscriber(ms.povMessageCh, PovStatus))
	netService.Register(NewSubscriber(ms.povMessageCh, PovPublishReq))
	netService.Register(NewSubscriber(ms.povMessageCh, PovBulkPullReq))
	netService.Register(NewSubscriber(ms.povMessageCh, PovBulkPullRsp))
	// start loop().
	go ms.startLoop()
	go ms.syncService.Start()
	go ms.publishReqLoop()
	go ms.confirmReqLoop()
	go ms.confirmAckLoop()
	go ms.povMessageLoop()
	go ms.processBlockCacheLoop()
	go ms.messageResponseLoop()
}

func (ms *MessageService) processBlockCacheLoop() {
	ms.netService.node.logger.Info("Started process blockCache loop.")
	ticker := time.NewTicker(checkBlockCacheInterval)
	for {
		select {
		case <-ms.ctx.Done():
			return
		case <-ticker.C:
			ms.processBlockCache()
		}
	}
}

func (ms *MessageService) processBlockCache() {
	blocks := make([]*types.StateBlock, 0)
	err := ms.ledger.GetBlockCaches(func(block *types.StateBlock) error {
		blocks = append(blocks, block)
		return nil
	})
	if err != nil {
		ms.netService.node.logger.Error("get block cache error")
	}
	for _, blk := range blocks {
		if b, err := ms.ledger.HasStateBlockConfirmed(blk.GetHash()); b && err == nil {
			_ = ms.ledger.DeleteBlockCache(blk.GetHash())
		} else {
			b, _ := ms.ledger.HasStateBlock(blk.GetHash())
			if b {
				ms.netService.msgEvent.Publish(common.EventBroadcast, PublishReq, blk)
				ms.netService.msgEvent.Publish(common.EventGenerateBlock, process.Progress, blk)
			}
		}
	}
}

func (ms *MessageService) startLoop() {
	ms.netService.node.logger.Info("Started Message Service.")
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.messageCh:
			switch message.MessageType() {
			case FrontierRequest:
				go func() {
					if err := ms.syncService.onFrontierReq(message); err != nil {
						ms.netService.node.logger.Error(err)
					}
				}()
			case FrontierRsp:
				go ms.syncService.checkFrontier(message)
			case BulkPullRequest:
				go func() {
					if err := ms.syncService.onBulkPullRequest(message); err != nil {
						ms.netService.node.logger.Error(err)
					}
				}()
			case BulkPullRsp:
				go func() {
					if err := ms.syncService.onBulkPullRsp(message); err != nil {
						ms.netService.node.logger.Error(err)
					}
				}()
			case BulkPushBlock:
				go func() {
					if err := ms.syncService.onBulkPushBlock(message); err != nil {
						ms.netService.node.logger.Error(err)
					}
				}()
			default:
				ms.netService.node.logger.Error("Received unknown message.")
			}
		}
	}
}

func (ms *MessageService) messageResponseLoop() {
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.rspMessageCh:
			switch message.MessageType() {
			case MessageResponse:
				ms.onMessageResponse(message)
			}
		}
	}
}

func (ms *MessageService) publishReqLoop() {
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.publishMessageCh:
			switch message.MessageType() {
			case PublishReq:
				ms.onPublishReq(message)
			}
		}
	}
}

func (ms *MessageService) confirmReqLoop() {
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.confirmReqMessageCh:
			switch message.MessageType() {
			case ConfirmReq:
				ms.onConfirmReq(message)
			}
		}
	}
}

func (ms *MessageService) confirmAckLoop() {
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.confirmAckMessageCh:
			switch message.MessageType() {
			case ConfirmAck:
				ms.onConfirmAck(message)
			}
		}
	}
}

func (ms *MessageService) povMessageLoop() {
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.povMessageCh:
			switch message.MessageType() {
			case PovStatus:
				ms.onPovStatus(message)
			case PovPublishReq:
				ms.onPovPublishReq(message)
			case PovBulkPullReq:
				ms.onPovBulkPullReq(message)
			case PovBulkPullRsp:
				ms.onPovBulkPullRsp(message)
			default:
				ms.netService.node.logger.Warn("Received unknown pov message.")
			}
		}
	}
}

func (ms *MessageService) onMessageResponse(message *Message) {
	ma, err := protos.MessageAckFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}
	if v, ok := ms.pullRspMap.Load(message.from); ok {
		pr := v.(*peerPullRsp)
		if ma.MessageHash == pr.pullRspHash {
			select {
			case pr.pullRspStartCh <- true:
			default:
			}
		}
	}
}

func (ms *MessageService) onPublishReq(message *Message) {
	if ms.netService.node.cfg.PerformanceEnabled {
		blk, err := protos.PublishBlockFromProto(message.Data())
		if err != nil {
			ms.netService.node.logger.Error(err)
			return
		}
		hash := blk.Blk.GetHash()
		ms.addPerformanceTime(hash)
	}
	p, err := protos.PublishBlockFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}
	ms.netService.msgEvent.Publish(common.EventPublish, p.Blk, message.MessageFrom())
}

func (ms *MessageService) onConfirmReq(message *Message) {
	if ms.netService.node.cfg.PerformanceEnabled {
		blk, err := protos.ConfirmReqBlockFromProto(message.Data())
		if err != nil {
			ms.netService.node.logger.Error(err)
			return
		}

		for _, b := range blk.Blk {
			hash := b.GetHash()
			ms.addPerformanceTime(hash)
		}
	}
	r, err := protos.ConfirmReqBlockFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Error(err)
		return
	}
	ms.netService.msgEvent.Publish(common.EventConfirmReq, r.Blk, message.MessageFrom())
}

func (ms *MessageService) onConfirmAck(message *Message) {
	if ms.netService.node.cfg.PerformanceEnabled {
		ack, err := protos.ConfirmAckBlockFromProto(message.Data())
		if err != nil {
			ms.netService.node.logger.Error(err)
			return
		}

		for _, h := range ack.Hash {
			ms.addPerformanceTime(h)
		}
	}
	ack, err := protos.ConfirmAckBlockFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}
	ms.netService.msgEvent.Publish(common.EventConfirmAck, ack, message.MessageFrom())
}

func (ms *MessageService) onPovStatus(message *Message) {
	status, err := protos.PovStatusFromProto(message.data)
	if err != nil {
		ms.netService.node.logger.Errorf("failed to decode PovStatus from peer %s", message.from)
		return
	}
	ms.netService.msgEvent.Publish(common.EventPovPeerStatus, status, message.MessageFrom())
}

func (ms *MessageService) onPovPublishReq(message *Message) {

	p, err := protos.PovPublishBlockFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}

	ms.netService.msgEvent.Publish(common.EventPovRecvBlock, p.Blk, types.PovBlockFromRemoteBroadcast, message.MessageFrom())
}

func (ms *MessageService) onPovBulkPullReq(message *Message) {
	req, err := protos.PovBulkPullReqFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}

	ms.netService.msgEvent.Publish(common.EventPovBulkPullReq, req, message.MessageFrom())
}

func (ms *MessageService) onPovBulkPullRsp(message *Message) {
	rsp, err := protos.PovBulkPullRspFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}

	ms.netService.msgEvent.Publish(common.EventPovBulkPullRsp, rsp, message.MessageFrom())
}

func (ms *MessageService) Stop() {
	//ms.netService.node.logger.Info("stopped message monitor")
	// quit.
	ms.cancel()
	ms.syncService.quitCh <- true
	ms.netService.Deregister(NewSubscriber(ms.publishMessageCh, PublishReq))
	ms.netService.Deregister(NewSubscriber(ms.confirmReqMessageCh, ConfirmReq))
	ms.netService.Deregister(NewSubscriber(ms.confirmAckMessageCh, ConfirmAck))
	ms.netService.Deregister(NewSubscriber(ms.messageCh, FrontierRequest))
	ms.netService.Deregister(NewSubscriber(ms.messageCh, FrontierRsp))
	ms.netService.Deregister(NewSubscriber(ms.messageCh, BulkPullRequest))
	ms.netService.Deregister(NewSubscriber(ms.messageCh, BulkPullRsp))
	ms.netService.Deregister(NewSubscriber(ms.messageCh, BulkPushBlock))
	ms.netService.Deregister(NewSubscriber(ms.rspMessageCh, MessageResponse))
	ms.netService.Deregister(NewSubscriber(ms.povMessageCh, PovStatus))
	ms.netService.Deregister(NewSubscriber(ms.povMessageCh, PovPublishReq))
	ms.netService.Deregister(NewSubscriber(ms.povMessageCh, PovBulkPullReq))
	ms.netService.Deregister(NewSubscriber(ms.povMessageCh, PovBulkPullRsp))
}

func marshalMessage(messageName MessageType, value interface{}) ([]byte, error) {
	switch messageName {
	case PublishReq:
		packet := protos.PublishBlock{
			Blk: value.(*types.StateBlock),
		}
		data, err := protos.PublishBlockToProto(&packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case ConfirmReq:
		packet := &protos.ConfirmReqBlock{
			Blk: value.([]*types.StateBlock),
		}
		data, err := protos.ConfirmReqBlockToProto(packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case ConfirmAck:
		data, err := protos.ConfirmAckBlockToProto(value.(*protos.ConfirmAckBlock))
		if err != nil {
			return nil, err
		}
		return data, nil
	case FrontierRequest:
		data, err := protos.FrontierReqToProto(value.(*protos.FrontierReq))
		if err != nil {
			return nil, err
		}
		return data, nil
	case FrontierRsp:
		packet := value.(*protos.FrontierResponse)
		data, err := protos.FrontierResponseToProto(packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case BulkPullRequest:
		data, err := protos.BulkPullReqPacketToProto(value.(*protos.BulkPullReqPacket))
		if err != nil {
			return nil, err
		}
		return data, nil
	case BulkPullRsp:
		data, err := protos.BulkPullRspPacketToProto(value.(*protos.BulkPullRspPacket))
		if err != nil {
			return nil, err
		}
		return data, err
	case BulkPushBlock:
		push := &protos.BulkPush{
			Blocks: value.(types.StateBlockList),
		}
		data, err := protos.BulkPushBlockToProto(push)
		if err != nil {
			return nil, err
		}
		return data, nil
	case PovStatus:
		status := value.(*protos.PovStatus)
		data, err := protos.PovStatusToProto(status)
		if err != nil {
			return nil, err
		}
		return data, nil
	case PovPublishReq:
		packet := protos.PovPublishBlock{
			Blk: value.(*types.PovBlock),
		}
		data, err := protos.PovPublishBlockToProto(&packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case PovBulkPullReq:
		req := value.(*protos.PovBulkPullReq)
		data, err := protos.PovBulkPullReqToProto(req)
		if err != nil {
			return nil, err
		}
		return data, nil
	case PovBulkPullRsp:
		rsp := value.(*protos.PovBulkPullRsp)
		data, err := protos.PovBulkPullRspToProto(rsp)
		if err != nil {
			return nil, err
		}
		return data, nil
	case MessageResponse:
		rsp := &protos.MessageAckPacket{
			MessageHash: value.(types.Hash),
		}
		data, err := protos.MessageAckToProto(rsp)
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		return nil, errors.New("unKnown Message Type")
	}
}

func (ms *MessageService) addPerformanceTime(hash types.Hash) {
	if exit, err := ms.ledger.IsPerformanceTimeExist(hash); !exit && err == nil {
		if b, err := ms.ledger.HasStateBlock(hash); !b && err == nil {
			t := &types.PerformanceTime{
				Hash: hash,
				T0:   time.Now().UnixNano(),
				T1:   0,
				T2:   0,
				T3:   0,
			}
			err = ms.ledger.AddOrUpdatePerformance(t)
			if err != nil {
				ms.netService.node.logger.Error("error when run AddOrUpdatePerformance in onConfirmAck func")
			}
		}
	}
}
