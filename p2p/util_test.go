package p2p

import (
	"testing"

	ma "github.com/multiformats/go-multiaddr"
)

func TestIsPublicIP(t *testing.T) {
	m1, _ := ma.NewMultiaddr("/ip4/169.254.28.138/tcp/19734")
	m2, _ := ma.NewMultiaddr("/ip4/192.168.80.1/tcp/19734")
	m3, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/19734")
	var ips []ma.Multiaddr
	ips = append(ips, m1)
	ips = append(ips, m2)
	ips = append(ips, m3)
	b := judgeNetAttribute(ips)
	if b != Intranet {
		t.Fatal("should be Intranet")
	}
	m4, _ := ma.NewMultiaddr("/ip4/47.75.145.146/tcp/19734")
	ips = append(ips, m4)
	b1 := judgeNetAttribute(ips)
	if b1 != PublicNet {
		t.Fatal("should be publicNet")
	}
}
