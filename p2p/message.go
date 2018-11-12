package p2p

import (
	"bytes"
	"encoding/binary"
)

var (
	MagicNumber    = []byte{0x51, 0x4C, 0x43}
	CurrentVersion = byte(0x01)
)

const (
	QlcMessageHeaderLength      = 20
	QlcMessageMagicNumberEndIdx = 3
	QlcMessageVersionEndIdx     = 4
	QlcMessageTypeEndIdx        = 16
	QlcMessageDataLengthEndIdx  = 20
)

type QlcMessage struct {
	content     []byte
	messageType string
}

// MagicNumber return magicNumber
func (message *QlcMessage) MagicNumber() []byte {
	return message.content[0:QlcMessageMagicNumberEndIdx]
}
func (message *QlcMessage) Version() byte {

	return message.content[QlcMessageMagicNumberEndIdx]
}
func (message *QlcMessage) MessageType() string {
	if message.messageType == "" {
		data := message.content[QlcMessageVersionEndIdx:QlcMessageTypeEndIdx]
		pos := bytes.IndexByte(data, 0)
		if pos != -1 {
			message.messageType = string(data[0:pos])
		} else {
			message.messageType = string(data)
		}
	}
	return message.messageType
}
func (message *QlcMessage) MessageData() []byte {
	return message.content[QlcMessageDataLengthEndIdx:]
}

// DataLength return dataLength
func (message *QlcMessage) DataLength() uint32 {
	return binary.BigEndian.Uint32(message.content[QlcMessageTypeEndIdx:QlcMessageDataLengthEndIdx])
}

// NewQlcMessage new qlc message
func NewQlcMessage(data []byte, messagetype string) []byte {
	message := &QlcMessage{
		content: make([]byte, QlcMessageHeaderLength+len(data)),
	}
	// copy header.
	copy(message.content[0:QlcMessageMagicNumberEndIdx], MagicNumber)
	message.content[QlcMessageMagicNumberEndIdx] = CurrentVersion
	copy(message.content[QlcMessageVersionEndIdx:QlcMessageTypeEndIdx], []byte(messagetype))
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, (uint32)(len(data)))
	//copy length
	copy(message.content[QlcMessageTypeEndIdx:QlcMessageDataLengthEndIdx], b)

	// copy data.
	copy(message.content[QlcMessageDataLengthEndIdx:], data)

	return message.content
}

// ParseqlcMessage parse qlc message
func ParseQlcMessage(data []byte) (*QlcMessage, error) {
	message := &QlcMessage{
		content: make([]byte, QlcMessageHeaderLength),
	}
	copy(message.content, data)

	return message, nil
}
