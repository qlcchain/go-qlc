package pov

import (
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	peerStatusInit = 0
	peerStatusGood = 1
	peerStatusDead = 2
	peerStatusSlow = 3
)

type PovSyncPeer struct {
	peerID         string
	currentHeight  uint64
	currentTD      *big.Int
	timestamp      int64
	lastStatusTime time.Time
	status         int

	waitSyncRspMsg  bool
	waitLocatorRsp  bool
	lastSyncReqTime time.Time
}

// PeerSetByTD is in descend order
type PovSyncPeerSetByTD []*PovSyncPeer

func (s PovSyncPeerSetByTD) Len() int           { return len(s) }
func (s PovSyncPeerSetByTD) Less(i, j int) bool { return s[i].currentTD.Cmp(s[j].currentTD) > 0 }
func (s PovSyncPeerSetByTD) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type PovSyncPeerSetByHeight []*PovSyncPeer

func (s PovSyncPeerSetByHeight) Len() int           { return len(s) }
func (s PovSyncPeerSetByHeight) Less(i, j int) bool { return s[i].currentHeight < s[j].currentHeight }
func (s PovSyncPeerSetByHeight) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (ss *PovSyncer) onAddP2PStream(peerID string) {
	ss.logger.Debugf("add peer %s", peerID)

	peer := &PovSyncPeer{
		peerID:         peerID,
		currentHeight:  0,
		currentTD:      big.NewInt(0),
		lastStatusTime: time.Now(),
		status:         peerStatusInit,
	}

	ss.allPeers.Store(peerID, peer)

	ss.eventCh <- &PovSyncEvent{eventType: common.EventAddP2PStream, eventData: peerID}
}

func (ss *PovSyncer) onDeleteP2PStream(peerID string) {
	ss.logger.Debugf("delete peer %s", peerID)
	ss.allPeers.Delete(peerID)

	ss.eventCh <- &PovSyncEvent{eventType: common.EventDeleteP2PStream, eventData: peerID}
}

func (ss *PovSyncer) onPovStatus(status *protos.PovStatus, msgHash types.Hash, msgPeer string) {
	if v, ok := ss.allPeers.Load(msgPeer); ok {
		peer := v.(*PovSyncPeer)

		td := new(big.Int).SetBytes(status.CurrentTD)
		ss.logger.Infof("recv PovStatus from peer %s, head %d/%s, td %d/%s",
			msgPeer, status.CurrentHeight, status.CurrentHash, td.BitLen(), td.Text(16))
		if status.GenesisHash != ss.chain.GenesisBlock().GetHash() {
			ss.logger.Warnf("peer %s genesis hash %s is invalid", msgPeer, status.GenesisHash)
			return
		}

		peer.currentHeight = status.CurrentHeight
		peer.currentTD = td
		peer.timestamp = status.Timestamp
		peer.lastStatusTime = time.Now()
		if peer.status != peerStatusSlow {
			peer.status = peerStatusGood
		}
	}
}

func (ss *PovSyncer) processStreamEvent(event *PovSyncEvent) {
	peerID := event.eventData.(string)

	genesisBlock := ss.chain.GenesisBlock()
	latestBlock := ss.chain.LatestBlock()
	latestTD := ss.chain.GetBlockTDByHash(latestBlock.GetHash())

	status := &protos.PovStatus{
		CurrentHeight: latestBlock.GetHeight(),
		CurrentTD:     latestTD.Bytes(),
		CurrentHash:   latestBlock.GetHash(),
		GenesisHash:   genesisBlock.GetHash(),
		Timestamp:     time.Now().Unix(),
	}
	ss.logger.Debugf("send PovStatus to peer %s", peerID)
	ss.eb.Publish(common.EventSendMsgToSingle, p2p.PovStatus, status, peerID)
}

