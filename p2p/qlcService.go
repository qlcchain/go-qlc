package p2p

import (
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

// QlcService service for qlc p2p network
type QlcService struct {
	node       *QlcNode
	dispatcher *Dispatcher
	msgEvent   *EventQueue
	msgService *MessageService
}

// NewQlcService create netService
func NewQlcService(cfg *config.Config, ledger *ledger.Ledger) (*QlcService, error) {
	node, err := NewNode(cfg)
	if err != nil {
		return nil, err
	}
	ns := &QlcService{
		node:       node,
		dispatcher: NewDispatcher(),
		msgEvent:   NeweventQueue(),
	}
	node.SetQlcService(ns)
	msgService := NewMessageService(ns, ledger)
	ns.msgService = msgService
	return ns, nil
}

// Node return the peer node
func (ns *QlcService) Node() *QlcNode {
	return ns.node
}

// EventQueue return EventQueue
func (ns *QlcService) MessageEvent() *EventQueue {
	return ns.msgEvent
}

// Start start p2p manager.
func (ns *QlcService) Start() error {
	logger.Info("Starting QlcService...")

	// start dispatcher.
	ns.dispatcher.Start()

	// start node.
	if err := ns.node.StartServices(); err != nil {
		ns.dispatcher.Stop()
		logger.Error("Failed to start QlcService.")
		return err
	}
	// start msgService
	ns.msgService.Start()
	logger.Info("Started QlcService.")
	return nil
}

// Stop stop p2p manager.
func (ns *QlcService) Stop() {
	logger.Info("Stopping QlcService...")

	ns.node.Stop()
	ns.dispatcher.Stop()
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
func (ns *QlcService) PutMessage(msg Message) {
	ns.dispatcher.PutMessage(msg)
}

// Broadcast message.
func (ns *QlcService) Broadcast(name string, msg []byte) {
	ns.node.BroadcastMessage(name, msg)
}

// SendMessageToPeer send message to a peer.
func (ns *QlcService) SendMessageToPeer(messageName string, data []byte, peerID string) error {
	return ns.node.SendMessageToPeer(messageName, data, peerID)
}
