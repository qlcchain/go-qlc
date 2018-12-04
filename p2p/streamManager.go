package p2p

import (
	"math/rand"
	"sync"
	"time"

	libnet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
)

// const
const (
	FindPeerInterval = time.Second * 30
)

// StreamManager manages all streams
type StreamManager struct {
	mu         sync.Mutex
	quitCh     chan bool
	allStreams *sync.Map
	node       *QlcNode
}

// NewStreamManager return a new stream manager
func NewStreamManager() *StreamManager {
	return &StreamManager{
		quitCh:     make(chan bool, 1),
		allStreams: new(sync.Map),
	}
}

// SetQlcService set netService
func (sm *StreamManager) SetQlcNode(node *QlcNode) {
	sm.node = node
}

// Start stream manager service
func (sm *StreamManager) Start() {
	logger.Info("Start Qlc StreamManager...")
	go sm.loop()
}

// Stop stream manager service
func (sm *StreamManager) Stop() {
	logger.Info("Stop Qlc StreamManager...")

	sm.quitCh <- true
}

// Add a new stream into the stream manager
func (sm *StreamManager) Add(s libnet.Stream) {
	stream := NewStream(s, sm.node)
	sm.AddStream(stream)
}

// AddStream into the stream manager
func (sm *StreamManager) AddStream(stream *Stream) {

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// check & close old stream
	if v, ok := sm.allStreams.Load(stream.pid.Pretty()); ok {
		old, _ := v.(*Stream)

		logger.Info("Removing old stream.")

		sm.allStreams.Delete(old.pid.Pretty())

		if old.stream != nil {
			old.stream.Close()
		}
	}

	logger.Info("Added a new stream.")

	sm.allStreams.Store(stream.pid.Pretty(), stream)
	stream.StartLoop()
}

// RemoveStream from the stream manager
func (sm *StreamManager) RemoveStream(s *Stream) {

	sm.mu.Lock()
	defer sm.mu.Unlock()

	v, ok := sm.allStreams.Load(s.pid.Pretty())
	if !ok {
		return
	}

	exist, _ := v.(*Stream)
	if s != exist {
		return
	}

	logger.Info("Removing a stream.")

	sm.allStreams.Delete(s.pid.Pretty())
}

// FindByPeerID find the stream with the given peerID
func (sm *StreamManager) FindByPeerID(peerID string) *Stream {
	v, _ := sm.allStreams.Load(peerID)
	if v == nil {
		return nil
	}
	return v.(*Stream)
}

// Find the stream with the given pid
func (sm *StreamManager) Find(pid peer.ID) *Stream {
	return sm.FindByPeerID(pid.Pretty())
}
func (sm *StreamManager) RandomPeer() (string, error) {
	allPeers := make(PeersSlice, 0)

	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if stream.IsConnected() {
			allPeers = append(allPeers, value)
		}
		return true
	})
	var peerID string
	rand.Seed(time.Now().Unix())
	if (len(allPeers)) == 0 {
		return "", ErrNoStream
	}
	randNum := rand.Intn(len(allPeers))
	for i, v := range allPeers {
		stream := v.(*Stream)
		if i == randNum {
			peerID = stream.pid.Pretty()
		}
	}
	return peerID, nil
}
func (sm *StreamManager) loop() {
	ticker := time.NewTicker(FindPeerInterval)
	sm.findPeers()
	for {
		select {
		case <-sm.quitCh:
			logger.Info("Stopped Stream Manager Loop.")
			return
		case <-ticker.C:
			sm.findPeers()
		}
	}
}

// CloseStream with the given pid and reason
func (sm *StreamManager) CloseStream(peerID string) {
	stream := sm.FindByPeerID(peerID)
	if stream != nil {
		stream.close()
	}
}

// findPeers
func (sm *StreamManager) findPeers() error {

	peers, err := sm.node.dhtFoundPeers()
	if err != nil {
		return err
	}
	for _, p := range peers {
		if p.ID == sm.node.ID || len(p.Addrs) == 0 {
			// No sense connecting to ourselves or if addrs are not available
			continue
		}
		sm.CreateStreamWithPeer(p.ID)
	}
	return nil
}

// CreateStreamWithPeer create stream with a peer.
func (sm *StreamManager) CreateStreamWithPeer(pid peer.ID) {

	stream := sm.Find(pid)

	if stream == nil {
		stream = NewStreamFromPID(pid, sm.node)
		sm.AddStream(stream)
	}
}

// BroadcastMessage broadcast the message
func (sm *StreamManager) BroadcastMessage(messageName string, messageContent []byte) {

	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		stream.SendMessage(messageName, messageContent)
		return true
	})
}
