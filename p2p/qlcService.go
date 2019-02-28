package p2p

import (
	"errors"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

// QlcService service for qlc p2p network
type QlcService struct {
	common.ServiceLifecycle
	node       *QlcNode
	dispatcher *Dispatcher
	msgEvent   *EventQueue
	msgService *MessageService
}

func (ns *QlcService) Init() error {
	if !ns.PreInit() {
		return errors.New("pre init fail")
	}
	defer ns.PostInit()

	return nil
}

func (ns *QlcService) Status() int32 {
	return ns.Status()
}

// NewQlcService create netService
func NewQlcService(cfg *config.Config) (*QlcService, error) {
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
	l := ledger.NewLedger(cfg.LedgerDir())
	msgService := NewMessageService(ns, l)
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
	if !ns.PreStart() {
		return errors.New("pre start fail")
	}
	defer ns.PostStart()
	ns.node.logger.Info("Starting QlcService...")

	// start dispatcher.
	ns.dispatcher.Start()

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

// Stop stop p2p manager.
func (ns *QlcService) Stop() error {
	if !ns.PreStop() {
		return errors.New("pre stop fail")
	}
	defer ns.PostStop()
	ns.node.logger.Info("Stopping QlcService...")

	ns.node.Stop()
	ns.dispatcher.Stop()

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
func (ns *QlcService) PutMessage(msg Message) {
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
