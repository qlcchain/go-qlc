package p2p

import (
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

// QlcService service for qlc p2p network
type QlcService struct {
	node       *QlcNode
	dispatcher *Dispatcher
	msgEvent   event.EventBus
	msgService *MessageService
}

// NewQlcService create netService
func NewQlcService(cfg *config.Config, eb event.EventBus) (*QlcService, error) {
	node, err := NewNode(cfg)
	if err != nil {
		return nil, err
	}
	ns := &QlcService{
		node:       node,
		dispatcher: NewDispatcher(),
		msgEvent:   eb,
	}
	node.SetQlcService(ns)
	l := ledger.NewLedger(cfg.LedgerDir(), eb)
	msgService := NewMessageService(ns, l)
	ns.msgService = msgService
	return ns, nil
}

// Node return the peer node
func (ns *QlcService) Node() *QlcNode {
	return ns.node
}

// EventQueue return EventQueue
func (ns *QlcService) MessageEvent() event.EventBus {
	return ns.msgEvent
}

// Start start p2p manager.
func (ns *QlcService) Start() error {
	//ns.node.logger.Info("Starting QlcService...")

	// start dispatcher.
	ns.dispatcher.Start()

	//set event
	err := ns.setEvent()
	if err != nil {
		return err
	}
	// start node.
	if err := ns.node.StartServices(); err != nil {
		ns.dispatcher.Stop()
		ns.node.logger.Error("Failed to start QlcService.")
		return err
	}
	// start msgService
	ns.msgService.Start()
	ns.node.logger.Info("Started QlcService.")
	return nil
}

func (ns *QlcService) setEvent() error {
	err := ns.msgEvent.SubscribeAsync(string(common.EventBroadcast), ns.Broadcast, false)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.SubscribeAsync(string(common.EventSendMsgToPeers), ns.SendMessageToPeers, false)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	return nil
}

func (ns *QlcService) unsubscribeEvent() error {
	err := ns.msgEvent.Unsubscribe(string(common.EventBroadcast), ns.Broadcast)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.Unsubscribe(string(common.EventSendMsgToPeers), ns.SendMessageToPeers)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	return nil
}

// Stop stop p2p manager.
func (ns *QlcService) Stop() error {
	//ns.node.logger.Info("Stopping QlcService...")

	ns.node.Stop()
	ns.dispatcher.Stop()
	ns.msgService.Stop()
	err := ns.unsubscribeEvent()
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Register register the subscribers.
func (ns *QlcService) Register(subscribers ...*Subscriber) {
	ns.dispatcher.Register(subscribers...)
}

// Deregister Deregister the subscribers.
func (ns *QlcService) Deregister(subscribers ...*Subscriber) {
	ns.dispatcher.Deregister(subscribers...)
}

// PutMessage put message to dispatcher.
func (ns *QlcService) PutMessage(msg *Message) {
	ns.dispatcher.PutMessage(msg)
}

// Broadcast message.
func (ns *QlcService) Broadcast(name string, value interface{}) {
	ns.node.BroadcastMessage(name, value)
}

func (ns *QlcService) SendMessageToPeers(messageName string, value interface{}, peerID string) {
	ns.node.SendMessageToPeers(messageName, value, peerID)
}

// SendMessageToPeer send message to a peer.
func (ns *QlcService) SendMessageToPeer(messageName string, value interface{}, peerID string) error {
	return ns.node.SendMessageToPeer(messageName, value, peerID)
}
