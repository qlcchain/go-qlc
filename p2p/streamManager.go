package p2p

import (
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/qlcchain/go-qlc/common/types"
)

const (
	MaxStreamNumForPublicNet = 50
	MaxStreamNumForIntranet  = 8
)

// StreamManager manages all streams
type StreamManager struct {
	mu               sync.Mutex
	allStreams       *sync.Map
	activePeersCount int32
	maxStreamNum     int32
	node             *QlcNode
}

// NewStreamManager return a new stream manager
func NewStreamManager() *StreamManager {
	return &StreamManager{
		allStreams:       new(sync.Map),
		activePeersCount: 0,
	}
}

//SetQlcService set netService
func (sm *StreamManager) SetQlcNodeAndMaxStreamNum(node *QlcNode) {
	sm.node = node
	if node.localNetAttribute == PublicNet {
		sm.maxStreamNum = MaxStreamNumForPublicNet
	}
	if node.localNetAttribute == Intranet {
		sm.maxStreamNum = MaxStreamNumForIntranet
	}
}

// Add a new stream into the stream manager
func (sm *StreamManager) Add(s network.Stream) {
	stream := NewStream(s, sm.node)
	sm.AddStream(stream)
}

// AddStream into the stream manager
func (sm *StreamManager) AddStream(stream *Stream) {

	if sm.activePeersCount >= sm.maxStreamNum {
		if stream.stream != nil {
			_ = stream.stream.Close()
		}
		return
	}
	// check & close old stream
	if v, ok := sm.allStreams.Load(stream.pid.Pretty()); ok {
		old, _ := v.(*Stream)

		sm.node.logger.Info("Removing old stream.")

		sm.allStreams.Delete(old.pid.Pretty())
		sm.activePeersCount--
		if old.stream != nil {
			if err := old.stream.Close(); err != nil {
				sm.node.logger.Error("stream close error")
			}
		}
	}

	sm.node.logger.Infof("Added a new stream:[%s]", stream.pid.Pretty())
	sm.activePeersCount++
	sm.allStreams.Store(stream.pid.Pretty(), stream)
	stream.StartLoop()
}

// RemoveStream from the stream manager
func (sm *StreamManager) RemoveStream(s *Stream) {
	if v, ok := sm.allStreams.Load(s.pid.Pretty()); ok {
		exist, _ := v.(*Stream)
		if s != exist {
			return
		}
		sm.node.logger.Debugf("Removing a stream:[%s]", s.pid.Pretty())
		sm.allStreams.Delete(s.pid.Pretty())
		sm.activePeersCount--
	}
}

