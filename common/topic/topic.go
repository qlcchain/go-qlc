/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package topic

type TopicType string

//Topic type
const (
	EventPublish           TopicType = "publish"
	EventConfirmReq        TopicType = "confirmReq"
	EventConfirmAck        TopicType = "confirmAck"
	EventSyncBlock         TopicType = "syncBlock"
	EventConfirmedBlock    TopicType = "confirmedBlock"
	EventBroadcast         TopicType = "broadcast"
	EventGetBandwidthStats TopicType = "getBandwidthStats"
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

	EventAddBlockCache TopicType = "addBlockCache"
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
	Syncing:      "Synchronizing",
	SyncDone:     "SyncDone",
	SyncFinish:   "SyncFinish",
}

func (s SyncState) String() string {
	if s > SyncFinish {
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
