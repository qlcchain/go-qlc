package sync

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/sync/pb"
)

type BulkPullReqPacket struct {
	StartHash types.Hash
	EndHash   types.Hash
}

func NewBulkPullReqPacket(start, end types.Hash) (packet *BulkPullReqPacket) {
	return &BulkPullReqPacket{
		StartHash: start,
		EndHash:   end,
	}
}

// ToProto converts domain BulkPull into proto BulkPull
func BulkPullReqPacketToProto(bp *BulkPullReqPacket) ([]byte, error) {

	bppb := &pb.BulkPullRep{
		StartHash: bp.StartHash[:],
		EndHash:   bp.EndHash[:],
	}
	data, err := proto.Marshal(bppb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

/// BulkPullPacketFromProto parse the data into BulkPull message
func BulkPullReqPacketFromProto(data []byte) (*BulkPullReqPacket, error) {
	bp := new(pb.BulkPullRep)
	var start, end types.Hash
	if err := proto.Unmarshal(data, bp); err != nil {
		fmt.Println("Failed to unmarshal BulkPullPacket message.")
		return nil, err
	}
	err := start.UnmarshalBinary(bp.StartHash)
	if err != nil {
		fmt.Println("StartHash error")
	}
	err = end.UnmarshalBinary(bp.EndHash)
	if err != nil {
		fmt.Println("EndHash error")
	}
	pb := &BulkPullReqPacket{
		StartHash: start,
		EndHash:   end,
	}
	return pb, nil
}

type BulkPullRspPacket struct {
	blk types.Block
}

// ToProto converts domain BulkPull into proto BulkPull
func BulkPullRspPacketToProto(bp *BulkPullRspPacket) ([]byte, error) {
	blkdata, err := bp.blk.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	bppb := &pb.BulkPullRsp{
		Block: blkdata,
	}
	data, err := proto.Marshal(bppb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

/// BulkPullPacketFromProto parse the data into BulkPull message
func BulkPullRspPacketFromProto(data []byte) (*BulkPullRspPacket, error) {
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
	pb := &BulkPullRspPacket{
		blk: blk,
	}
	return pb, nil
}
