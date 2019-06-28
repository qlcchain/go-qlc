// +build confidant

package common

var (
	//node type
	NodeType = nodeTypeConfidant

	//DPOS params
	DPoSMaxBlocks      = 1024
	DPoSMaxCacheBlocks = 1024
	DPoSVoteCacheEn    = false

	//EventBus params
	EventBusWaitingQueueSize = 1024

	//Consensus params
	ConsensusMsgCacheSize = 1024

	//Badger params
	BadgerMaxTableSize = int64(16 << 20)

	//P2P params
	P2PMsgChanSize         = 1024
	P2PCacheSize           = 1024
	P2PMonitorMsgChanSize  = 1024
	P2PMonitorMsgCacheSize = 1024
)
