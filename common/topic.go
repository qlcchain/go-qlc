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
	EventGetBandwidthStats TopicType = "getBandwidthStats"
	EventSyncing           TopicType = "syncing"
	EventAddRelation       TopicType = "addRelation"
	EventAddSyncBlocks     TopicType = "addSyncBlocks"
	EventDeleteRelation    TopicType = "deleteRelation"
	EventGenerateBlock     TopicType = "generateBlock"
	EventRollback          TopicType = "rollback"
	EventRestartChain      TopicType = "restartChain"

	EventSendMsgToSingle TopicType = "sendMsgToSingle"
	EventAddP2PStream    TopicType = "addP2PStream"
	EventDeleteP2PStream TopicType = "deleteP2PStream"
	EventPovPeerStatus   TopicType = "povPeerStatus"
	EventPovRecvBlock    TopicType = "povRecvBlock"
	EventPovBulkPullReq  TopicType = "povBulkPullReq"
	EventPovBulkPullRsp  TopicType = "povBulkPullRsp"
	EventPovSyncState    TopicType = "povSyncState"

	EventPovConnectBestBlock    TopicType = "povConnectBestBlock"
	EventPovDisconnectBestBlock TopicType = "povDisconnectBestBlock"
	EventRpcSyncCall            TopicType = "rpcSyncCall"
	EventFrontiersReq           TopicType = "FrontiersReq"
	EventFrontierConsensus      TopicType = "frontierConsensus"
	EventFrontierConfirmed      TopicType = "frontierConfirmed"
	EventSyncStateChange        TopicType = "syncStateChange"
	EventConsensusSyncFinished  TopicType = "consensusSyncFinished"
	EventRepresentativeNode     TopicType = "representativeNode"
)

// Sync state
type SyncState uint

const (
	SyncNotStart SyncState = iota
	Syncing
	SyncDone
	SyncFinish
)

var syncStatus = [...]string{
	SyncNotStart: "SyncNotStart",
	Syncing:      "Synchronising",
	SyncDone:     "SyncDone",
}

func (s SyncState) String() string {
	if s > SyncDone {
		return "unknown sync state"
	}
	return syncStatus[s]
}

func (s SyncState) IsSyncExited() bool {
	if s == SyncDone {
		return true
	}

	return false
}