// FindByPeerID find the stream with the given peerID
func (sm *StreamManager) FindByPeerID(peerID string) *Stream {
	if v, ok := sm.allStreams.Load(peerID); ok {
		return v.(*Stream)
	}
	return nil
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

// CloseStream with the given pid and reason
func (sm *StreamManager) CloseStream(peerID string) error {
	stream := sm.FindByPeerID(peerID)
	if stream != nil {
		err := stream.close()
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateStreamWithPeer create stream with a peer.
func (sm *StreamManager) createStreamWithPeer(pid peer.ID) {

	stream := sm.FindByPeerID(pid.Pretty())

	if stream == nil {
		stream = NewStreamFromPID(pid, sm.node)
		sm.AddStream(stream)
	}
}

// BroadcastMessage broadcast the message
func (sm *StreamManager) BroadcastMessage(messageName MessageType, v interface{}) {
	messageContent, err := marshalMessage(messageName, v)
	if err != nil {
		sm.node.logger.Error(err)
		return
	}
	version := p2pVersion
	message := NewQlcMessage(messageContent, byte(version), messageName)
	qlcMessage := &QlcMessage{
		messageType: messageName,
		content:     message,
	}
	hash, err := types.HashBytes(message)
	if err != nil {
		sm.node.logger.Error(err)
		return
	}

	msgNeedCache := false
	if messageName == PublishReq || messageName == ConfirmReq || messageName == ConfirmAck ||
		messageName == PovPublishReq {
		msgNeedCache = true
	}

	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if msgNeedCache {
			if sm.hasMsgInCache(stream, hash) {
				return true
			}
		}
		_ = stream.SendMessageToChan(qlcMessage)
		if msgNeedCache {
			sm.searchCache(stream, hash, message, messageName)
		}
		return true
	})
}

func (sm *StreamManager) SendMessageToPeers(messageName MessageType, v interface{}, peerID string) {
	messageContent, err := marshalMessage(messageName, v)
	if err != nil {
		sm.node.logger.Error(err)
		return
	}
	version := p2pVersion
	message := NewQlcMessage(messageContent, byte(version), messageName)
	qlcMessage := &QlcMessage{
		messageType: messageName,
		content:     message,
	}
	hash, err := types.HashBytes(message)
	if err != nil {
		sm.node.logger.Error(err)
		return
	}
	msgNeedCache := false
	if messageName == PublishReq || messageName == ConfirmReq || messageName == ConfirmAck ||
		messageName == PovPublishReq {
		msgNeedCache = true
	}

	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if stream.pid.Pretty() != peerID {
			if msgNeedCache {
				if sm.hasMsgInCache(stream, hash) {
					return true
				}
			}
			_ = stream.SendMessageToChan(qlcMessage)
			if msgNeedCache {
				sm.searchCache(stream, hash, message, messageName)
			}
		}
		return true
	})
}

func (sm *StreamManager) searchCache(stream *Stream, hash types.Hash, message []byte, messageName MessageType) {
	var cs []*cacheValue
	var c *cacheValue
	if sm.node.netService.msgService.cache.Has(hash) {
		exitCache, e := sm.node.netService.msgService.cache.Get(hash)
		if e != nil {
			return
		}
		cs = exitCache.([]*cacheValue)
		for k, v := range cs {
			if v.peerID == stream.pid.Pretty() {
				v.resendTimes++
				break
			}
			if k == (len(cs) - 1) {
				c = &cacheValue{
					peerID:      stream.pid.Pretty(),
					resendTimes: 0,
					startTime:   time.Now(),
					data:        message,
					t:           messageName,
				}
				cs = append(cs, c)
				err := sm.node.netService.msgService.cache.Set(hash, cs)
				if err != nil {
					sm.node.logger.Error(err)
				}
			}
		}
	} else {
		c = &cacheValue{
			peerID:      stream.pid.Pretty(),
			resendTimes: 0,
			startTime:   time.Now(),
			data:        message,
			t:           messageName,
		}
		cs = append(cs, c)
		err := sm.node.netService.msgService.cache.Set(hash, cs)
		if err != nil {
			sm.node.logger.Error(err)
		}
	}
}

func (sm *StreamManager) hasMsgInCache(stream *Stream, hash types.Hash) bool {
	exitCache, e := sm.node.netService.msgService.cache.Get(hash)
	if e == nil && exitCache != nil {
		cs := exitCache.([]*cacheValue)
		for _, v := range cs {
			if v.peerID == stream.pid.Pretty() {
				return true
			}
		}
	}

	return false
}

func (sm *StreamManager) PeerCounts() int {
	allPeers := make(PeersSlice, 0)

	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if stream.IsConnected() {
			allPeers = append(allPeers, value)
		}
		return true
	})

	return len(allPeers)
}

func (sm *StreamManager) GetAllConnectPeersInfo(p map[string]string) {
	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if stream.IsConnected() {
			p[value.(*Stream).pid.Pretty()] = value.(*Stream).addr.String()
		}
		return true
	})
}

func (sm *StreamManager) IsConnectWithPeerId(peerID string) bool {
	s := sm.FindByPeerID(peerID)
	if s == nil {
		return false
	}
	return s.IsConnected()
}