func (ss *PovSyncer) checkAllPeers() {
	peerCount := ss.PeerCount()
	if peerCount <= 0 {
		return
	}

	genesisBlock := ss.chain.GenesisBlock()
	latestBlock := ss.chain.LatestBlock()
	latestTD := ss.chain.GetBlockTDByHash(latestBlock.GetHash())

	status := &protos.PovStatus{
		CurrentHeight: latestBlock.GetHeight(),
		CurrentTD:     latestTD.Bytes(),
		CurrentHash:   latestBlock.GetHash(),
		GenesisHash:   genesisBlock.GetHash(),
		Timestamp:     time.Now().Unix(),
	}
	ss.logger.Infof("broadcast PovStatus to %d peers", peerCount)
	ss.eb.Publish(common.EventBroadcast, p2p.PovStatus, status)

	now := time.Now()
	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if now.Sub(peer.lastStatusTime) >= 10*time.Minute {
			if peer.status != peerStatusDead {
				ss.logger.Infof("peer %s may be dead", peer.peerID)
				peer.status = peerStatusDead
			}
		}
		return true
	})
}

func (ss *PovSyncer) FindPeer(peerID string) *PovSyncPeer {
	if peerID == "" {
		return nil
	}

	if v, ok := ss.allPeers.Load(peerID); ok {
		peer := v.(*PovSyncPeer)
		return peer
	}

	return nil
}

func (ss *PovSyncer) FindPeerWithStatus(peerID string, status int) *PovSyncPeer {
	peer := ss.FindPeer(peerID)
	if peer != nil {
		if peer.status == status {
			return peer
		}
	}

	return nil
}

func (ss *PovSyncer) GetBestPeer(lastPeerID string) *PovSyncPeer {
	bestPeer := ss.FindPeerWithStatus(lastPeerID, peerStatusGood)

	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if peer.status != peerStatusGood {
			return true
		}
		if bestPeer == nil {
			bestPeer = peer
		} else if peer.currentTD.Cmp(bestPeer.currentTD) > 0 {
			bestPeer = peer
		}
		return true
	})

	return bestPeer
}

func (ss *PovSyncer) GetBestPeers(limit int) []*PovSyncPeer {
	var allPeers PovSyncPeerSetByTD

	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if peer.status != peerStatusGood {
			return true
		}
		allPeers = append(allPeers, peer)
		return true
	})
	sort.Sort(allPeers)

	if len(allPeers) <= limit {
		return allPeers
	}

	return allPeers[:limit]
}

// GetRandomTopPeer select one peer from top peers
func (ss *PovSyncer) GetRandomTopPeer(top int) *PovSyncPeer {
	peers := ss.GetBestPeers(top)
	if len(peers) <= 0 {
		return nil
	}
	if len(peers) == 1 {
		return peers[0]
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := rd.Intn(len(peers))
	if idx >= len(peers) {
		idx = idx - 1
	}
	if idx < 0 {
		idx = 0
	}
	return peers[idx]
}

func (ss *PovSyncer) GetRandomPeers(limit int) []*PovSyncPeer {
	var allPeers []*PovSyncPeer
	var selectPeers []*PovSyncPeer

	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if peer.status != peerStatusGood {
			return true
		}
		allPeers = append(allPeers, peer)
		return true
	})

	if len(allPeers) <= limit {
		return allPeers
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	idxSeqs := rd.Perm(len(allPeers))

	for i := 0; i < limit; i++ {
		selectPeers = append(selectPeers, allPeers[idxSeqs[i]])
	}

	return selectPeers
}

func (ss *PovSyncer) GetPeerLocators() []*PovSyncPeer {
	var allPeers PovSyncPeerSetByTD

	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if peer.status != peerStatusGood {
			return true
		}
		allPeers = append(allPeers, peer)
		return true
	})
	sort.Sort(allPeers)

	if len(allPeers) <= 3 {
		return allPeers
	}

	var selectPeers []*PovSyncPeer
	selectPeers = append(selectPeers, allPeers[0])
	selectPeers = append(selectPeers, allPeers[len(allPeers)/2])
	selectPeers = append(selectPeers, allPeers[len(allPeers)-1])
	return selectPeers
}

func (ss *PovSyncer) PeerCount() int {
	peerCount := 0
	ss.allPeers.Range(func(key, value interface{}) bool {
		peerCount++
		return true
	})

	return peerCount
}

func (ss *PovSyncer) PeerCountWithStatus(status int) int {
	peerCount := 0
	ss.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*PovSyncPeer)
		if peer.status == status {
			peerCount++
		}
		return true
	})

	return peerCount
}
