package p2p

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/qlcchain/go-qlc/vm/contract/abi"

	"github.com/qlcchain/go-qlc/common/topic"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/ledger"
)

// QlcService service for qlc p2p network
type QlcService struct {
	subscriber *event.ActorSubscriber
	node       *QlcNode
	dispatcher *Dispatcher
	msgEvent   event.EventBus
	msgService *MessageService
	cc         *chainctx.ChainContext
}

// NewQlcService create netService
func NewQlcService(cfgFile string) (*QlcService, error) {
	cc := chainctx.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	node, err := NewNode(cfg)
	if err != nil {
		return nil, err
	}
	ns := &QlcService{
		node:       node,
		dispatcher: NewDispatcher(),
		msgEvent:   cc.EventBus(),
		cc:         cc,
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
	//ns.node.logger.VInfo("Starting QlcService...")

	// start dispatcher.
	ns.dispatcher.Start()

	//set event
	if err := ns.setEvent(); err != nil {
		return err
	}
	// set whitelist
	if err := ns.setWhiteList(); err != nil {
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

func (ns *QlcService) setWhiteList() error {
	cfg, _ := ns.cc.Config()
	if cfg.WhiteList.Enable {
		ctx := vmstore.NewVMContext(ns.msgService.ledger)
		pn, err := abi.PermissionGetAllNodes(ctx)
		if err != nil {
			return err
		}
		for _, v := range pn {
			if len(v.NodeId) != 0 {
				ns.node.protector.whiteList = append(ns.node.protector.whiteList, v.NodeId)
			}
			ns.node.protector.whiteList = append(ns.node.protector.whiteList, v.NodeUrl)
		}
		var wls []string
		for _, v := range cfg.WhiteList.WhiteListInfos {
			if len(v.Addr) != 0 {
				wls = append(wls, v.Addr)
			}
			if len(v.PeerId) != 0 {
				wls = append(wls, v.PeerId)
			}
		}
		for _, v := range wls {
			var b bool
			for _, value := range ns.node.protector.whiteList {
				if value == v {
					b = true
					break
				}
			}
			if !b {
				ns.node.protector.whiteList = append(ns.node.protector.whiteList, v)
			}
		}
	}
	return nil
}

func (ns *QlcService) setEvent() error {
	ns.subscriber = event.NewActorSubscriber(event.SpawnWithPool(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *EventBroadcastMsg:
			ns.Broadcast(msg.Type, msg.Message)
		case *topic.EventBroadcastMsg:
			ns.Broadcast(MessageType(msg.Type), msg.Message)
		case *EventSendMsgToSingleMsg:
			if err := ns.SendMessageToPeer(msg.Type, msg.Message, msg.PeerID); err != nil {
				ns.node.logger.Error(err)
			}
		case *EventFrontiersReqMsg:
			ns.msgService.syncService.requestFrontiersFromPov(msg.PeerID)
		case bool:
			ns.node.setRepresentativeNode(msg)
		case *topic.EventP2PSyncStateMsg:
			ns.msgService.syncService.onConsensusSyncFinished()
		case *topic.PermissionEvent:
			if len(msg.NodeUrl) != 0 {
				ns.node.updateWhiteList(msg.NodeUrl)
			}
			if len(msg.NodeId) != 0 {
				ns.node.updateWhiteList(msg.NodeId)
			}
		}
	}), ns.msgEvent)

	if err := ns.subscriber.Subscribe(topic.EventBroadcast, topic.EventSendMsgToSingle, topic.EventFrontiersReq,
		topic.EventRepresentativeNode, topic.EventConsensusSyncFinished, topic.EventPermissionNodeUpdate); err != nil {
		ns.node.logger.Error(err)
		return err
	}

	return nil
}

// Stop stop p2p manager.
func (ns *QlcService) Stop() error {
	// ns.node.logger.VInfo("Stopping QlcService...")

	// this must be the first step
	err := ns.subscriber.UnsubscribeAll()
	if err != nil {
		return err
	}

	if err := ns.node.Stop(); err != nil {
		return err
	}

	ns.dispatcher.Stop()
	ns.msgService.Stop()

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
