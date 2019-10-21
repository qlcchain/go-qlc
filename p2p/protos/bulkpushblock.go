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
	Blocks types.StateBlockList
}

// ToProto converts domain BulkPush into proto BulkPush
func BulkPushBlockToProto(bp *BulkPush) ([]byte, error) {
	blockBytes, err := bp.Blocks.Serialize()
	if err != nil {
		return nil, err
	}
	bpPb := &pb.BulkPushBlock{}
	bpPb.Blocks = blockBytes
	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// BulkPushBlockFromProto parse the data into  BulkPush message
func BulkPushBlockFromProto(data []byte) (*BulkPush, error) {
	bp := new(pb.BulkPushBlock)
	if err := proto.Unmarshal(data, bp); err != nil {
		return nil, err
	}
	pushBlock := &BulkPush{}
	if len(bp.Blocks) > 0 {
		blocks := make(types.StateBlockList, 0)
		err := blocks.Deserialize(bp.Blocks)
		if err != nil {
			return nil, err
		}
		pushBlock.Blocks = blocks
	}
	return pushBlock, nil
}
