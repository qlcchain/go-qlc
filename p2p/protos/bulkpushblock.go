package protos

import (
	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type Bulk struct {
	StartHash types.Hash
	EndHash   types.Hash
}

type BulkPush struct {
	Blk types.Block
}

// ToProto converts domain BulkPush into proto BulkPush
func BulkPushBlockToProto(bp *BulkPush) ([]byte, error) {
	blkData, err := bp.Blk.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	blockType := bp.Blk.GetType()
	bpPb := &pb.BulkPushBlock{
		Blocktype: uint32(blockType),
		Block:     blkData,
	}
	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// BulkPushBlockFromProto parse the data into  BulkPush message
func BulkPushBlockFromProto(data []byte) (*BulkPush, error) {
	bp := new(pb.BulkPullRsp)
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
	bPush := &BulkPush{
		Blk: blk,
	}
	return bPush, nil
}
