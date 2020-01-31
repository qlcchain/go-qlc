package p2p

import (
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"

	version2 "github.com/qlcchain/go-qlc/chain/version"
)

const (
	MaxStreamNum         = 50
	MaxPeersNumForRandom = 8
)

// StreamManager manages all streams
type StreamManager struct {
	mu               sync.Mutex
	allStreams       *sync.Map
	activePeersCount int32
	maxStreamNum     int32
	node             *QlcNode
	onlinePeersInfo  *sync.Map
}

// NewStreamManager return a new stream manager
func NewStreamManager() *StreamManager {
	return &StreamManager{
		allStreams:       new(sync.Map),
		activePeersCount: 0,
		onlinePeersInfo:  new(sync.Map),
	}
}

//SetQlcService set netService
func (sm *StreamManager) SetQlcNodeAndMaxStreamNum(node *QlcNode) {
	sm.node = node
	sm.maxStreamNum = MaxStreamNum
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

func (sm *StreamManager) AddOrUpdateOnlineInfo(info *types.PeerInfo) {
	// check & close old stream
	if _, ok := sm.onlinePeersInfo.Load(info.PeerID); ok {
		sm.onlinePeersInfo.Delete(info.PeerID)
	}
	sm.onlinePeersInfo.Store(info.PeerID, info)
}

func (sm *StreamManager) AddOnlineInfo(info *types.PeerInfo) {
	// check & close old stream
	if _, ok := sm.onlinePeersInfo.Load(info.PeerID); ok {
		return
	}
	sm.onlinePeersInfo.Store(info.PeerID, info)
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

func (sm *StreamManager) lowestLatencyPeer() (string, error) {
	var allPeers []*Stream
	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if stream.IsConnected() {
			allPeers = append(allPeers, stream)
		}
		return true
	})
	if (len(allPeers)) == 0 {
		return "", ErrNoStream
	}
	sort.Slice(allPeers, func(i, j int) bool { return allPeers[i].rtt < allPeers[j].rtt })
	return allPeers[0].pid.Pretty(), nil
}

func (sm *StreamManager) randomLowerLatencyPeer() (string, error) {
	var allPeers []*Stream
	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if stream.IsConnected() {
			allPeers = append(allPeers, stream)
		}
		return true
	})

	if (len(allPeers)) == 0 {
		return "", ErrNoStream
	}
	var ps []*Stream
	if version2.Version == "" {
		for i := 0; i < len(allPeers); i++ {
			if allPeers[i].p2pVersion >= p2pVersion {
				ps = append(ps, allPeers[i])
			}
		}
	} else {
		vs := regexp.MustCompile("[0-9]").FindAllString(version2.Version, 3)
		c := strings.Join(vs, "")
		v, err := strconv.Atoi(c)
		if err != nil {
			return "", ErrNoStream
		}
		for i := 0; i < len(allPeers); i++ {
			if allPeers[i].globalVersion != "" {
				a := strings.Split(allPeers[i].globalVersion, "-")
				vsp := regexp.MustCompile("[0-9]").FindAllString(a[0], 3)
				c := strings.Join(vsp, "")
				vp, err := strconv.Atoi(c)
				if err != nil {
					continue
				}
				if vp >= v {
					ps = append(ps, allPeers[i])
				}
			}
		}
	}
	if (len(ps)) == 0 {
		ps = allPeers
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].rtt < ps[j].rtt })
	var peerID string
	rand.Seed(time.Now().Unix())
	var temp []*Stream
	if len(ps) >= MaxPeersNumForRandom {
		temp = ps[:MaxPeersNumForRandom]
	} else {
		temp = ps
	}
	randNum := rand.Intn(len(temp))
	for i, v := range temp {
		if i == randNum {
			peerID = v.pid.Pretty()
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
	err = sm.node.publisher.Publish(MsgTopic, message)
	if err != nil {
		sm.node.logger.Error(err)
		return
	}
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

func (sm *StreamManager) GetAllConnectPeersInfo(pr *[]*types.PeerInfo) {
	var p []*types.PeerInfo
	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if stream.IsConnected() {
			ps := &types.PeerInfo{
				PeerID:         stream.pid.Pretty(),
				Address:        stream.addr.String(),
				Version:        stream.globalVersion,
				Rtt:            stream.rtt.Seconds(),
				LastUpdateTime: stream.lastUpdateTime,
			}
			p = append(p, ps)
		}
		return true
	})
	*pr = p
}

func (sm *StreamManager) GetOnlinePeersInfo(pr *[]*types.PeerInfo) {
	var p []*types.PeerInfo
	sm.onlinePeersInfo.Range(func(key, value interface{}) bool {
		ps := value.(*types.PeerInfo)
		p = append(p, ps)
		return true
	})
	*pr = p
}

func (sm *StreamManager) IsConnectWithPeerId(peerID string) bool {
	s := sm.FindByPeerID(peerID)
	if s == nil {
		return false
	}
	return s.IsConnected()
}
