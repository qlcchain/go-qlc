package p2p

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

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
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
	ping "github.com/qlcchain/go-qlc/p2p/pinger"
	"github.com/qlcchain/go-qlc/p2p/pubsub"
)

// Error types
var (
	ErrPeerIsNotConnected = errors.New("peer is not connected")
	ErrNoBootNode         = errors.New("can not get bootNode")
)

// p2p protocol version
var p2pVersion byte = 9

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
	getBootNodeInterval      = 10 * time.Second
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
	MessageTopic     *pubsub.Topic
	pubSub           *libp2pps.PubSub
	MessageSub       pubsub.Subscription
	isMiner          bool
	isRepresentative bool
	reporter         p2pmetrics.Reporter
	ping             *ping.Pinger
	connectionGater  *ConnectionGater
}

// NewNode return new QlcNode according to the config.
func NewNode(config *config.Config) (*QlcNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	node := &QlcNode{
		cfg:             config,
		ctx:             ctx,
		cancel:          cancel,
		streamManager:   NewStreamManager(),
		logger:          log.NewLogger("p2p"),
		isMiner:         config.PoV.PovEnabled,
		connectionGater: NewConnectionGater(),
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

func (node *QlcNode) updateWhiteList(id string, url string) {
	if node.cfg.WhiteList.Enable {
		if node.connectionGater != nil {
			var b bool
			peerId, err := peer.Decode(id)
			if err != nil {
				return
			}
			for _, v := range node.connectionGater.whiteList {
				if v.id == peerId {
					b = true
					break
				}
			}
			if !b {
				wl := WhiteList{}
				wl.id = peerId
				ss := strings.Split(url, ":")
				if len(ss) >= 2 {
					multiAddrString := "/ip4/" + ss[0] + "/tcp/" + ss[1]
					multiAddr, err := ma.NewMultiaddr(multiAddrString)
					if err != nil {
						return
					}
					wl.addr = multiAddr
				}
				node.connectionGater.whiteList = append(node.connectionGater.whiteList, wl)
			}
		}
	}
}

func (node *QlcNode) buildHost() error {
	var err error
	go node.getBootNode(node.cfg.P2P.BootNodes)
	node.logger.Info("Start Qlc Host...")
	sourceMultiAddr, _ := ma.NewMultiaddr(node.cfg.P2P.Listen)
	if node.cfg.WhiteList.Enable {
		node.host, err = libp2p.New(
			node.ctx,
			libp2p.ListenAddrs(sourceMultiAddr),
			libp2p.Identity(node.privateKey),
			//libp2p.NATPortMap(),
			libp2p.BandwidthReporter(node.reporter),
			libp2p.Ping(false),
			libp2p.ConnectionGater(node.connectionGater),
			// libp2p.NoSecurity,
			// libp2p.DefaultMuxers,
		)
		if err != nil {
			return err
		}
	} else {
		node.host, err = libp2p.New(
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
	}
	node.host.SetStreamHandler(QlcProtocolID, node.handleStream)
	node.kadDht, err = dht.New(node.ctx, node.host, dht.Mode(dht.ModeServer))
	if err != nil {
		return err
	}
	node.dis = discovery.NewRoutingDiscovery(node.kadDht)
	node.peerStore = node.host.Peerstore()
	if err := node.peerStore.AddPrivKey(node.ID, node.privateKey); err != nil {
		return err
	}
	if err := node.peerStore.AddPubKey(node.ID, node.privateKey.GetPublic()); err != nil {
		return err
	}
	node.streamManager.SetQlcNodeAndMaxStreamNum(node)
	// Set up libp2p pubsub
	node.pubSub, err = libp2pps.NewGossipSub(node.ctx, node.host, libp2pps.WithMessageSigning(false))
	if err != nil {
		return errors.New("failed to set up pubsub")
	}
	tp, err := node.pubSub.Join(MsgTopic)
	if err != nil {
		return err
	}
	node.MessageTopic = pubsub.NewTopic(tp)
	node.MessageSub, err = node.pubsubscribe(node.ctx, node.MessageTopic, node.processMessage)
	if err != nil {
		return err
	}
	// New ping service
	node.ping = ping.NewPinger(node.host)
	if node.cfg.P2P.IsBootNode {
		node.boostrapAddrs = append(node.boostrapAddrs, strings.ReplaceAll(node.cfg.P2P.Listen, "0.0.0.0", node.cfg.P2P.ListeningIp)+"/p2p/"+node.cfg.P2P.ID.PeerID)
	}
	pi := &types.PeerInfo{
		PeerID:         node.ID.Pretty(),
		Address:        strings.ReplaceAll(node.cfg.P2P.Listen, "0.0.0.0", node.cfg.P2P.ListeningIp),
		LastUpdateTime: time.Now().Format(time.RFC3339),
	}
	_ = node.netService.msgService.ledger.AddPeerInfo(pi)
	node.streamManager.AddOnlineInfo(pi)
	return nil
}

// Subscribes a handler function to a pubsub topic.
func (node *QlcNode) pubsubscribe(ctx context.Context, topic *pubsub.Topic, handler pubSubHandler) (pubsub.Subscription, error) {
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, errors.New("failed to subscribe")
	}
	go node.handleSubscription(ctx, sub, handler)
	return sub, nil
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
			node.startPeerDiscovery()
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
		node.logger.Error(err)
		return
	}
}

func (node *QlcNode) startPeerDiscovery() {
	ticker := time.NewTicker(time.Duration(node.cfg.P2P.Discovery.DiscoveryInterval) * time.Second)
	ticker1 := time.NewTicker(ConnectBootstrapInterval)
	pInfoS, err := convertPeers(node.boostrapAddrs)
	if err != nil {
		node.logger.Errorf("Failed to convert bootNode address")
	}
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
			pInfoS, err := convertPeers(node.boostrapAddrs)
			if err != nil {
				node.logger.Errorf("Failed to convert bootNode address")
			}
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
	node.updatePeersInfo(peers)
	for _, p := range peers {
		if p.ID == node.ID || len(p.Addrs) == 0 {
			// No sense connecting to ourselves or if addrs are not available
			continue
		}
		node.streamManager.createStreamWithPeer(p.ID)
	}
	return nil
}

func (node *QlcNode) updatePeersInfo(peers []peer.AddrInfo) {
	node.streamManager.onlinePeersInfo.Range(func(key, value interface{}) bool {
		k := key.(string)
		var i int
		for i = 0; i < len(peers); i++ {
			if peers[i].ID.Pretty() == k {
				break
			}
		}
		if i == len(peers) {
			if k != node.ID.Pretty() {
				node.streamManager.onlinePeersInfo.Delete(k)
			}
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
				pi.Address = strings.ReplaceAll(node.cfg.P2P.Listen, "0.0.0.0", node.cfg.P2P.ListeningIp)
			} else if len(p.Addrs) != 0 {
				pi.Address = findPublicIP(p.Addrs)
			}
		}
		pi.LastUpdateTime = time.Now().Format(time.RFC3339)
		_ = node.netService.msgService.ledger.AddPeerInfo(pi)
		node.streamManager.AddOnlineInfo(pi)
	}
}

func (node *QlcNode) handleStream(s network.Stream) {
	node.logger.Infof("Got a new stream from %s!", s.Conn().RemotePeer().Pretty())
	var addrs []ma.Multiaddr
	var infos []peer.AddrInfo
	addrs = append(addrs, s.Conn().RemoteMultiaddr())
	info := peer.AddrInfo{
		ID:    s.Conn().RemotePeer(),
		Addrs: addrs,
	}
	infos = append(infos, info)
	node.updatePeersInfo(infos)
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

func (node *QlcNode) handleSubscription(ctx context.Context, sub pubsub.Subscription, handler pubSubHandler) {
	for {
		received, err := sub.Next(ctx)
		if err != nil {
			if ctx.Err() != context.Canceled {
				node.logger.Errorf("error reading message from topic %s: %s", sub.Topic(), err)
			}
			return
		}

		if err := handler(ctx, received); err != nil {
			handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
			if err != context.Canceled {
				node.logger.Errorf("error in handler %s for topic %s: %s", handlerName, sub.Topic(), err)
			}
		}
	}
}

func (node *QlcNode) processMessage(ctx context.Context, pubSubMsg pubsub.Message) error {
	var message *QlcMessage
	var err error
	data := pubSubMsg.GetData()
	peerID := pubSubMsg.GetSender().Pretty()
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

type pubSubHandler func(ctx context.Context, msg pubsub.Message) error

func (node *QlcNode) GetBandwidthStats(stats *p2pmetrics.Stats) {
	*stats = node.reporter.GetBandwidthTotals()
}

func (node *QlcNode) getBootNode(urls []string) {
	ticker := time.NewTicker(getBootNodeInterval)
	for {
		select {
		case <-node.ctx.Done():
			return
		case <-ticker.C:
			for _, v := range urls {
				boot, err := accessHttpServer(v)
				if err != nil {
					continue
				}
				if node.cfg.WhiteList.Enable {
					ss := strings.Split(boot, "/")
					if len(ss) >= 7 {
						multiAddrString := fmt.Sprintf("%s:%s", ss[2], ss[4])
						node.updateWhiteList(ss[6], multiAddrString)
					}
				}
				if len(node.boostrapAddrs) == 0 {
					node.boostrapAddrs = append(node.boostrapAddrs, boot)
				} else {
					for k, v := range node.boostrapAddrs {
						if v == boot {
							break
						}
						if k == len(node.boostrapAddrs)-1 {
							node.boostrapAddrs = append(node.boostrapAddrs, boot)
						}
					}
				}
			}
		}
	}
}

func accessHttpServer(url string) (string, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	s := strings.Contains(string(body), "p2p")
	if s {
		return string(body), nil
	} else {
		return "", errors.New("body is not a bootNode")
	}
}
