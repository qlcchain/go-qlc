package protos

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
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

	bppb := &pb.BulkPullReq{
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
	bp := new(pb.BulkPullReq)
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
	bprp := &BulkPullReqPacket{
		StartHash: start,
		EndHash:   end,
	}
	return bprp, nil
}

type BulkPullRspPacket struct {
	Blk types.Block
}

// ToProto converts domain BulkPull into proto BulkPull
func BulkPullRspPacketToProto(bp *BulkPullRspPacket) ([]byte, error) {
	blkdata, err := bp.Blk.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	blocktype := bp.Blk.GetType()
	bppb := &pb.BulkPullRsp{
		Blocktype: uint32(blocktype),
		Block:     blkdata,
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
	blk, err := types.NewBlock(types.BlockType(blockType))
	if err != nil {
		return nil, err
	}
	if _, err = blk.UnmarshalMsg(bp.Block); err != nil {
		return nil, err
	}
	bprp := &BulkPullRspPacket{
		Blk: blk,
	}
	return bprp, nil
}
