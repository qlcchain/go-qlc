package p2p

import (
	"fmt"
	"sync"
	"time"

	libnet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

const PingInterval = time.Second * 30

// Stream define the structure of a stream in p2p network
type Stream struct {
	syncMutex   sync.Mutex
	pid         peer.ID
	addr        ma.Multiaddr
	stream      libnet.Stream
	node        *QlcNode
	quitWriteCh chan bool
}

// NewStream return a new Stream
func NewStream(stream libnet.Stream, node *QlcNode) *Stream {
	return newStreamInstance(stream.Conn().RemotePeer(), stream.Conn().RemoteMultiaddr(), stream, node)
}

// NewStreamFromPID return a new Stream based on the pid
func NewStreamFromPID(pid peer.ID, node *QlcNode) *Stream {
	return newStreamInstance(pid, nil, nil, node)
}

func newStreamInstance(pid peer.ID, addr ma.Multiaddr, stream libnet.Stream, node *QlcNode) *Stream {
	return &Stream{
		pid:         pid,
		addr:        addr,
		stream:      stream,
		node:        node,
		quitWriteCh: make(chan bool, 1),
	}
}

// Connect to the stream
func (s *Stream) Connect() error {
	logger.Info("Connecting to peer.")

	// connect to host.
	stream, err := s.node.host.NewStream(
		s.node.ctx,
		s.pid,
		QlcProtocolID,
	)
	if err != nil {
		return err
	}
	logger.Info("connect success to :", s.pid)
	s.stream = stream
	s.addr = stream.Conn().RemoteMultiaddr()
	return nil
}

// IsConnected return if the stream is connected
func (s *Stream) IsConnected() bool {
	return s.stream != nil
}

func (s *Stream) String() string {
	addrStr := ""
	if s.addr != nil {
		addrStr = s.addr.String()
	}

	return fmt.Sprintf("Peer Stream: %s,%s", s.pid.Pretty(), addrStr)
}

// StartLoop start stream ping loop.
func (s *Stream) StartLoop() {
	go s.writeLoop()
	go s.readLoop()
}
func (s *Stream) readLoop() {

	if !s.IsConnected() {
		if err := s.Connect(); err != nil {
			logger.Debug(err)
			s.close()
			return
		}

	}
	logger.Info("connect ", s.pid.Pretty(), " success")
	// loop.
	buf := make([]byte, 1024*4)

	for {
		logger.Info("wait for data from stream")
		_, err := s.stream.Read(buf)
		if err != nil {
			logger.Errorf("Error occurred when reading data from network connection.")
			s.close()
			return
		}
	}
}
func (s *Stream) writeLoop() {
	// ping func
	ts, err := s.node.ping.Ping(s.node.ctx, s.pid)
	if err != nil {
		logger.Debug("ping error:", err)
		return
	}

	for {
		select {
		case took := <-ts:
			logger.Info("ping took: ", took)
			if took == 0 {
				logger.Debug("failed to receive ping")
				s.close()
				return
			}
		case <-s.quitWriteCh:
			logger.Debug("Quiting Stream Write Loop.")
			return
		}
	}

}

// Close close the stream
func (s *Stream) close() {
	// Add lock & close flag to prevent multi call.
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()
	logger.Info("Closing stream.")

	// cleanup.
	s.node.streamManager.RemoveStream(s)

	// quit.
	s.quitWriteCh <- true

	// close stream.
	if s.stream != nil {
		s.stream.Close()
	}
}
