package protos

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type ConfirmReqBlock struct {
	Blk types.Block
}

// ToProto converts domain ConfirmReqBlock into proto ConfirmReqBlock
func ConfirmReqBlockToProto(confirmreq *ConfirmReqBlock) ([]byte, error) {
	blkdata, err := confirmreq.Blk.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	blocktype := confirmreq.Blk.GetType()
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
func ConfirmReqBlockFromProto(data []byte) (*ConfirmReqBlock, error) {
	bp := new(pb.ConfirmReq)
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
	confirmreqblock := &ConfirmReqBlock{
		Blk: blk,
	}
	return confirmreqblock, nil
}
