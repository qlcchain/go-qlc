package p2p

import (
	"fmt"
	"net"
)

type Protector struct {
	whiteList []string
}

func NewProtector() *Protector {
	return &Protector{}
}

func (pt *Protector) Protect(s net.Conn) (net.Conn, error) {
	remoteAddr := s.RemoteAddr().String()
	for _, v := range pt.whiteList {
		if v == remoteAddr {
			return s, nil
		}
	}
	return nil, fmt.Errorf("remote addr %s is not in whiteList", remoteAddr)
}

func (pt *Protector) Fingerprint() []byte {
	return nil
}
