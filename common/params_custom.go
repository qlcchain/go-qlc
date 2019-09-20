// +build !confidant

package common

var (
	//node type
	NodeType = nodeTypeNormal

	//DPOS params
	DPoSMaxBlocks      = 102400
	DPoSMaxCacheBlocks = 102400
	DPoSVoteCacheEn    = true

	//EventBus params
	EventBusWaitingQueueSize = 102400

	//Consensus params
	ConsensusMsgCacheSize = 409600

	//Badger params
	BadgerMaxTableSize = int64(64 << 20)

	//P2P params
	P2PMsgChanSize         = 655350
	P2PMsgCacheSize        = 51200
	P2PMonitorMsgChanSize  = 65535
	P2PMonitorMsgCacheSize = 65535
)
