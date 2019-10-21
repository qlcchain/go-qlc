package protos

import (
	"github.com/gogo/protobuf/proto"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type PublishBlock struct {
	Blk *types.StateBlock
}

// ToProto converts domain PublishBlock into proto PublishBlock
func PublishBlockToProto(publish *PublishBlock) ([]byte, error) {
	blkData, err := publish.Blk.Serialize()
	if err != nil {
		return nil, err
	}
	bpPb := &pb.PublishBlock{
		Block: blkData,
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
	blk := new(types.StateBlock)
	if err := blk.Deserialize(bp.Block); err != nil {
		return nil, err
	}
	bPush := &PublishBlock{
		Blk: blk,
	}
	return bPush, nil
}
