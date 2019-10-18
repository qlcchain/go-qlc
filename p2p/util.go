package p2p

import (
	ma "github.com/multiformats/go-multiaddr"
	"net"
	"strings"
)

func judgeNetAttribute(ms []ma.Multiaddr) netAttribute {
	na := Intranet
	var ips []string
	if len(ms) != 0 {
		for _, m := range ms {
			s := m.String()
			ss := strings.Split(s, "/")
			ips = append(ips, ss[2])
		}
	} else {
		return na
	}

	for _, v := range ips {
		ip := new(net.IP)
		if v != "" {
			_ = ip.UnmarshalText([]byte(v))
		} else {
			continue
		}
		if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
			continue
		}
		if ip4 := ip.To4(); ip4 != nil {
			switch true {
			case ip4[0] == 10:
			case ip4[0] == 169 && ip4[1] == 254:
			case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			case ip4[0] == 192 && ip4[1] == 168:
			default:
				na = PublicNet
				return na
			}
		} else {
			continue
		}
	}
	return na
}
