package p2p

import (
	"context"
	"time"

	discovery "github.com/libp2p/go-libp2p-discovery"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	localdiscovery "github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/qlcchain/go-qlc/config"
)

const (
	QlcProtocolID        = "qlc/1.0.0"
	QlcProtocolFOUND     = "/qlc/discovery/1.0.0"
	discoveryConnTimeout = time.Second * 30
)

func (node *QlcNode) dhtFoundPeers() ([]pstore.PeerInfo, error) {
	//discovery peers
	peers, err := discovery.FindPeers(node.ctx, node.dis, QlcProtocolFOUND, node.cfg.P2P.Discovery.Limit)
	if err != nil {
		return nil, err
	}
	node.logger.Infof("Found %d peers!", len(peers))
	//for _, p := range peers {
	//	node.logger.Info("Peer: ", p)
	//}
	return peers, nil
}

// HandlePeerFound attempts to connect to peer from `PeerInfo`.
func (node *QlcNode) HandlePeerFound(p pstore.PeerInfo) {
	ctx, cancel := context.WithTimeout(node.ctx, discoveryConnTimeout)
	defer cancel()
	if err := node.host.Connect(ctx, p); err != nil {
		node.logger.Error("Failed to connect to peer found by discovery: ", err)
	}
	node.logger.Info("find a local peer , ID:", p.ID.Pretty())
	node.streamManager.createStreamWithPeer(p.ID)
}

func setupDiscoveryOption(cfg *config.Config) DiscoveryOption {
	if cfg.P2P.Discovery.MDNSEnabled {
		return func(ctx context.Context, h host.Host) (localdiscovery.Service, error) {
			if cfg.P2P.Discovery.MDNSInterval == 0 {
				cfg.P2P.Discovery.MDNSInterval = 5
			}
			return localdiscovery.NewMdnsService(ctx, h, time.Duration(cfg.P2P.Discovery.MDNSInterval)*time.Second, QlcProtocolID)
		}
	}
	return nil
}

type DiscoveryOption func(context.Context, host.Host) (localdiscovery.Service, error)
