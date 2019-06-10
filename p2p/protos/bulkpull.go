package protos

import (
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

	bpPb := &pb.BulkPullReq{
		StartHash: bp.StartHash[:],
		EndHash:   bp.EndHash[:],
	}
	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// BulkPullPacketFromProto parse the data into BulkPull message
func BulkPullReqPacketFromProto(data []byte) (*BulkPullReqPacket, error) {
	bp := new(pb.BulkPullReq)
	var start, end types.Hash
	if err := proto.Unmarshal(data, bp); err != nil {
		return nil, err
	}
	err := start.UnmarshalBinary(bp.StartHash)
	if err != nil {
		return nil, err
	}
	err = end.UnmarshalBinary(bp.EndHash)
	if err != nil {
		return nil, err
	}
	bpRp := &BulkPullReqPacket{
		StartHash: start,
		EndHash:   end,
	}
	return bpRp, nil
}

type BulkPullRspPacket struct {
	Blk *types.StateBlock
}

// ToProto converts domain BulkPull into proto BulkPull
func BulkPullRspPacketToProto(bp *BulkPullRspPacket) ([]byte, error) {
	blkData, err := bp.Blk.Serialize()
	if err != nil {
		return nil, err
	}
	blockType := bp.Blk.GetType()
	bpPb := &pb.BulkPullRsp{
		Blocktype: uint32(blockType),
		Block:     blkData,
	}
	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// BulkPullPacketFromProto parse the data into BulkPull message
func BulkPullRspPacketFromProto(data []byte) (*BulkPullRspPacket, error) {
	bp := new(pb.BulkPullRsp)
	if err := proto.Unmarshal(data, bp); err != nil {
		return nil, err
	}
	blk := new(types.StateBlock)
	if err := blk.Deserialize(bp.Block); err != nil {
		return nil, err
	}
	bpRp := &BulkPullRspPacket{
		Blk: blk,
	}
	return bpRp, nil
}
