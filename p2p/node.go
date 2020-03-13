package p2p

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/qlcchain/go-qlc/common/types"

	ping "github.com/qlcchain/go-qlc/p2p/pinger"

	"github.com/qlcchain/go-qlc/common/topic"

	"github.com/qlcchain/go-qlc/p2p/pubsub"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	p2pmetrics "github.com/libp2p/go-libp2p-core/metrics"
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
	ErrNoBootNode         = errors.New("can not get bootNode")
)

// p2p protocol version
var p2pVersion byte = 8

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

const (
	MaxPingTimeOutTimes      = 4
	PingTimeInterval         = 30 * time.Second
	ConnectBootstrapInterval = 20 * time.Second
	PublishConnectPeersInfo  = 20 * time.Second
	PublishOnlinePeersInfo   = 25 * time.Second
	PublishBandWithPeersInfo = 60 * time.Second
)

type QlcNode struct {
	ID               peer.ID
	privateKey       crypto.PrivKey
	cfg              *config.Config
	ctx              context.Context
	cancel           context.CancelFunc
	localDiscovery   localdiscovery.Service
	host             host.Host
	peerStore        peerstore.Peerstore
	boostrapAddrs    []string
	streamManager    *StreamManager
	dis              *discovery.RoutingDiscovery
	kadDht           *dht.IpfsDHT
	netService       *QlcService
	logger           *zap.SugaredLogger
	publisher        *pubsub.Publisher
	subscriber       *pubsub.Subscriber
	MessageSub       pubsub.Subscription
	isMiner          bool
	isRepresentative bool
	reporter         p2pmetrics.Reporter
	ping             *ping.Pinger
}

