package protos

import (
	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type ConfirmAckBlock struct {
	Account   types.Address
	Signature types.Signature
	Sequence  uint32
	Blk       types.Block
}

// ToProto converts domain ConfirmAckBlock into proto ConfirmAckBlock
func ConfirmAckBlockToProto(confirmAck *ConfirmAckBlock) ([]byte, error) {
	blkData, err := confirmAck.Blk.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	bpPb := &pb.ConfirmAck{
		Account:   confirmAck.Account.Bytes(),
		Signature: confirmAck.Signature[:],
		Sequence:  confirmAck.Sequence,
		Blocktype: uint32(confirmAck.Blk.GetType()),
		Block:     blkData,
	}
	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ConfirmAckBlockFromProto parse the data into ConfirmAckBlock message
func ConfirmAckBlockFromProto(data []byte) (*ConfirmAckBlock, error) {
	ca := new(pb.ConfirmAck)
	if err := proto.Unmarshal(data, ca); err != nil {
		logger.Error("Failed to unmarshal BulkPullRspPacket message.")
		return nil, err
	}
	blockType := ca.Blocktype
	blk, err := types.NewBlock(types.BlockType(blockType))
	if err != nil {
		return nil, err
	}
	if _, err = blk.UnmarshalMsg(ca.Block); err != nil {
		return nil, err
	}
	account, err := types.BytesToAddress(ca.Account)
	if err != nil {
		return nil, err
	}
	var sign types.Signature
	err = sign.UnmarshalBinary(ca.Signature)
	if err != nil {
		return nil, err
	}
	ack := &ConfirmAckBlock{
		Account:   account,
		Signature: sign,
		Sequence:  ca.Sequence,
		Blk:       blk,
	}
	return ack, nil
}
