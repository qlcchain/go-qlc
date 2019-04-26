package common

type TopicType string

//Topic type
const (
	EventPublish         TopicType = "publish"
	EventConfirmReq      TopicType = "confirmReq"
	EventConfirmAck      TopicType = "confirmAck"
	EventSyncBlock       TopicType = "syncBlock"
	EventConfirmedBlock  TopicType = "confirmedBlock"
	EventBroadcast       TopicType = "broadcast"
	EventSendMsgToPeers  TopicType = "sendMsgToPeers"
	EventSendMsgToPeer   TopicType = "sendMsgToPeer"
	EventAddRelation     TopicType = "addRelation"
	EventDeleteRelation  TopicType = "deleteRelation"
	EventAddP2PStream    TopicType = "addP2PStream"
	EventDeleteP2PStream TopicType = "deleteP2PStream"
	EventPovPeerStatus   TopicType = "povPeerStatus"
	EventPovRecvBlock    TopicType = "povRecvBlock"
	EventPovBulkPullReq  TopicType = "povBulkPullReq"
	EventPovBulkPullRsp  TopicType = "povBulkPullRsp"
	EventPovSyncState    TopicType = "povSyncState"
)

//  Message Type
const (
	PublishReq      = "0" //PublishReq
	ConfirmReq      = "1" //ConfirmReq
	ConfirmAck      = "2" //ConfirmAck
	FrontierRequest = "3" //FrontierReq
	FrontierRsp     = "4" //FrontierRsp
	BulkPullRequest = "5" //BulkPullRequest
	BulkPullRsp     = "6" //BulkPullRsp
	BulkPushBlock   = "7" //BulkPushBlock
	MessageResponse = "8" //MessageResponse

	PovStatus      = "20"
	PovPublishReq  = "21"
	PovBulkPullReq = "22"
	PovBulkPullRsp = "23"
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
	SyncNotStart:   "Sync Not Start",
	Syncing:        "Synchronising",
	Syncdone:       "Sync done",
	Syncerr:        "Sync error",
}

func (s SyncState) String() string {
	if s > Syncerr {
		return "unknown sync state"
	}
	return syncStatus[s]
}