// NewNode return new QlcNode according to the config.
func NewNode(config *config.Config) (*QlcNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	node := &QlcNode{
		cfg:           config,
		ctx:           ctx,
		cancel:        cancel,
		streamManager: NewStreamManager(),
		logger:        log.NewLogger("p2p"),
		isMiner:       config.PoV.PovEnabled,
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
	node.reporter = p2pmetrics.NewBandwidthCounter()
	return node, nil
}

func (node *QlcNode) setRepresentativeNode(isRepresentative bool) {
	node.isRepresentative = isRepresentative
}

func (node *QlcNode) buildHost() error {
	bns := getBootNode(node.cfg.P2P.BootNodes)
	if len(bns) == 0 {
		return ErrNoBootNode
	}
	node.boostrapAddrs = bns
	node.logger.Info("Start Qlc Host...")
	sourceMultiAddr, _ := ma.NewMultiaddr(node.cfg.P2P.Listen)
	qlcHost, err := libp2p.New(
		node.ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(node.privateKey),
		//libp2p.NATPortMap(),
		libp2p.BandwidthReporter(node.reporter),
		libp2p.Ping(false),
		// libp2p.NoSecurity,
		// libp2p.DefaultMuxers,
	)
	if err != nil {
		return err
	}
	qlcHost.SetStreamHandler(QlcProtocolID, node.handleStream)
	node.host = qlcHost
	kadDht, err := dht.New(node.ctx, node.host)
	if err != nil {
		return err
	}
	node.kadDht = kadDht
	node.dis = discovery.NewRoutingDiscovery(node.kadDht)
	node.peerStore = qlcHost.Peerstore()
	if err := node.peerStore.AddPrivKey(node.ID, node.privateKey); err != nil {
		return err
	}
	if err := node.peerStore.AddPubKey(node.ID, node.privateKey.GetPublic()); err != nil {
		return err
	}
	node.streamManager.SetQlcNodeAndMaxStreamNum(node)
	// Set up libp2p pubsub
	gsub, err := libp2pps.NewGossipSub(node.ctx, node.host, libp2pps.WithMessageSigning(false))
	if err != nil {
		return errors.New("failed to set up pubsub")
	}
	node.publisher = pubsub.NewPublisher(gsub)
	node.subscriber = pubsub.NewSubscriber(gsub)
	// New ping service
	node.ping = ping.NewPinger(node.host)
	return nil
}

func (node *QlcNode) startLocalDiscovery() error {
	// setup local discovery
	node.logger.Info("Start Qlc Local Discovery...")
	do := setupDiscoveryOption(node.cfg)
	if do != nil {
		service, err := do(node.ctx, node.host)
		if err != nil {
			// node.logger.Error("mDNS error: ", err)
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
		err := node.buildHost()
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
				node.logger.Warnf("mDNS start fail: %s", err)
			}
		}
	}

	go func() {
		node.startPingService()
	}()
	go func() {
		node.publishPeersInfoToChainContext()
	}()

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
	var err error
	ticker := time.NewTicker(PingTimeInterval)
	for {
		select {
		case <-node.ctx.Done():
			node.logger.Info("Stopped pingService Loop.")
			return
		case <-ticker.C:
			node.streamManager.allStreams.Range(func(key, value interface{}) bool {
				stream := value.(*Stream)
				if stream.pid != node.ID && stream.IsConnected() {
					if stream.pingResult == nil {
						stream.pingResult = make(<-chan ping.Result, 1)
						stream.pingResult, err = stream.node.ping.Ping(stream.pingCtx, stream.pid)
						if err != nil {
							stream.pingTimeoutTimes++
							if stream.pingTimeoutTimes >= MaxPingTimeOutTimes {
								_ = stream.close()
							}
						}
						res := <-stream.pingResult
						node.logger.Infof("ping peer %s took %f s version %s", stream.pid, res.RTT.Seconds(), res.Version)
						stream.rtt = res.RTT
						stream.pingTimeoutTimes = 0
						stream.lastUpdateTime = time.Now().Format(time.RFC3339)
						stream.globalVersion = res.Version
						pi := &types.PeerInfo{
							PeerID:         stream.pid.Pretty(),
							Address:        stream.addr.String(),
							Version:        stream.globalVersion,
							Rtt:            stream.rtt.Seconds(),
							LastUpdateTime: time.Now().Format(time.RFC3339),
						}

						node.streamManager.AddOrUpdateOnlineInfo(pi)
						_ = node.netService.msgService.ledger.AddOrUpdatePeerInfo(pi)
					} else {
						select {
						case res := <-stream.pingResult:
							if res.Error != nil {
								stream.pingTimeoutTimes++
								if stream.pingTimeoutTimes >= MaxPingTimeOutTimes {
									_ = stream.close()
								}
							} else {
								node.logger.Infof("ping peer %s took %f s version %s", stream.pid, res.RTT.Seconds(), res.Version)
								stream.rtt = res.RTT
								stream.pingTimeoutTimes = 0
								stream.lastUpdateTime = time.Now().Format(time.RFC3339)
								stream.globalVersion = res.Version
								pi := &types.PeerInfo{
									PeerID:         stream.pid.Pretty(),
									Address:        stream.addr.String(),
									Version:        stream.globalVersion,
									Rtt:            stream.rtt.Seconds(),
									LastUpdateTime: time.Now().Format(time.RFC3339),
								}
								node.streamManager.AddOrUpdateOnlineInfo(pi)
								_ = node.netService.msgService.ledger.AddOrUpdatePeerInfo(pi)
							}
						default:
						}
					}
				}
				return true
			})
		}
	}
}

