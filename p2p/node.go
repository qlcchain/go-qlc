package p2p

import (
	"context"
	"errors"
	"time"

	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"

	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	discovery "github.com/libp2p/go-libp2p-discovery"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	localdiscovery "github.com/libp2p/go-libp2p/p2p/discovery"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/qlcchain/go-qlc/config"
)

// Error types
var (
	ErrPeerIsNotConnected = errors.New("peer is not connected")
)

// p2p protocol version
var p2pVersion = 4

//var logger = log.NewLogger("p2p")

type QlcNode struct {
	ID                  peer.ID
	privateKey          crypto.PrivKey
	cfg                 *config.Config
	ctx                 context.Context
	localDiscovery      localdiscovery.Service
	host                host.Host
	peerStore           pstore.Peerstore
	boostrapAddrs       []string
	streamManager       *StreamManager
	ping                *PingService
	dis                 *discovery.RoutingDiscovery
	kadDht              *dht.IpfsDHT
	netService          *QlcService
	logger              *zap.SugaredLogger
	quitPeerDiscoveryCh chan bool
}

// NewNode return new QlcNode according to the config.
func NewNode(config *config.Config) (*QlcNode, error) {
	node := &QlcNode{
		cfg:                 config,
		ctx:                 context.Background(),
		boostrapAddrs:       config.P2P.BootNodes,
		streamManager:       NewStreamManager(),
		logger:              log.NewLogger("p2p"),
		quitPeerDiscoveryCh: make(chan bool, 1),
	}
	privateKey, err := config.DecodePrivateKey()
	if err != nil {
		node.logger.Error(err)
		return nil, err
	}
	node.privateKey = privateKey
	node.streamManager.SetQlcNode(node)
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
		libp2p.NATPortMap(),
		libp2p.DefaultMuxers,
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
	node.ping = NewPingService(node.host)
	node.peerStore = qlcHost.Peerstore()
	if err := node.peerStore.AddPrivKey(node.ID, node.privateKey); err != nil {
		return err
	}
	if err := node.peerStore.AddPubKey(node.ID, node.privateKey.GetPublic()); err != nil {
		return err
	}
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

	if node.localDiscovery == nil {
		if node.cfg.P2P.Discovery.MDNSEnabled {
			err := node.startLocalDiscovery()
			if err != nil {
				return err
			}
		}
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

	//if len(node.boostrapAddrs) != 0 {
	//	go node.connectBootstrap()
	//}
	return nil
}

func (node *QlcNode) connectBootstrap(pInfoS []pstore.PeerInfo) {

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

func (node *QlcNode) startPeerDiscovery(pInfoS []pstore.PeerInfo) {
	ticker := time.NewTicker(time.Duration(node.cfg.P2P.Discovery.DiscoveryInterval) * time.Second)
	ticker1 := time.NewTicker(15 * time.Second)
	node.connectBootstrap(pInfoS)
	if err := node.findPeers(); err != nil {
		node.logger.Errorf("find node error[%s]", err)
	}
	for {
		select {
		case <-node.quitPeerDiscoveryCh:
			node.logger.Info("Stopped peer discovery Loop.")
			return
		case <-ticker.C:
			if err := node.findPeers(); err != nil {
				node.logger.Errorf("find node error[%s]", err)
				continue
			}
		case <-ticker1.C:
			node.connectBootstrap(pInfoS)
		default:
			time.Sleep(5 * time.Millisecond)
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

func (node *QlcNode) handleStream(s inet.Stream) {
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

func (node *QlcNode) stopPeerDiscovery() {
	node.quitPeerDiscoveryCh <- true
}

// SetQlcService set netService
func (node *QlcNode) SetQlcService(ns *QlcService) {
	node.netService = ns
}

// Stop stop a node.
func (node *QlcNode) Stop() error {
	node.logger.Info("Stop QlcService Node...")
	node.stopPeerDiscovery()
	if err := node.stopHost(); err != nil {
		return err
	}
	return nil
}

// BroadcastMessage broadcast message.
func (node *QlcNode) BroadcastMessage(messageName string, value interface{}) {

	node.streamManager.BroadcastMessage(messageName, value)
}

// BroadcastMessage broadcast message.
func (node *QlcNode) SendMessageToPeers(messageName string, value interface{}, peerID string) {
	node.streamManager.SendMessageToPeers(messageName, value, peerID)
}

// SendMessageToPeer send message to a peer.
func (node *QlcNode) SendMessageToPeer(messageName string, value interface{}, peerID string) error {
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
