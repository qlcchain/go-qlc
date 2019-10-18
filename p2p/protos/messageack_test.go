package protos

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestMessageAck(t *testing.T) {
	var h types.Hash
	_ = h.Of("12F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47")
	ma := &MessageAckPacket{
		MessageHash: h,
	}
	b, err := MessageAckToProto(ma)
	if err != nil {
		t.Fatal(err)
	}
	mb, err := MessageAckFromProto(b)
	if err != nil {
		t.Fatal(err)
	}
	if ma.MessageHash != mb.MessageHash {
		t.Fatal("messageHash error")
	}
}