func (node *QlcNode) publishPeersInfoToChainContext() {
	node.logger.Info("start publish Peers VInfo Loop.")
	ticker1 := time.NewTicker(PublishConnectPeersInfo)
	ticker2 := time.NewTicker(PublishOnlinePeersInfo)
	ticker3 := time.NewTicker(PublishBandWithPeersInfo)
	for {
		select {
		case <-node.ctx.Done():
			return
		case <-ticker1.C:
			var p []*types.PeerInfo
			node.streamManager.allStreams.Range(func(key, value interface{}) bool {
				stream := value.(*Stream)
				if stream.IsConnected() {
					ps := &types.PeerInfo{
						PeerID:         stream.pid.Pretty(),
						Address:        stream.addr.String(),
						Version:        stream.globalVersion,
						Rtt:            stream.rtt.Seconds(),
						LastUpdateTime: stream.lastUpdateTime,
					}
					p = append(p, ps)
				}
				return true
			})
			node.netService.MessageEvent().Publish(topic.EventPeersInfo, &topic.EventP2PConnectPeersMsg{PeersInfo: p})
		case <-ticker2.C:
			var p []*types.PeerInfo
			node.streamManager.onlinePeersInfo.Range(func(key, value interface{}) bool {
				ps := value.(*types.PeerInfo)
				p = append(p, ps)
				return true
			})
			node.netService.MessageEvent().Publish(topic.EventOnlinePeersInfo, &topic.EventP2POnlinePeersMsg{PeersInfo: p})
		case <-ticker3.C:
			stats := node.reporter.GetBandwidthTotals()
			bwState := &topic.EventBandwidthStats{
				TotalIn:  stats.TotalIn,
				TotalOut: stats.TotalOut,
				RateIn:   stats.RateIn,
				RateOut:  stats.RateOut,
			}
			node.netService.msgEvent.Publish(topic.EventGetBandwidthStats, bwState)
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
	ticker1 := time.NewTicker(ConnectBootstrapInterval)
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
	node.streamManager.onlinePeersInfo.Range(func(key, value interface{}) bool {
		k := key.(string)
		var i int
		for i = 0; i < len(peers); i++ {
			if peers[i].ID.Pretty() == k {
				break
			}
		}
		if i == len(peers) {
			node.streamManager.onlinePeersInfo.Delete(k)
		}
		return true
	})
	for _, p := range peers {
		var pi *types.PeerInfo
		if v, ok := node.streamManager.onlinePeersInfo.Load(p.ID.Pretty()); ok {
			pi = v.(*types.PeerInfo)
		} else {
			pi = &types.PeerInfo{
				PeerID: p.ID.Pretty(),
			}
			if p.ID == node.ID {
				pi.Address = node.cfg.P2P.Listen
			} else if len(p.Addrs) != 0 {
				pi.Address = findPublicIP(p.Addrs)
			}
		}
		pi.LastUpdateTime = time.Now().Format(time.RFC3339)
		_ = node.netService.msgService.ledger.AddPeerInfo(pi)
		node.streamManager.AddOnlineInfo(pi)
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
	if node.localDiscovery != nil {
		node.localDiscovery.UnregisterNotifee(node)
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
	node.logger.Debugf("node [%s] receive topic from [%s]", node.ID.Pretty(), peerID)
	message, err = ParseQlcMessage(data)
	if err != nil {
		return err
	}
	node.logger.Debug("message Type is :", message.messageType)
	messageBuffer := data[QlcMessageHeaderLength:]
	if len(messageBuffer) < int(message.DataLength()) {
		return errors.New("data length error")
	}
	if err := message.ParseMessageData(messageBuffer); err != nil {
		return err
	}
	if message.Version() < p2pVersion {
		node.logger.Debugf("message Version [%d] is less then p2pVersion [%d]", message.Version(), p2pVersion)
		return nil
	}
	m := NewMessage(message.MessageType(), peerID, message.MessageData(), message.content)
	node.netService.PutMessage(m)
	return nil
}

type pubSubProcessorFunc func(ctx context.Context, msg pubsub.Message) error

func (node *QlcNode) GetBandwidthStats(stats *p2pmetrics.Stats) {
	*stats = node.reporter.GetBandwidthTotals()
}

func getBootNode(urls []string) []string {
	var bn []string
	for _, v := range urls {
		url := "http://" + v + "/bootNode"
		rsp, err := http.Get(url)
		if err != nil {
			continue
		}
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			continue
		}
		bn = append(bn, string(body))
		_ = rsp.Body.Close()
	}
	return bn
}
