package pinger

import (
	"context"
	"encoding/binary"
	"io"
	"time"

	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"

	version2 "github.com/qlcchain/go-qlc/chain/version"
)

var log = logging.Logger("ping")

const (
	dataLen   = 4
	gitRevLen = 7
	PingSize  = 32
)

const ID = "/qlc/ping/1.0.0"

const pingTimeout = time.Second * 60

type PingService struct {
	Host host.Host
}

func NewPingService(h host.Host) *PingService {
	ps := &PingService{h}
	h.SetStreamHandler(ID, ps.PingHandler)
	return ps
}

func (p *PingService) PingHandler(s network.Stream) {
	buf := make([]byte, PingSize)
	errCh := make(chan error, 1)
	defer close(errCh)
	timer := time.NewTimer(pingTimeout)
	defer timer.Stop()

	go func() {
		select {
		case <-timer.C:
			log.Debug("ping timeout")
		case err, ok := <-errCh:
			if ok {
				log.Debug(err)
			} else {
				log.Error("ping loop failed without error")
			}
		}
		s.Reset()
	}()

	for {
		_, err := io.ReadFull(s, buf)
		if err != nil {
			errCh <- err
			return
		}
		buf1 := make([]byte, dataLen+len(version2.Version)+gitRevLen)
		copy(buf1[:dataLen], FromUint32(uint32(len(version2.Version))))
		copy(buf1[dataLen:dataLen+len(version2.Version)], []byte(version2.Version))
		copy(buf1[dataLen+len(version2.Version):dataLen+len(version2.Version)+gitRevLen], []byte(version2.GitRev))
		sbuf := make([]byte, PingSize)
		copy(sbuf[:len(buf1)], buf1)
		_, err = s.Write(sbuf)
		if err != nil {
			errCh <- err
			return
		}
		timer.Reset(pingTimeout)
	}
}

// Result is a result of a ping attempt, either an RTT or an error.
type Result struct {
	RTT     time.Duration
	Version string
	Error   error
}

func (ps *PingService) Ping(ctx context.Context, p peer.ID) <-chan Result {
	return Ping(ctx, ps.Host, p)
}

// Ping pings the remote peer until the context is canceled, returning a stream
// of RTTs or errors.
func Ping(ctx context.Context, h host.Host, p peer.ID) <-chan Result {
	s, err := h.NewStream(ctx, p, ID)
	if err != nil {
		ch := make(chan Result, 1)
		ch <- Result{Error: err}
		close(ch)
		return ch
	}

	ctx, cancel := context.WithCancel(ctx)

	out := make(chan Result)
	go func() {
		defer close(out)
		defer cancel()

		for ctx.Err() == nil {
			var res Result
			res.RTT, res.Version, res.Error = ping(s)

			// canceled, ignore everything.
			if ctx.Err() != nil {
				return
			}

			// No error, record the RTT.
			if res.Error == nil {
				h.Peerstore().RecordLatency(p, res.RTT)
			}

			select {
			case out <- res:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		// forces the ping to abort.
		<-ctx.Done()
		s.Reset()
	}()

	return out
}

func ping(s network.Stream) (time.Duration, string, error) {
	buf := make([]byte, PingSize)
	buf1 := make([]byte, dataLen+len(version2.Version)+gitRevLen)
	copy(buf1[:dataLen], FromUint32(uint32(len(version2.Version))))
	copy(buf1[dataLen:dataLen+len(version2.Version)], []byte(version2.Version))
	copy(buf1[dataLen+len(version2.Version):dataLen+len(version2.Version)+gitRevLen], []byte(version2.GitRev))

	before := time.Now()
	_, err := s.Write(buf)
	if err != nil {
		return 0, "", err
	}

	rbuf := make([]byte, PingSize)
	_, err = io.ReadFull(s, rbuf)
	if err != nil {
		return 0, "", err
	}
	l := binary.BigEndian.Uint32(rbuf[:dataLen])
	v := string(rbuf[dataLen : dataLen+l])
	hash := string(rbuf[dataLen+l : dataLen+l+gitRevLen])
	if len(v) == 0 || len(hash) == 0 {
		return time.Since(before), "", nil
	}
	version := v + "-" + hash
	return time.Since(before), version, nil
}

// FromUint32 decodes uint32.
func FromUint32(v uint32) []byte {
	b := make([]byte, dataLen)
	binary.BigEndian.PutUint32(b, v)
	return b
}
