package protos

import (
	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type MessageAckPacket struct {
	MessageHash types.Hash
}

func MessageAckToProto(messageAck *MessageAckPacket) ([]byte, error) {
	bpPb := &pb.MessageAck{
		MessageHash: messageAck.MessageHash.Bytes(),
	}

	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func MessageAckFromProto(data []byte) (*MessageAckPacket, error) {
	ma := new(pb.MessageAck)
	if err := proto.Unmarshal(data, ma); err != nil {
		return nil, err
	}
	h, err := types.BytesToHash(ma.MessageHash)
	if err != nil {
		return nil, err
	}
	ack := &MessageAckPacket{
		MessageHash: h,
	}
	return ack, nil
}
