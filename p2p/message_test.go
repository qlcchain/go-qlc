package p2p

import (
	"bytes"
	"testing"
)

func TestQlcMessage(t *testing.T) {
	data := "testmessage"
	msgtype := "test"
	content := NewQlcMessage([]byte(data), msgtype)
	qlcMsg := &QlcMessage{
		content:     content,
		messageType: MessageType(msgtype),
	}
	if bytes.Compare(qlcMsg.MagicNumber(), MagicNumber) != 0 {
		t.Fatal("Magic error")
	}
	if qlcMsg.Version() != CurrentVersion {
		t.Fatal("Version error")
	}
	if qlcMsg.MessageType() != MessageType(msgtype) {
		t.Fatal("messageType error")
	}
	if bytes.Compare(qlcMsg.MessageData(), []byte(data)) != 0 {
		t.Fatal("MessageData error")
	}
	if qlcMsg.DataLength() != uint32(len(data)) {
		t.Fatal("DataLength error")
	}
	msg, _ := ParseQlcMessage(qlcMsg.content[:QlcMessageHeaderLength])
	if bytes.Compare(msg.content, qlcMsg.content[:QlcMessageHeaderLength]) != 0 {
		t.Fatal("ParseQlcMessage error")
	}
}
