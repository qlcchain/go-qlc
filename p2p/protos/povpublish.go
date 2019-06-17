package protos

import (
	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type PovPublishBlock struct {
	Blk *types.PovBlock
}

func PovPublishBlockToProto(publish *PovPublishBlock) ([]byte, error) {
	blkData, err := publish.Blk.Serialize()
	if err != nil {
		return nil, err
	}
	bpPb := &pb.PovPublishBlock{
		Blocktype: 0,
		Block:     blkData,
	}
	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func PovPublishBlockFromProto(data []byte) (*PovPublishBlock, error) {
	bp := new(pb.PovPublishBlock)
	if err := proto.Unmarshal(data, bp); err != nil {
		return nil, err
	}
	blk := new(types.PovBlock)
	if err := blk.Deserialize(bp.Block); err != nil {
		return nil, err
	}
	bPush := &PovPublishBlock{
		Blk: blk,
	}
	return bPush, nil
}
