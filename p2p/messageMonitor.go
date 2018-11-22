package p2p

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
)

//  Message Type
const (
	PublishRequest = "publish" // publish
)
const BlockEndIdx = 1

var log = common.NewLogger("message")

type MessageService struct {
	netService *QlcService
	quitCh     chan bool
	messageCh  chan Message
}

// NewService return new Service.
func NewMessageService() *MessageService {
	return &MessageService{
		quitCh:    make(chan bool, 1),
		messageCh: make(chan Message, 128),
	}
}

// SetQlcService set netService
func (ss *MessageService) SetQlcService(ns *QlcService) {
	ss.netService = ns
}

// Start start sync service.
func (ss *MessageService) Start() {
	log.Info("Starting Message Service.")
	// register the network handler.
	netService := ss.netService
	netService.Register(NewSubscriber(ss, ss.messageCh, false, PublishRequest))
	// start loop().
	go ss.startLoop()
}
func (ss *MessageService) startLoop() {
	log.Info("Started Message Service.")

	for {
		select {
		case <-ss.quitCh:
			log.Info("Stopped Message Service.")
			return
		case message := <-ss.messageCh:
			switch message.MessageType() {
			case PublishRequest:
				log.Info("receive PublishRequest")
				ss.onPublishRequest(message)
			default:
				log.Error("Received unknown message.")
			}
		}
	}
}
func (ss *MessageService) onPublishRequest(message Message) error {
	BlockType := message.Data()[0]
	Block, err := types.NewBlock(BlockType)
	if err != nil {
		return err
	}
	blockdata := message.Data()[BlockEndIdx:]
	if _, err = Block.UnmarshalMsg(blockdata); err != nil {
		return err
	}
	ss.processBlock(Block, message.MessageType())
	return nil
}
func (ss *MessageService) processBlock(block types.Block, msgtype MessageType) {
	e := &Event{
		MsgType: msgtype,
		Blocks:  block,
	}
	ss.netService.MessageEvent().Notify(e)
}
func (ss *MessageService) Stop() {
	log.Info("stopped message monitor")
	// quit.
	ss.quitCh <- true
}
