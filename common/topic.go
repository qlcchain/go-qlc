package common

type TopicType string

//Topic type
const (
	EventPublish           TopicType = "publish"
	EventConfirmReq        TopicType = "confirmReq"
	EventConfirmAck        TopicType = "confirmAck"
	EventSyncBlock         TopicType = "syncBlock"
	EventConfirmedBlock    TopicType = "confirmedBlock"
	EventBroadcast         TopicType = "broadcast"
	EventSendMsgToPeers    TopicType = "sendMsgToPeers"
	EventPeersInfo         TopicType = "peersInfo"
	EventSyncing           TopicType = "syncing"
	EventAddRelation       TopicType = "addRelation"
	EventDeleteRelation    TopicType = "deleteRelation"
	EventGenerateBlock     TopicType = "generateBlock"
	EventRollbackUnchecked TopicType = "rollbackUnchecked"

	EventSendMsgToSingle TopicType = "sendMsgToSingle"
	EventAddP2PStream    TopicType = "addP2PStream"
	EventDeleteP2PStream TopicType = "deleteP2PStream"
	EventPovPeerStatus   TopicType = "povPeerStatus"
	EventPovRecvBlock    TopicType = "povRecvBlock"
	EventPovBulkPullReq  TopicType = "povBulkPullReq"
	EventPovBulkPullRsp  TopicType = "povBulkPullRsp"
	EventPovSyncState    TopicType = "povSyncState"

	EventPullBlocksReq     TopicType = "pullBlocksReq"
	EventFrontierConsensus TopicType = "frontierConsensus"
	EventFrontierConfirmed TopicType = "frontierConfirmed"
	EventSyncStateChange   TopicType = "syncStateChange"
)

// Sync state
type SyncState uint

const (
	SyncNotStart SyncState = iota
	Syncing
	Syncdone
)

var syncStatus = [...]string{
	SyncNotStart: "SyncNotStart",
	Syncing:      "Synchronising",
	Syncdone:     "Syncdone",
}

func (s SyncState) String() string {
	if s > Syncdone {
		return "unknown sync state"
	}
	return syncStatus[s]
}

func (s SyncState) IsSyncExited() bool {
	if s == Syncdone {
		return true
	}

	return false
}
