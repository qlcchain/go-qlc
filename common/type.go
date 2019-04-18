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
	EventAddRelation    TopicType = "addRelation"
	EventDeleteRelation TopicType = "deleteRelation"
	EventRecvPovBlock   TopicType = "recvPovBlock"
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
)
