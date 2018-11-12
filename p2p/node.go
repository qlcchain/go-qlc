package p2p

import (
	"context"
	"time"

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
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
)

var logger = common.NewLogger("p2p")

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
}

// NewNode return new QlcNode according to the config.
func NewNode(config *config.Config) (*QlcNode, error) {
	node := &QlcNode{
		cfg:           config,
		ctx:           context.Background(),
		boostrapAddrs: config.BootNodes,
	}
	node.streamManager = NewStreamManager(node)
	// load private key
	if err := node.LoadPrivateKey(); err != nil {
		return nil, err
	}
	return node, nil
}
func (node *QlcNode) LoadPrivateKey() error {
	if node.privateKey != nil {
		logger.Warn("private key already loaded")
		return nil
	}

	sk, err := LoadNetworkKeyFromFileOrCreateNew(node.cfg.PrivateKeyPath)
	if err != nil {
		return err
	}

	node.privateKey = sk
	node.ID, err = peer.IDFromPublicKey(sk.GetPublic())
	if err != nil {
		logger.Errorf("Failed to generate ID from network key file.")
		return err
	}
	return nil
}
func (node *QlcNode) startHost() error {
	logger.Info("Start Qlc Host...")
	sourceMultiAddr, _ := ma.NewMultiaddr(node.cfg.Listen)
	host, err := libp2p.New(
		node.ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(node.privateKey),
		libp2p.NATPortMap(),
	)
	if err != nil {
		return err
	}
	host.SetStreamHandler(QlcProtocolID, node.handleStream)
	node.host = host
	kadDht, err := dht.New(node.ctx, node.host)
	if err != nil {
		return err
	}
	node.kadDht = kadDht
	node.ping = NewPingService(node.host)
	node.peerStore = host.Peerstore()
	node.peerStore.AddPrivKey(node.ID, node.privateKey)
	node.peerStore.AddPubKey(node.ID, node.privateKey.GetPublic())
	return nil
}
func (node *QlcNode) startLocalDiscovery() error {
	// setup local discovery
	logger.Info("Start Qlc Local Discovery...")
	do := setupDiscoveryOption(node.cfg)
	if do != nil {
		service, err := do(node.ctx, node.host)
		if err != nil {
			logger.Error("mdns error: ", err)
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
			logger.Errorf("Failed to convert bootnode address")
			return err
		}

		for {
			err = bootstrapConnect(node.ctx, node.host, pinfos)
			if err != nil {
				time.Sleep(time.Duration(10) * time.Second)
				continue
			}
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
	logger.Infof("Got a new stream from %s!", s.Conn().RemotePeer().Pretty())
	node.streamManager.Add(s)
}

// ID return node ID.
func (node *QlcNode) GetID() string {
	return node.ID.Pretty()
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
	logger.Info("Stop QlcService Node...")

	node.stopHost()
	node.streamManager.Stop()
}

// BroadcastMessage broadcast message.
func (node *QlcNode) BroadcastMessage(messageName string, data []byte) {

	node.streamManager.BroadcastMessage(messageName, data)
}
