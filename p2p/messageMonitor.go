package p2p

import (
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

//  Message Type
const (
	PublishReq      = "0" //PublishReq
	ConfirmReq      = "1" //ConfirmReq
	ConfirmAck      = "2" //ConfirmAck
	FrontierRequest = "3" //frontierreq
	FrontierRsp     = "4" //frontierrsp
	BulkPullRequest = "5" //bulkpull
	BulkPullRsp     = "6" //bulkpullrsp
	BulkPushBlock   = "7" //bulkpushblock
)

type MessageService struct {
	netService  *QlcService
	quitCh      chan bool
	messageCh   chan Message
	ledger      *ledger.Ledger
	syncService *ServiceSync
}

// NewService return new Service.
func NewMessageService(netService *QlcService, ledger *ledger.Ledger) *MessageService {
	ms := &MessageService{
		quitCh:     make(chan bool, 1),
		messageCh:  make(chan Message, 128),
		ledger:     ledger,
		netService: netService,
	}
	ms.syncService = NewSyncService(netService, ledger)
	return ms
}

// Start start message service.
func (ms *MessageService) Start() {
	// register the network handler.
	netService := ms.netService
	netService.Register(NewSubscriber(ms, ms.messageCh, false, PublishReq))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, ConfirmReq))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, ConfirmAck))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, FrontierRequest))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, FrontierRsp))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, BulkPullRequest))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, BulkPullRsp))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, BulkPushBlock))
	// start loop().
	go ms.startLoop()
	go ms.syncService.Start()
}
func (ms *MessageService) startLoop() {
	logger.Info("Started Message Service.")

	for {
		select {
		case <-ms.quitCh:
			logger.Info("Stopped Message Service.")
			return
		case message := <-ms.messageCh:
			switch message.MessageType() {
			case PublishReq:
				logger.Info("receive PublishReq")
				ms.onPublishReq(message)
			case ConfirmReq:
				logger.Info("receive ConfirmReq")
				ms.onConfirmReq(message)
			case ConfirmAck:
				logger.Info("receive ConfirmAck")
				ms.onConfirmAck(message)
			case FrontierRequest:
				logger.Info("receive FrontierReq")
				ms.syncService.onFrontierReq(message)
			case FrontierRsp:
				logger.Info("receive FrontierRsp")
				ms.syncService.onFrontierRsp(message)
			case BulkPullRequest:
				logger.Info("receive BulkPullRequest")
				ms.syncService.onBulkPullRequest(message)
			case BulkPullRsp:
				logger.Info("receive BulkPullRsp")
				ms.syncService.onBulkPullRsp(message)
			case BulkPushBlock:
				logger.Info("receive BulkPushBlock")
				ms.syncService.onBulkPushBlock(message)
			default:
				logger.Error("Received unknown message.")
			}
		}
	}
}
func (ms *MessageService) onPublishReq(message Message) error {
	blk, err := protos.PublishBlockFromProto(message.Data())
	if err != nil {
		return err
	}
	ms.netService.msgEvent.GetEvent("consensus").Notify(EventPublish, blk)
	return nil
}
func (ms *MessageService) onConfirmReq(message Message) error {
	blk, err := protos.ConfirmReqBlockFromProto(message.Data())
	if err != nil {
		return err
	}
	ms.netService.msgEvent.GetEvent("consensus").Notify(EventConfirmReq, blk)
	return nil
}
func (ms *MessageService) onConfirmAck(message Message) error {
	ack, err := protos.ConfirmAckBlockFromProto(message.Data())
	if err != nil {
		return err
	}
	ms.netService.msgEvent.GetEvent("consensus").Notify(EventConfirmAck, ack)
	return nil
}
func (ms *MessageService) Stop() {
	logger.Info("stopped message monitor")
	// quit.
	ms.quitCh <- true
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, PublishReq))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, ConfirmReq))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, ConfirmAck))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, FrontierRequest))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, FrontierRsp))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, BulkPullRequest))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, BulkPullRsp))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, BulkPushBlock))
}
