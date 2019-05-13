package common

type TopicType string

//Topic type
const (
	EventPublish        TopicType = "publish"
	EventConfirmReq     TopicType = "confirmReq"
	EventConfirmAck     TopicType = "confirmAck"
	EventSyncBlock      TopicType = "syncBlock"
	EventConfirmedBlock TopicType = "confirmedBlock"
	EventBroadcast      TopicType = "broadcast"
	EventSendMsgToPeers TopicType = "sendMsgToPeers"
	EventPeersInfo      TopicType = "peersInfo"
	EventAddRelation    TopicType = "addRelation"
	EventDeleteRelation TopicType = "deleteRelation"

	EventSendMsgToPeer   TopicType = "sendMsgToPeer"
	EventAddP2PStream    TopicType = "addP2PStream"
	EventDeleteP2PStream TopicType = "deleteP2PStream"
	EventPovPeerStatus   TopicType = "povPeerStatus"
	EventPovRecvBlock    TopicType = "povRecvBlock"
	EventPovBulkPullReq  TopicType = "povBulkPullReq"
	EventPovBulkPullRsp  TopicType = "povBulkPullRsp"
	EventPovSyncState    TopicType = "povSyncState"
)

// Sync state
type SyncState uint

const (
	SyncNotStart SyncState = iota
	Syncing
	Syncdone
	Syncerr
)

var syncStatus = [...]string{
	SyncNotStart: "Sync Not Start",
	Syncing:      "Synchronising",
	Syncdone:     "Sync done",
	Syncerr:      "Sync error",
}

func (s SyncState) String() string {
	if s > Syncerr {
		return "unknown sync state"
	}
	return syncStatus[s]
}
