package protos

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type PublishBlock struct {
	Blk types.Block
}

// ToProto converts domain PublishBlock into proto PublishBlock
func PublishBlockToProto(publish *PublishBlock) ([]byte, error) {
	blkdata, err := publish.Blk.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	blocktype := publish.Blk.GetType()
	bppb := &pb.PublishBlock{
		Blocktype: uint32(blocktype),
		Block:     blkdata,
	}
	data, err := proto.Marshal(bppb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

/// PublishBlockFromProto parse the data into PublishBlock message
func PublishBlockFromProto(data []byte) (*PublishBlock, error) {
	bp := new(pb.PublishBlock)
	if err := proto.Unmarshal(data, bp); err != nil {
		fmt.Println("Failed to unmarshal BulkPullRspPacket message.")
		return nil, err
	}
	blockType := bp.Blocktype
	blk, err := types.NewBlock(types.BlockType(blockType))
	if err != nil {
		return nil, err
	}
	if _, err = blk.UnmarshalMsg(bp.Block); err != nil {
		return nil, err
	}
	bpush := &PublishBlock{
		Blk: blk,
	}
	return bpush, nil
}
