package protos

import (
	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

const (
	PovReasonFetch = iota
	PovReasonSync
)

const (
	PovDirForward = iota
	PovDirBackward
)

type PovBulkPullReq struct {
	StartHash   types.Hash
	StartHeight uint64
	Count       uint32
	Direction   uint32
	Reason      uint32
}

type PovBulkPullRsp struct {
	Count  uint32
	Reason uint32
	Blocks types.PovBlocks
}

func PovBulkPullReqToProto(req *PovBulkPullReq) ([]byte, error) {
	pbReq := &pb.PovPullBlockReq{
		StartHash:   req.StartHash[:],
		StartHeight: req.StartHeight,
		Count:       req.Count,
		Direction:   req.Direction,
		Reason:      req.Reason,
	}
	data, err := proto.Marshal(pbReq)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func PovBulkPullReqFromProto(data []byte) (*PovBulkPullReq, error) {
	pbReq := new(pb.PovPullBlockReq)
	if err := proto.Unmarshal(data, pbReq); err != nil {
		return nil, err
	}

	req := &PovBulkPullReq{
		StartHeight: pbReq.StartHeight,
		Count:       pbReq.Count,
		Direction:   pbReq.Direction,
		Reason:      pbReq.Reason,
	}

	err := req.StartHash.UnmarshalBinary(pbReq.StartHash)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func PovBulkPullRspToProto(rsp *PovBulkPullRsp) ([]byte, error) {
	blockBytes, err := rsp.Blocks.Serialize()
	if err != nil {
		return nil, err
	}

	pbReq := &pb.PovPullBlockRsp{
		Blocktype: 0,
		Count:     rsp.Count,
		Block:     blockBytes,
		Reason:    rsp.Reason,
	}

	data, err := proto.Marshal(pbReq)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func PovBulkPullRspFromProto(data []byte) (*PovBulkPullRsp, error) {
	pbRsp := new(pb.PovPullBlockRsp)
	if err := proto.Unmarshal(data, pbRsp); err != nil {
		return nil, err
	}

	blocks := make(types.PovBlocks, pbRsp.Count)
	err := blocks.Deserialize(pbRsp.Block)
	if err != nil {
		return nil, err
	}

	rsp := &PovBulkPullRsp{
		Count:  pbRsp.Count,
		Blocks: blocks,
		Reason: pbRsp.Reason,
	}
	return rsp, nil
}
