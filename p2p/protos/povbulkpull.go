package protos

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

const (
	PovReasonFetch = iota
	PovReasonSync
)

const (
	PovPullTypeForward = iota
	PovPullTypeBackward
	PovPullTypeBatch
)

type PovBulkPullReq struct {
	StartHash   types.Hash
	StartHeight uint64
	Count       uint32
	PullType    uint32
	Reason      uint32
	Locators    []*types.Hash
}

type PovBulkPullRsp struct {
	Count  uint32
	Reason uint32
	Blocks types.PovBlocks
}

func PovBulkPullReqToProto(req *PovBulkPullReq) ([]byte, error) {
	locBytes := make([]byte, 0)
	for _, locItem := range req.Locators {
		locBytes = append(locBytes, locItem.Bytes()...)
	}
	pbReq := &pb.PovPullBlockReq{
		StartHash:   req.StartHash[:],
		StartHeight: req.StartHeight,
		Count:       req.Count,
		PullType:    req.PullType,
		Reason:      req.Reason,
		Locators:    locBytes,
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

	if len(pbReq.Locators)%types.HashSize != 0 {
		return nil, fmt.Errorf("invalid locators field length %d", len(pbReq.Locators))
	}
	locators := make([]*types.Hash, 0)
	for locIdx := 0; locIdx < len(pbReq.Locators); locIdx += types.HashSize {
		locHash, err := types.BytesToHash(pbReq.Locators[locIdx : locIdx+types.HashSize])
		if err != nil {
			return nil, err
		}
		locators = append(locators, &locHash)
	}

	req := &PovBulkPullReq{
		StartHeight: pbReq.StartHeight,
		Count:       pbReq.Count,
		PullType:    pbReq.PullType,
		Reason:      pbReq.Reason,
		Locators:    locators,
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
		Count:  rsp.Count,
		Block:  blockBytes,
		Reason: rsp.Reason,
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
