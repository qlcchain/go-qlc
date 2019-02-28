package protos

import (
	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type PublishBlock struct {
	Blk types.Block
}

// ToProto converts domain PublishBlock into proto PublishBlock
func PublishBlockToProto(publish *PublishBlock) ([]byte, error) {
	blkData, err := publish.Blk.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	blockType := publish.Blk.GetType()
	bpPb := &pb.PublishBlock{
		Blocktype: uint32(blockType),
		Block:     blkData,
	}
	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// PublishBlockFromProto parse the data into PublishBlock message
func PublishBlockFromProto(data []byte) (*PublishBlock, error) {
	bp := new(pb.PublishBlock)
	if err := proto.Unmarshal(data, bp); err != nil {
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
	bPush := &PublishBlock{
		Blk: blk,
	}
	return bPush, nil
}
