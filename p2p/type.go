package p2p

import (
	"encoding/hex"
	"fmt"
)

// MessageType a string for message type.
type MessageType string

// Message interface for message.
type Message interface {
	MessageType() string
	MessageFrom() string
	Data() []byte
	Hash() string
}

// Service net Service interface
type Service interface {
	Start() error
	Stop()

	Node() *QlcNode

	Register(...*Subscriber)
	Deregister(...*Subscriber)

	Broadcast(string, []byte)
}

// Subscriber subscriber.
type Subscriber struct {
	// id usually the owner/creator, used for troubleshooting .
	id interface{}

	// msgChan chan for subscribed message.
	msgChan chan Message

	// msgType message type to subscribe
	msgType string

	// doFilter dup message
	doFilter bool
}

// NewSubscriber return new Subscriber instance.
func NewSubscriber(id interface{}, msgChan chan Message, doFilter bool, msgType string) *Subscriber {
	return &Subscriber{id, msgChan, msgType, doFilter}
}

// ID return id.
func (s *Subscriber) ID() interface{} {
	return s.id
}

// MessageType return msgTypes.
func (s *Subscriber) MessageType() string {
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
	messageType string
	from        string
	data        []byte
}

// NewBaseMessage new base message
func NewBaseMessage(messageType string, from string, data []byte) Message {
	return &BaseMessage{messageType: messageType, from: from, data: data}
}

// MessageType get message type
func (msg *BaseMessage) MessageType() string {
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

// Hash return the message hash
func (msg *BaseMessage) Hash() string {
	return hex.EncodeToString(msg.data[:])
}

// String get the message to string
func (msg *BaseMessage) String() string {
	return fmt.Sprintf("BaseMessage {type:%s; data:%s; from:%s}",
		msg.messageType,
		msg.data,
		msg.from,
	)
}
