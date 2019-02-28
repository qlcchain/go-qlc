package p2p

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
)

var (
	MagicNumber = []byte{0x51, 0x4C, 0x43} //QLC
)

const (
	QlcMessageHeaderLength         = 22
	QlcMessageMagicNumberEndIdx    = 3
	QlcMessageVersionEndIdx        = 4
	QlcMessageTypeEndIdx           = 6
	QlcMessageDataLengthEndIdx     = 10
	QlcMessageReservedEndIdx       = 14
	QlcMessageHeaderCheckSumEndIdx = 18
	QlcMessageDataCheckSumEndIdx   = 22
)

// Error types
var (
	ErrInvalidMessageHeaderLength = errors.New("invalid message header length")
	ErrInvalidMessageDataLength   = errors.New("invalid message data length")
	ErrInvalidMagicNumber         = errors.New("invalid magic number")
	ErrInvalidHeaderCheckSum      = errors.New("invalid header checksum")
	ErrInvalidDataCheckSum        = errors.New("invalid data checksum")
)

type QlcMessage struct {
	content     []byte
	messageType MessageType
}

// MagicNumber return magicNumber
func (message *QlcMessage) MagicNumber() []byte {
	return message.content[:QlcMessageMagicNumberEndIdx]
}

func (message *QlcMessage) Version() byte {

	return message.content[QlcMessageMagicNumberEndIdx]
}

func (message *QlcMessage) MessageType() MessageType {
	if message.messageType == "" {
		data := message.content[QlcMessageVersionEndIdx:QlcMessageTypeEndIdx]
		pos := bytes.IndexByte(data, 0)
		if pos != -1 {
			message.messageType = MessageType(data[0:pos])
		} else {
			message.messageType = MessageType(data)
		}
	}
	return message.messageType
}

func (message *QlcMessage) MessageData() []byte {
	return message.content[QlcMessageDataCheckSumEndIdx:]
}

// DataLength return dataLength
func (message *QlcMessage) DataLength() uint32 {
	return Uint32(message.content[QlcMessageTypeEndIdx:QlcMessageDataLengthEndIdx])
}

// Reserved return reserved
func (message *QlcMessage) Reserved() []byte {
	return message.content[QlcMessageDataLengthEndIdx:QlcMessageReservedEndIdx]
}

// HeaderCheckSum return header checkSum
func (message *QlcMessage) HeaderCheckSum() uint32 {
	return Uint32(message.content[QlcMessageReservedEndIdx:QlcMessageHeaderCheckSumEndIdx])
}

// DataCheckSum return data checkSum
func (message *QlcMessage) DataCheckSum() uint32 {
	return Uint32(message.content[QlcMessageHeaderCheckSumEndIdx:QlcMessageDataCheckSumEndIdx])
}

// HeaderData return HeaderData
func (message *QlcMessage) HeaderData() []byte {
	return message.content[:QlcMessageReservedEndIdx]
}

// NewQlcMessage new qlc message
func NewQlcMessage(data []byte, currentVersion byte, messageType string) []byte {
	message := &QlcMessage{
		content: make([]byte, QlcMessageHeaderLength+len(data)),
	}
	// copy header.
	copy(message.content[0:QlcMessageMagicNumberEndIdx], MagicNumber)
	message.content[QlcMessageMagicNumberEndIdx] = currentVersion
	copy(message.content[QlcMessageVersionEndIdx:QlcMessageTypeEndIdx], []byte(messageType))

	//copy datalength
	copy(message.content[QlcMessageTypeEndIdx:QlcMessageDataLengthEndIdx], FromUint32(uint32(len(data))))

	// header checksum.
	headerCheckSum := crc32.ChecksumIEEE(message.content[:QlcMessageReservedEndIdx])
	copy(message.content[QlcMessageReservedEndIdx:QlcMessageHeaderCheckSumEndIdx], FromUint32(uint32(headerCheckSum)))

	// copy data.
	copy(message.content[QlcMessageDataCheckSumEndIdx:], data)

	// data checksum.
	dataCheckSum := crc32.ChecksumIEEE(message.content[QlcMessageDataCheckSumEndIdx:])
	copy(message.content[QlcMessageHeaderCheckSumEndIdx:QlcMessageDataCheckSumEndIdx], FromUint32(uint32(dataCheckSum)))

	return message.content
}

// ParseqlcMessage parse qlc message
func ParseQlcMessage(data []byte) (*QlcMessage, error) {
	if len(data) < QlcMessageHeaderLength {
		return nil, ErrInvalidMessageHeaderLength
	}
	message := &QlcMessage{
		content: make([]byte, QlcMessageHeaderLength),
	}
	copy(message.content, data)
	if err := message.VerifyHeader(); err != nil {
		return nil, err
	}
	return message, nil
}

// ParseMessageData parse qlc message data
func (message *QlcMessage) ParseMessageData(data []byte) error {
	if uint32(len(data)) < message.DataLength() {
		return ErrInvalidMessageDataLength
	}
	message.content = append(message.content, data[:message.DataLength()]...)
	return message.VerifyData()
}

//VerifyHeader verify qlc message header
func (message *QlcMessage) VerifyHeader() error {
	if !Equal(MagicNumber, message.MagicNumber()) {
		return ErrInvalidMagicNumber
	}
	headerCheckSum := crc32.ChecksumIEEE(message.HeaderData())
	if headerCheckSum != message.HeaderCheckSum() {
		return ErrInvalidHeaderCheckSum
	}
	return nil
}

// VerifyData verify qlc message data
func (message *QlcMessage) VerifyData() error {
	dataCheckSum := crc32.ChecksumIEEE(message.MessageData())
	if dataCheckSum != message.DataCheckSum() {
		return ErrInvalidDataCheckSum
	}
	return nil
}

// FromUint32 decodes uint32.
func FromUint32(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

// Uint32 encodes []byte.
func Uint32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

// Equal checks whether byte slice a and b are equal.
func Equal(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
