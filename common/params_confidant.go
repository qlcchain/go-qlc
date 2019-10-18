// +build confidant

package common

var (
	//node type
	NodeType = nodeTypeConfidant

	//DPOS params
	DPoSMaxBlocks      = 1024
	DPoSMaxCacheBlocks = 1024

	//EventBus params
	EventBusWaitingQueueSize = 1024

	//Badger params
	BadgerMaxTableSize = int64(16 << 20)

	//P2P params
	P2PMsgChanSize         = 1024
	P2PMsgCacheSize        = 1024
	P2PMonitorMsgChanSize  = 1024
)
