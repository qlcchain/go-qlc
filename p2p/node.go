package p2p

import (
	"context"
	"errors"
	"github.com/qlcchain/go-qlc/p2p/pubsub"
	"time"

	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pps "github.com/libp2p/go-libp2p-pubsub"
	localdiscovery "github.com/libp2p/go-libp2p/p2p/discovery"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/qlcchain/go-qlc/config"
)

// Error types
var (
	ErrPeerIsNotConnected = errors.New("peer is not connected")
)

// p2p protocol version
var p2pVersion = 5

type netAttribute byte

const (
	PublicNet netAttribute = iota
	Intranet
)

const (
	// Topic is the network pubsub topic identifier on which new messages are announced.
	MsgTopic = "/qlc/msgs"
	// BlockTopic is the pubsub topic identifier on which new blocks are announced.
	BlockTopic = "/qlc/blocks"
)

type QlcNode struct {
	ID                peer.ID
	privateKey        crypto.PrivKey
	cfg               *config.Config
	ctx               context.Context
	cancel            context.CancelFunc
	localDiscovery    localdiscovery.Service
	host              host.Host
	peerStore         peerstore.Peerstore
	boostrapAddrs     []string
	streamManager     *StreamManager
	ping              *Pinger
	dis               *discovery.RoutingDiscovery
	kadDht            *dht.IpfsDHT
	netService        *QlcService
	logger            *zap.SugaredLogger
	localNetAttribute netAttribute
	publisher         *pubsub.Publisher
	subscriber        *pubsub.Subscriber
	MessageSub        pubsub.Subscription
}

// NewNode return new QlcNode according to the config.
func NewNode(config *config.Config) (*QlcNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	node := &QlcNode{
		cfg:           config,
		ctx:           ctx,
		cancel:        cancel,
		boostrapAddrs: config.P2P.BootNodes,
		streamManager: NewStreamManager(),
		logger:        log.NewLogger("p2p"),
	}
	privateKey, err := config.DecodePrivateKey()
	if err != nil {
		node.logger.Error(err)
		return nil, err
	}
	node.privateKey = privateKey
	node.ID, err = peer.IDFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (node *QlcNode) startHost() error {
	node.logger.Info("Start Qlc Host...")
	sourceMultiAddr, _ := ma.NewMultiaddr(node.cfg.P2P.Listen)
	qlcHost, err := libp2p.New(
		node.ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(node.privateKey),
		//libp2p.NATPortMap(),
		libp2p.DefaultMuxers,
	)
	if err != nil {
		return err
	}
	ms := qlcHost.Addrs()
	node.localNetAttribute = judgeNetAttribute(ms)
	qlcHost.SetStreamHandler(QlcProtocolID, node.handleStream)
	node.host = qlcHost
	kadDht, err := dht.New(node.ctx, node.host)
	if err != nil {
		return err
	}
	node.kadDht = kadDht
	node.dis = discovery.NewRoutingDiscovery(node.kadDht)
	node.ping = NewPinger(node.host)
	node.peerStore = qlcHost.Peerstore()
	if err := node.peerStore.AddPrivKey(node.ID, node.privateKey); err != nil {
		return err
	}
	if err := node.peerStore.AddPubKey(node.ID, node.privateKey.GetPublic()); err != nil {
		return err
	}
	node.streamManager.SetQlcNodeAndMaxStreamNum(node)
	// Set up libp2p pubsub
	fsub, err := libp2pps.NewGossipSub(node.ctx, node.host)
	if err != nil {
		return errors.New("failed to set up pubsub")
	}
	node.publisher = pubsub.NewPublisher(fsub)
	node.subscriber = pubsub.NewSubscriber(fsub)
	return nil
}

func (node *QlcNode) startLocalDiscovery() error {
	// setup local discovery
	node.logger.Info("Start Qlc Local Discovery...")
	do := setupDiscoveryOption(node.cfg)
	if do != nil {
		service, err := do(node.ctx, node.host)
		if err != nil {
			node.logger.Error("mDNS error: ", err)
			return err
		} else {
			service.RegisterNotifee(node)
			node.localDiscovery = service
		}
	}
	return nil
}

func (node *QlcNode) StartServices() error {

	if node.host == nil {
		err := node.startHost()
		if err != nil {
			return err
		}
	}

	// subscribe to message notifications
	msgSub, err := node.subscriber.Subscribe(MsgTopic)
	if err != nil {
		return errors.New("failed to subscribe to message topic")
	}
	node.MessageSub = msgSub

	go node.handleSubscription(node.ctx, node.processMessage, "processMessage", node.MessageSub, "MessageSub")

	if node.localDiscovery == nil {
		if node.cfg.P2P.Discovery.MDNSEnabled {
			err := node.startLocalDiscovery()
			if err != nil {
				return err
			}
		}
	}

	if node.ping != nil {
		go func() {
			node.startPingService()
		}()
	}

	if node.dis != nil {
		go func() {
			pInfoS, err := convertPeers(node.boostrapAddrs)
			if err != nil {
				node.logger.Errorf("Failed to convert bootNode address")
				return
			}
			node.startPeerDiscovery(pInfoS)
		}()
	}

	return nil
}

func (node *QlcNode) startPingService() {
	node.logger.Info("start pingService Loop.")
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-node.ctx.Done():
			node.logger.Info("Stopped pingService Loop.")
			return
		case <-ticker.C:
			node.streamManager.allStreams.Range(func(key, value interface{}) bool {
				stream := value.(*Stream)
				if stream.pid != node.ID && stream.IsConnected() {
					ts, _ := node.ping.Ping(node.ctx, stream.pid)
					select {
					case res := <-ts:
						if res.Error != nil {
							node.logger.Errorf("error:[%s] when ping peer id :[%s]", res.Error, stream.pid)
						}
						node.logger.Debugf("ping peer %s,took: %f s", stream.pid, res.RTT.Seconds())
						stream.rtt = res.RTT
						stream.pingTimeoutCount = 0
					case <-time.After(time.Second * 4):
						node.logger.Error("failed to receive ping")
						stream.pingTimeoutCount++
						if stream.pingTimeoutCount == MaxPingTimeOutTimes {
							_ = stream.close()
						}
					}
				}
				return true
			})
		}
	}
}

