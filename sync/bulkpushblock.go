package sync

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/sync/pb"
)

type Bulk struct {
	StartHash types.Hash
	EndHash   types.Hash
}
type BulkPush struct {
	blk types.Block
}

// ToProto converts domain BulkPull into proto BulkPull
func BulkPushBlockToProto(bp *BulkPush) ([]byte, error) {
	blkdata, err := bp.blk.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	bppb := &pb.BulkPushBlock{
		Block: blkdata,
	}
	data, err := proto.Marshal(bppb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

/// BulkPullPacketFromProto parse the data into BulkPull message
func BulkPushBlockFromProto(data []byte) (*BulkPush, error) {
	bp := new(pb.BulkPullRsp)
	if err := proto.Unmarshal(data, bp); err != nil {
		fmt.Println("Failed to unmarshal BulkPullRspPacket message.")
		return nil, err
	}
	blockType := bp.Blocktype
	blk, err := types.NewBlock(byte(blockType))
	if err != nil {
		return nil, err
	}
	if _, err = blk.UnmarshalMsg(bp.Block); err != nil {
		return nil, err
	}
	pb := &BulkPush{
		blk: blk,
	}
	return pb, nil
}
