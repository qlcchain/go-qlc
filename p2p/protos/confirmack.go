package protos

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type ConfirmAckBlock struct {
	Account   types.Address
	Signature types.Signature
	Sequence  []byte
	Blk       types.Block
}

// ToProto converts domain ConfirmAckBlock into proto ConfirmAckBlock
func ConfirmAckBlockToProto(confirmack *ConfirmAckBlock) ([]byte, error) {
	blkdata, err := confirmack.Blk.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	bppb := &pb.ConfirmAck{
		Account:   confirmack.Account.Bytes(),
		Signature: confirmack.Signature[:],
		Sequence:  confirmack.Sequence[:],
		Blocktype: uint32(confirmack.Blk.GetType()),
		Block:     blkdata,
	}
	data, err := proto.Marshal(bppb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ConfirmAckBlockFromProto parse the data into ConfirmAckBlock message
func ConfirmAckBlockFromProto(data []byte) (*ConfirmAckBlock, error) {
	ca := new(pb.ConfirmAck)
	if err := proto.Unmarshal(data, ca); err != nil {
		fmt.Println("Failed to unmarshal BulkPullRspPacket message.")
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
