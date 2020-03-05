package p2p

import (
	"bytes"
	"testing"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage(PublishReq, "QmYPq8Cqqfyhaj6pKCiCMVX3KFRMZwi4w6fU6wGLU2T9JC", []byte{0x03, 0x04}, []byte{0x01, 0x02, 0x03, 0x04})
	if msg.messageType != PublishReq {
		t.Fatal("messageType error")
	}
	if msg.MessageFrom() != "QmYPq8Cqqfyhaj6pKCiCMVX3KFRMZwi4w6fU6wGLU2T9JC" {
		t.Fatal("message from error")
	}
	if !bytes.Equal(msg.data, []byte{0x03, 0x04}) {
		t.Fatal("message data error")
	}
	if !bytes.Equal(msg.Content(), []byte{0x01, 0x02, 0x03, 0x04}) {
		t.Fatal("message content error")
	}
	t.Log(msg.String())
}
