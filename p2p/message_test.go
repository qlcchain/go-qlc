package p2p

import (
	"bytes"
	"hash/crc32"
	"testing"
)

func TestQlcMessage(t *testing.T) {
	data := "testMessage"
	msgType := "test"
	version := byte(0x01)
	reserved := []byte{0x00, 0x00, 0x00, 0x00}
	content := NewQlcMessage([]byte(data), version, msgType)
	qlcMsg := &QlcMessage{
		content:     content,
		messageType: MessageType(msgType),
	}
	if bytes.Compare(qlcMsg.MagicNumber(), MagicNumber) != 0 {
		t.Fatal("Magic error")
	}
	if qlcMsg.Version() != version {
		t.Fatal("Version error")
	}
	if qlcMsg.MessageType() != MessageType(msgType) {
		t.Fatal("messageType error")
	}
	if qlcMsg.DataLength() != uint32(len(data)) {
		t.Fatal("DataLength error")
	}
	if bytes.Compare(qlcMsg.Reserved(), reserved) != 0 {
		t.Fatal("reserved error")
	}
	if qlcMsg.HeaderCheckSum() != crc32.ChecksumIEEE(qlcMsg.HeaderData()) {
		t.Fatal("HeaderCheckSum error")
	}
	if qlcMsg.DataCheckSum() != crc32.ChecksumIEEE([]byte(data)) {
		t.Fatal("DataCheckSum error")
	}
	if bytes.Compare(qlcMsg.MessageData(), []byte(data)) != 0 {
		t.Fatal("MessageData error")
	}
	msg, err := ParseQlcMessage(qlcMsg.content[:QlcMessageHeaderLength])
	if err != nil {
		t.Fatal("ParseQlcMessage error")
	}
	if bytes.Compare(msg.content, qlcMsg.content[:QlcMessageHeaderLength]) != 0 {
		t.Fatal("Compare ParseQlcMessage error")
	}
	qlcMsg.content = qlcMsg.content[:QlcMessageHeaderLength]
	err = qlcMsg.ParseMessageData([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	//test VerifyHeader()
	err = qlcMsg.VerifyHeader()
	if err != nil {
		t.Fatal(err)
	}
	h := []byte{0x00, 0x01, 0x02}
	copy(qlcMsg.content[0:QlcMessageMagicNumberEndIdx], h)
	err = qlcMsg.VerifyHeader()
	if err == nil {
		t.Fatal("VerifyHeader error")
	}

	//test VerifyData()
	err = qlcMsg.VerifyData()
	if err != nil {
		t.Fatal(err)
	}
	copy(qlcMsg.content[QlcMessageDataCheckSumEndIdx:QlcMessageDataCheckSumEndIdx+3], h)
	err = qlcMsg.VerifyData()
	if err == nil {
		t.Fatal("VerifyData error")
	}
}