func (node *QlcNode) connectBootstrap(pInfoS []peer.AddrInfo) {

	err := bootstrapConnect(node.ctx, node.host, pInfoS)
	if err != nil {
		return
	}
	node.logger.Info("connect to bootstrap success")

	_, err = node.dis.Advertise(node.ctx, QlcProtocolFOUND)
	if err != nil {
		return
	}
}

func (node *QlcNode) startPeerDiscovery(pInfoS []peer.AddrInfo) {
	ticker := time.NewTicker(time.Duration(node.cfg.P2P.Discovery.DiscoveryInterval) * time.Second)
	ticker1 := time.NewTicker(15 * time.Second)
	node.connectBootstrap(pInfoS)
	if err := node.findPeers(); err != nil {
		node.logger.Errorf("find node error[%s]", err)
	}
	for {
		select {
		case <-node.ctx.Done():
			node.logger.Info("Stopped peer discovery Loop.")
			return
		case <-ticker.C:
			if err := node.findPeers(); err != nil {
				node.logger.Errorf("find node error[%s]", err)
				continue
			}
		case <-ticker1.C:
			node.connectBootstrap(pInfoS)
		}
	}
}

// findPeers
func (node *QlcNode) findPeers() error {

	peers, err := node.dhtFoundPeers()
	if err != nil {
		return err
	}
	for _, p := range peers {
		if p.ID == node.ID || len(p.Addrs) == 0 {
			// No sense connecting to ourselves or if addrs are not available
			continue
		}
		node.streamManager.createStreamWithPeer(p.ID)
	}
	return nil
}

func (node *QlcNode) handleStream(s network.Stream) {
	node.logger.Infof("Got a new stream from %s!", s.Conn().RemotePeer().Pretty())
	node.streamManager.Add(s)
}

// ID return node ID.
func (node *QlcNode) GetID() string {
	return node.ID.Pretty()
}

// ID return node ID.
func (node *QlcNode) StreamManager() *StreamManager {
	return node.streamManager
}

func (node *QlcNode) stopHost() error {

	if node.host == nil {
		return errors.New("host not exit")
	}

	if err := node.host.Close(); err != nil {
		return err
	}
	return nil
}

// SetQlcService set netService
func (node *QlcNode) SetQlcService(ns *QlcService) {
	node.netService = ns
}

// Stop stop a node.
func (node *QlcNode) Stop() error {
	node.logger.Info("Stop QlcService Node...")
	if node.MessageSub != nil {
		node.MessageSub.Cancel()
		node.MessageSub = nil
	}
	node.cancel()
	if err := node.stopHost(); err != nil {
		return err
	}
	return nil
}

// BroadcastMessage broadcast message.
func (node *QlcNode) BroadcastMessage(messageName MessageType, value interface{}) {

	node.streamManager.BroadcastMessage(messageName, value)
}

// SendMessageToPeer send message to a peer.
func (node *QlcNode) SendMessageToPeer(messageName MessageType, value interface{}, peerID string) error {
	if messageName == BulkPushBlock {
		return nil
	}
	stream := node.streamManager.FindByPeerID(peerID)
	if stream == nil {
		node.logger.Debug("Failed to locate peer's stream")
		return ErrPeerIsNotConnected
	}
	data, err := marshalMessage(messageName, value)
	if err != nil {
		return err
	}
	return stream.SendMessageToPeer(messageName, data)
}

func (node *QlcNode) handleSubscription(ctx context.Context, f pubSubProcessorFunc, fname string, s pubsub.Subscription, sname string) {
	for {
		pubSubMsg, err := s.Next(ctx)
		if err != nil {
			node.logger.Debugf("%s.Next(): %s", sname, err)
			return
		}

		if err := f(ctx, pubSubMsg); err != nil {
			if err != context.Canceled {
				node.logger.Errorf("%s(): %s", fname, err)
			}
		}
	}
}

func (node *QlcNode) processMessage(ctx context.Context, pubSubMsg pubsub.Message) error {
	var message *QlcMessage
	var err error
	data := pubSubMsg.GetData()
	peerID := pubSubMsg.GetFrom().Pretty()
	if peerID == node.ID.Pretty() {
		return nil
	}
	node.logger.Infof("node [%s] receive topic from [%s]", node.ID.Pretty(), peerID)
	message, err = ParseQlcMessage(data)
	if err != nil {
		return err
	}
	node.logger.Info("message Type is :", message.messageType)
	messageBuffer := data[QlcMessageHeaderLength:]
	if len(messageBuffer) < int(message.DataLength()) {
		return errors.New("data length error")
	}
	if err := message.ParseMessageData(messageBuffer); err != nil {
		return err
	}

	if message.Version() < byte(p2pVersion) {
		node.logger.Debugf("message Version [%d] is less then p2pVersion [%d]", message.Version(), p2pVersion)
		return errors.New("P2P protocol version is low")
	}
	m := NewMessage(message.MessageType(), peerID, message.MessageData(), message.content)
	node.netService.PutMessage(m)
	return nil
}

type pubSubProcessorFunc func(ctx context.Context, msg pubsub.Message) error
