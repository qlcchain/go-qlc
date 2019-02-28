package p2p

import (
	"fmt"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/qlcchain/go-qlc/common"
)

// MessageType a string for message type.
type MessageType string

// Message interface for message.
type Message interface {
	MessageType() MessageType
	MessageFrom() string
	Data() []byte
	Hash() types.Hash
	Content() []byte
}

// PeersSlice is a slice which contains peers
type PeersSlice []interface{}

// Service net Service interface
type Service interface {
	common.Service
	Node() *QlcNode
	MessageEvent() *EventQueue
	Broadcast(messageName string, value interface{})
	SendMessageToPeer(messageName string, value interface{}, peerID string) error
	//Broadcast message, except for the peerID in the parameter
	SendMessageToPeers(messageName string, value interface{}, peerID string)
}

// Subscriber subscriber.
type Subscriber struct {
	// id usually the owner/creator, used for troubleshooting .
	id interface{}

	// msgChan chan for subscribed message.
	msgChan chan Message

	// msgType message type to subscribe
	msgType MessageType

	// doFilter dup message
	doFilter bool
}

// NewSubscriber return new Subscriber instance.
func NewSubscriber(id interface{}, msgChan chan Message, doFilter bool, msgType MessageType) *Subscriber {
	return &Subscriber{id, msgChan, msgType, doFilter}
}

// ID return id.
func (s *Subscriber) ID() interface{} {
	return s.id
}

// MessageType return msgTypes.
func (s *Subscriber) MessageType() MessageType {
	return s.msgType
}

// MessageChan return msgChan.
func (s *Subscriber) MessageChan() chan Message {
	return s.msgChan
}

// DoFilter return doFilter
func (s *Subscriber) DoFilter() bool {
	return s.doFilter
}

// BaseMessage base message
type BaseMessage struct {
	messageType MessageType
	from        string
	data        []byte //removed the header
	content     []byte //complete message data
}

// NewBaseMessage new base message
func NewBaseMessage(messageType MessageType, from string, data []byte, content []byte) Message {
	return &BaseMessage{messageType: messageType, from: from, data: data, content: content}
}

// MessageType get message type
func (msg *BaseMessage) MessageType() MessageType {
	return msg.messageType
}

// MessageFrom get message who send
func (msg *BaseMessage) MessageFrom() string {
	return msg.from
}

// Data get the message data
func (msg *BaseMessage) Data() []byte {
	return msg.data
}

// Content get the message content
func (msg *BaseMessage) Content() []byte {
	return msg.content
}

// Hash return the message hash
func (msg *BaseMessage) Hash() types.Hash {
	hash, _ := types.HashBytes(msg.content)
	return hash
}

// String get the message to string
func (msg *BaseMessage) String() string {
	return fmt.Sprintf("BaseMessage {type:%s; data:%s; from:%s}",
		msg.messageType,
		msg.data,
		msg.from,
	)
}
