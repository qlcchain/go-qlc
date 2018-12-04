package p2p

import (
	"errors"
	"fmt"
	"sync"

	libnet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

// Stream Errors
var (
	ErrShouldCloseConnectionAndExitLoop = errors.New("should close connection and exit loop")
	ErrStreamIsNotConnected             = errors.New("stream is not connected")
	ErrNoStream                         = errors.New("no stream")
)

// Stream define the structure of a stream in p2p network
type Stream struct {
	syncMutex   sync.Mutex
	pid         peer.ID
	addr        ma.Multiaddr
	stream      libnet.Stream
	node        *QlcNode
	quitWriteCh chan bool
	messageChan chan []byte
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
		messageChan: make(chan []byte, 2*1024),
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
	messageBuffer := make([]byte, 0)

	var message *QlcMessage

	for {
		n, err := s.stream.Read(buf)
		if err != nil {
			logger.Debug("Error occurred when reading data from network connection.")
			s.close()
			return
		}

		messageBuffer = append(messageBuffer, buf[:n]...)

		for {
			if message == nil {
				var err error

				// waiting for header data.
				if len(messageBuffer) < QlcMessageHeaderLength {
					// continue reading.
					break
				}
				message, err = ParseQlcMessage(messageBuffer)
				if err != nil {
					return
				}
				messageBuffer = messageBuffer[QlcMessageHeaderLength:]

			}
			// waiting for data.
			if len(messageBuffer) < int(message.DataLength()) {
				// continue reading.
				break
			}
			message.content = append(message.content, messageBuffer[:message.DataLength()]...)
			// remove data from buffer.
			messageBuffer = messageBuffer[message.DataLength():]

			// handle message.
			if err := s.handleMessage(message); err == ErrShouldCloseConnectionAndExitLoop {
				return
			}

			// reset message.
			message = nil
		}
	}
}
func (s *Stream) writeLoop() {
	// ping func
	/*	ts, err := s.node.ping.Ping(s.node.ctx, s.pid)
		if err != nil {
			logger.Debug("ping error:", err)
			return
		}
	*/
	for {
		select {
		/*		case took := <-ts:
				//logger.Info("ping took: ", took)
				if took == 0 {
					logger.Debug("failed to receive ping")
					s.close()
					return
				}*/
		case <-s.quitWriteCh:
			logger.Debug("Quiting Stream Write Loop.")
			return
		case message := <-s.messageChan:
			s.WriteQlcMessage(message)
			continue
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

// SendMessage send msg to buffer
func (s *Stream) SendMessage(messageType string, data []byte) error {
	message := NewQlcMessage(data, messageType)
	s.messageChan <- message
	return nil
}

// WriteQlcMessage write qlc msg in the stream
func (s *Stream) WriteQlcMessage(message []byte) error {

	err := s.Write(message)

	return err
}
func (s *Stream) Write(data []byte) error {
	if s.stream == nil {
		s.close()
		return ErrStreamIsNotConnected
	}

	n, err := s.stream.Write(data)
	if err != nil {
		logger.Error("Failed to send message to peer.")
		s.close()
		return err
	}
	logger.Infof("%d byte send to %v ", n, s.pid.Pretty())
	return nil
}
func (s *Stream) handleMessage(message *QlcMessage) error {

	s.node.netService.PutMessage(NewBaseMessage(message.MessageType(), s.pid.Pretty(), message.MessageData()))

	return nil
}
