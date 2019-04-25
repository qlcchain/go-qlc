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
)
