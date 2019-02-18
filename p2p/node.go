package p2p

import (
	"context"
	"errors"
	"time"

	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	localdiscovery "github.com/libp2p/go-libp2p/p2p/discovery"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/qlcchain/go-qlc/config"
)

// Error types
var (
	ErrPeerIsNotConnected = errors.New("peer is not connected")
)

//var logger = log.NewLogger("p2p")

type QlcNode struct {
	ID             peer.ID
	privateKey     crypto.PrivKey
	cfg            *config.Config
	ctx            context.Context
	localDiscovery localdiscovery.Service
	host           host.Host
	peerStore      pstore.Peerstore
	boostrapAddrs  []string
	streamManager  *StreamManager
	ping           *PingService
	dis            *discovery.RoutingDiscovery
	kadDht         *dht.IpfsDHT
	netService     *QlcService
	logger         *zap.SugaredLogger
}

// NewNode return new QlcNode according to the config.
func NewNode(config *config.Config) (*QlcNode, error) {
	node := &QlcNode{
		cfg:           config,
		ctx:           context.Background(),
		boostrapAddrs: config.P2P.BootNodes,
		streamManager: NewStreamManager(),
		ID:            peer.ID(config.ID.PeerID),
		logger:        log.NewLogger("p2p"),
	}
	privateKey, err := config.DecodePrivateKey()
	if err != nil {
		node.logger.Error(err)
		return nil, err
	}
	node.privateKey = privateKey
	node.streamManager.SetQlcNode(node)

	return node, nil
}
func (node *QlcNode) startHost() error {
	node.logger.Info("Start Qlc Host...")
	sourceMultiAddr, _ := ma.NewMultiaddr(node.cfg.P2P.Listen)
	qlchost, err := libp2p.New(
		node.ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(node.privateKey),
		libp2p.NATPortMap(),
	)
	if err != nil {
		return err
	}
	qlchost.SetStreamHandler(QlcProtocolID, node.handleStream)
	node.host = qlchost
	kadDht, err := dht.New(node.ctx, node.host)
	if err != nil {
		return err
	}
	node.kadDht = kadDht
	node.ping = NewPingService(node.host)
	node.peerStore = qlchost.Peerstore()
	node.peerStore.AddPrivKey(node.ID, node.privateKey)
	node.peerStore.AddPubKey(node.ID, node.privateKey.GetPublic())
	return nil
}
func (node *QlcNode) startLocalDiscovery() error {
	// setup local discovery
	node.logger.Info("Start Qlc Local Discovery...")
	do := setupDiscoveryOption(node.cfg)
	if do != nil {
		service, err := do(node.ctx, node.host)
		if err != nil {
			node.logger.Error("mdns error: ", err)
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
		err := node.startLocalDiscovery()
		if err != nil {
			return err
		}
	}

	if len(node.boostrapAddrs) != 0 {
		pinfos, err := convertPeers(node.boostrapAddrs)
		if err != nil {
			node.logger.Errorf("Failed to convert bootnode address")
			return err
		}

		for {
			err = bootstrapConnect(node.ctx, node.host, pinfos)
			if err != nil {
				time.Sleep(time.Duration(10) * time.Second)
				continue
			}
			node.logger.Info("connect to bootstrap success")
			break

		}
		dis := discovery.NewRoutingDiscovery(node.kadDht)
		for {
			_, err := dis.Advertise(node.ctx, QlcProtocolFOUND)
			if err != nil {
				time.Sleep(time.Duration(1) * time.Second)
				continue
			}
			node.dis = dis
			break
		}
		if node.streamManager != nil {
			node.streamManager.Start()
		}
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
func (node *QlcNode) stopHost() {

	if node.host == nil {
		return
	}

	node.host.Close()
}

// SetQlcService set netService
func (node *QlcNode) SetQlcService(ns *QlcService) {
	node.netService = ns
}

// Stop stop a node.
func (node *QlcNode) Stop() {
	node.logger.Info("Stop QlcService Node...")

	node.stopHost()
	node.streamManager.Stop()
}

// BroadcastMessage broadcast message.
func (node *QlcNode) BroadcastMessage(messageName string, value interface{}) {

	node.streamManager.BroadcastMessage(messageName, value)
}

// SendMessageToPeer send message to a peer.
func (node *QlcNode) SendMessageToPeer(messageName string, value interface{}, peerID string) error {
	stream := node.streamManager.FindByPeerID(peerID)
	if stream == nil {
		node.logger.Debug("Failed to locate peer's stream")
		return ErrPeerIsNotConnected
	}
	data, err := MarshalMessage(messageName, value)
	if err != nil {
		return err
	}
	return stream.SendMessage(messageName, data)
}
