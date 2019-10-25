package p2p

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/qlcchain/go-qlc/common"
)

// Stream Errors
var (
	ErrStreamIsNotConnected = errors.New("stream is not connected")
	ErrNoStream             = errors.New("no stream")
	ErrCloseStream          = errors.New("stream close error")
)

// Message Priority.
const (
	MessagePriorityHigh = iota
	MessagePriorityNormal
	MessagePriorityLow
)

// Stream define the structure of a stream in p2p network
type Stream struct {
	syncMutex                 sync.Mutex
	pid                       peer.ID
	addr                      ma.Multiaddr
	stream                    network.Stream
	node                      *QlcNode
	quitWriteCh               chan bool
	messageNotifyChan         chan int
	highPriorityMessageChan   chan *QlcMessage
	normalPriorityMessageChan chan *QlcMessage
	lowPriorityMessageChan    chan *QlcMessage
	rtt                       time.Duration
}

// NewStream return a new Stream
func NewStream(stream network.Stream, node *QlcNode) *Stream {
	return newStreamInstance(stream.Conn().RemotePeer(), stream.Conn().RemoteMultiaddr(), stream, node)
}

// NewStreamFromPID return a new Stream based on the pid
func NewStreamFromPID(pid peer.ID, node *QlcNode) *Stream {
	return newStreamInstance(pid, nil, nil, node)
}

func newStreamInstance(pid peer.ID, addr ma.Multiaddr, stream network.Stream, node *QlcNode) *Stream {
	return &Stream{
		pid:                       pid,
		addr:                      addr,
		stream:                    stream,
		node:                      node,
		quitWriteCh:               make(chan bool, 1),
		messageNotifyChan:         make(chan int, 60*1024),
		highPriorityMessageChan:   make(chan *QlcMessage, 20*1024),
		normalPriorityMessageChan: make(chan *QlcMessage, 20*1024),
		lowPriorityMessageChan:    make(chan *QlcMessage, 20*1024),
	}
}

// Connect to the stream
func (s *Stream) Connect() error {
	//s.node.logger.Info("Connecting to peer.")

	// connect to host.
	stream, err := s.node.host.NewStream(
		s.node.ctx,
		s.pid,
		QlcProtocolID,
	)
	if err != nil {
		return err
	}
	//s.node.logger.Info("connect success to :", s.pid.Pretty())
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
			//			s.node.logger.Error(err)
			err = s.close()
			if err != nil {
				s.node.logger.Error(err)
			}
			return
		}
	}
	s.node.logger.Info("connect ", s.pid.Pretty(), " success")

	s.node.netService.MessageEvent().Publish(common.EventAddP2PStream, s.pid.Pretty())

	// loop.
	buf := make([]byte, 1024*4)
	messageBuffer := make([]byte, 0)

	var message *QlcMessage

	for {
		n, err := s.stream.Read(buf)
		if err != nil {
			s.node.logger.Debugf("Error occurred when reading data from network connection.")
			if err := s.close(); err != nil {
				s.node.logger.Error(err)
			}
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
			if err := message.ParseMessageData(messageBuffer); err != nil {
				return
			}
			// remove data from buffer.
			messageBuffer = messageBuffer[message.DataLength():]

			// handle message.
			s.handleMessage(message)
			// reset message.
			message = nil
		}
	}
}

func (s *Stream) writeLoop() {
	for {
		select {
		case <-s.quitWriteCh:
			s.node.logger.Debug("Quiting Stream Write Loop.")
			return
		case <-s.messageNotifyChan:
			select {
			case message := <-s.highPriorityMessageChan:
				_ = s.WriteQlcMessage(message)
				continue
			default:
			}

			select {
			case message := <-s.normalPriorityMessageChan:
				_ = s.WriteQlcMessage(message)
				continue
			default:
			}

			select {
			case message := <-s.lowPriorityMessageChan:
				_ = s.WriteQlcMessage(message)
				continue
			default:
			}
		}
	}
}

// Close close the stream
func (s *Stream) close() error {
	// Add lock & close flag to prevent multi call.
	//s.syncMutex.Lock()
	//defer s.syncMutex.Unlock()
	//s.node.logger.Info("Closing stream.")

	if s.stream != nil {
		s.node.netService.MessageEvent().Publish(common.EventDeleteP2PStream, s.pid.Pretty())
	}

	// cleanup.
	s.node.streamManager.RemoveStream(s)

	// quit.
	s.quitWriteCh <- true
	// close stream.
	if s.stream != nil {
		if err := s.stream.Close(); err != nil {
			return ErrCloseStream
		}
	}
	return nil
}

// SendMessage send msg to peer
func (s *Stream) SendMessageToPeer(messageType MessageType, data []byte) error {
	version := p2pVersion
	message := NewQlcMessage(data, byte(version), messageType)
	qlcMessage := &QlcMessage{
		messageType: messageType,
		content:     message,
	}

	err := s.SendMessageToChan(qlcMessage)

	return err
}

// WriteQlcMessage write qlc msg in the stream
func (s *Stream) WriteQlcMessage(message *QlcMessage) error {
	err := s.Write(message.content)

	return err
}

func (s *Stream) Write(data []byte) error {
	if s.stream == nil {
		if err := s.close(); err != nil {
			return ErrCloseStream
		}

		return ErrStreamIsNotConnected
	}

	// at least 5kb/s to write message
	deadline := time.Now().Add(time.Duration(len(data)/1024/5+1) * time.Second)
	if err := s.stream.SetWriteDeadline(deadline); err != nil {
		return err
	}

	n, err := s.stream.Write(data)
	if err != nil {
		s.node.logger.Debugf("Failed to send message to peer [%s].", s.pid.Pretty())
		//s.close()
		return err
	}
	s.node.logger.Debugf("%d byte send to %v ", n, s.pid.Pretty())
	return nil
}

func (s *Stream) handleMessage(message *QlcMessage) {
	if message.Version() < byte(p2pVersion) {
		s.node.logger.Debugf("message Version [%d] is less then p2pVersion [%d]", message.Version(), p2pVersion)
		return
	}
	m := NewMessage(message.MessageType(), s.pid.Pretty(), message.MessageData(), message.content)
	s.node.netService.PutSyncMessage(m)
}

// SendMessage send msg to buffer
func (s *Stream) SendMessageToChan(message *QlcMessage) error {
	var priority uint32
	if message.messageType == MessageResponse {
		priority = MessagePriorityHigh
	} else {
		priority = MessagePriorityNormal
	}
	switch priority {
	case MessagePriorityHigh:
		select {
		case s.highPriorityMessageChan <- message:
		default:
			s.node.logger.Debugf("Received too many normal priority message.")
			return nil
		}
	case MessagePriorityNormal:
		select {
		case s.normalPriorityMessageChan <- message:
		default:
			s.node.logger.Debugf("Received too many normal priority message.")
			return nil
		}
	default:
		select {
		case s.lowPriorityMessageChan <- message:
		default:
			s.node.logger.Debugf("Received too many low priority message.")
			return nil
		}
	}
	select {
	case s.messageNotifyChan <- 1:
	default:
		s.node.logger.Debugf("Received too many message notifyChan.")
		return nil
	}
	return nil
}
