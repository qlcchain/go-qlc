package p2p

import (
	"time"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
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
func NewQlcService(cfgFile string) (*QlcService, error) {
	cc := context.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	node, err := NewNode(cfg)
	if err != nil {
		return nil, err
	}
	ns := &QlcService{
		node:       node,
		dispatcher: NewDispatcher(),
		msgEvent:   cc.EventBus(),
	}
	node.SetQlcService(ns)
	l := ledger.NewLedger(cfgFile)
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
	err := ns.msgEvent.Subscribe(common.EventBroadcast, ns.Broadcast)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.Subscribe(common.EventSendMsgToSingle, ns.SendMessageToPeer)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.SubscribeSync(common.EventPeersInfo, ns.node.streamManager.GetAllConnectPeersInfo)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.Subscribe(common.EventFrontiersReq, ns.msgService.syncService.requestFrontiersFromPov)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.Subscribe(common.EventRepresentativeNode, ns.node.setRepresentativeNode)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.SubscribeSync(common.EventGetBandwidthStats, ns.node.GetBandwidthStats)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.SubscribeSync(common.EventConsensusSyncFinished, ns.msgService.syncService.onConsensusSyncFinished)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.SubscribeSync(common.EventSyncStatus, ns.msgService.syncService.GetSyncState)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	return nil
}

func (ns *QlcService) unsubscribeEvent() error {
	err := ns.msgEvent.Unsubscribe(common.EventBroadcast, ns.Broadcast)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.Unsubscribe(common.EventSendMsgToSingle, ns.SendMessageToPeer)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.Unsubscribe(common.EventFrontiersReq, ns.msgService.syncService.requestFrontiersFromPov)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	err = ns.msgEvent.Unsubscribe(common.EventRepresentativeNode, ns.node.setRepresentativeNode)
	if err != nil {
		ns.node.logger.Error(err)
		return err
	}
	return nil
}

// Stop stop p2p manager.
func (ns *QlcService) Stop() error {
	//ns.node.logger.Info("Stopping QlcService...")

	if err := ns.node.Stop(); err != nil {
		return err
	}
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
func (ns *QlcService) Deregister(subscribers *Subscriber) {
	ns.dispatcher.Deregister(subscribers)
}

// PutMessage put snyc message to dispatcher.
func (ns *QlcService) PutSyncMessage(msg *Message) {
	ns.dispatcher.PutSyncMessage(msg)
}

// PutMessage put dpos message to dispatcher.
func (ns *QlcService) PutMessage(msg *Message) {
	ns.dispatcher.PutMessage(msg)
}

// Broadcast message.
func (ns *QlcService) Broadcast(name MessageType, value interface{}) {
	ns.node.BroadcastMessage(name, value)
}

//func (ns *QlcService) SendMessageToPeers(messageName MessageType, value interface{}, peerID string) {
//	ns.node.SendMessageToPeers(messageName, value, peerID)
//}

// SendMessageToPeer send message to a peer.
func (ns *QlcService) SendMessageToPeer(messageName MessageType, value interface{}, peerID string) error {
	return ns.node.SendMessageToPeer(messageName, value, peerID)
}
