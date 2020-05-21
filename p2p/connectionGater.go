package p2p

import (
	"github.com/libp2p/go-libp2p-core/control"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

type WhiteList struct {
	id   peer.ID
	addr ma.Multiaddr
}

type ConnectionGater struct {
	whiteList []WhiteList
}

func NewConnectionGater() *ConnectionGater {
	return &ConnectionGater{}
}

func (cg *ConnectionGater) InterceptPeerDial(peer.ID) bool {
	return true
}

func (cg *ConnectionGater) InterceptAddrDial(p peer.ID, addr ma.Multiaddr) bool {
	var allow bool
	for _, v := range cg.whiteList {
		if p == v.id {
			if v.addr != nil {
				if v.addr.Equal(addr) {
					allow = true
					break
				}
			} else {
				allow = true
				break
			}
		}
	}
	return allow
}

func (cg *ConnectionGater) InterceptAccept(network.ConnMultiaddrs) bool {
	return true
}

func (cg *ConnectionGater) InterceptSecured(network.Direction, peer.ID, network.ConnMultiaddrs) bool {
	return true
}

func (cg *ConnectionGater) InterceptUpgraded(network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}
